package build

import (
	"fmt"
	"maps"
	"path/filepath"
	"strconv"
	"time"

	"github.com/McTalian/wow-build-tools/internal/cmdargs"
	"github.com/McTalian/wow-build-tools/internal/config"
	"github.com/McTalian/wow-build-tools/internal/injector"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/McTalian/wow-build-tools/internal/tokens"
)

type BuildArgs struct {
	TopDir      string
	ReleaseDir  string
	PkgmetaFile string
	GameVersion string

	WatchMode bool

	CurseId string
	WagoId  string
	WowiId  string

	SkipChangelog    bool
	SkipCopy         bool
	SkipExternals    bool
	SkipUpload       bool
	SkipLocalization bool
	SkipZip          bool

	ForceExternals   bool
	OnlyLocalization bool

	CreateNoLib     bool
	KeepPackageDir  bool
	NameTemplate    string
	SplitToc        bool
	UnixLineEndings bool
}

var BuildParams = &BuildArgs{}

var l = logger.DefaultLogger

func configureLogger(args *BuildArgs) {
	if args.WatchMode {
		l.SetLogLevel(logger.WARN)
	} else if cmdargs.RootParams.LevelVerbose {
		l.SetLogLevel(logger.VERBOSE)
	} else if cmdargs.RootParams.LevelDebug {
		l.SetLogLevel(logger.DEBUG)
	} else {
		l.SetLogLevel(logger.INFO)
	}
}

// Build is the implementation of the build command.
func Build(args *BuildArgs) (err error) {
	start := time.Now()
	configureLogger(args)
	defer func() {
		if err != nil {
			l.Error("Build failed: %v", err)
		}
		l.Clear()
	}()

	err = toc.ParseGameVersionFlag(args.GameVersion)
	if err != nil {
		err = fmt.Errorf("error validating game version input argument: %v", err)
		return
	}

	var templateTokens *tokens.NameTemplate
	if args.NameTemplate == "help" {
		l.Info("%s", tokens.NameTemplateUsageInfo())
		return nil
	}
	templateTokens, err = tokens.NewNameTemplate(args.NameTemplate)
	if err != nil {
		err = fmt.Errorf("error parsing name template: %v", err)
		return
	}

	timeNow := time.Now()
	timeNowUtc := timeNow.UTC()
	buildTimestamp := timeNow.Unix()
	populateTokensArgs := populateTokensArgs{
		buildTimestampStr: strconv.FormatInt(buildTimestamp, 10),
		buildDate:         timeNowUtc.Format("2006-01-02"),
		buildDateIso:      timeNowUtc.Format("2006-01-02T15:04:05Z"),
		buildDateInteger:  timeNowUtc.Format("20060102150405"),
		buildYear:         timeNowUtc.Format("2006"),
	}

	topDir := args.TopDir

	if _, err = config.CreateExternalsCache(); err != nil {
		err = fmt.Errorf("cache error: %v", err)
		return
	}

	projectName, tocFiles, err := getTocFiles(topDir)
	if err != nil {
		err = fmt.Errorf("TOC error: %v", err)
		return
	}
	l.Info("%sBuilding %s...", logger.Build, projectName)

	preVr := time.Now()
	vR, err := prepareRepo(topDir)
	if err != nil {
		return
	}
	l.Timing("Creating VcsRepo took %s", time.Since(preVr))

	projectName, pkgMeta, err := parsePkgmeta(projectName, topDir, args.PkgmetaFile)
	if err != nil {
		return
	}

	packageDir, err := initializePkgDir(topDir, args, projectName, pkgMeta, vR)
	if err != nil {
		return
	}

	populateTokensArgs.projectName = projectName
	populateTokensArgs.vR = vR
	populateTokensArgs.packageDir = packageDir
	populateTokensArgs.unixLineEndings = args.UnixLineEndings
	tokensResult, err := populateTokens(populateTokensArgs)
	if err != nil {
		return
	}
	tokenMap := tokensResult.tokenMap
	flags := tokensResult.flags
	bTTM := tokensResult.bTTM
	releaseType := tokensResult.releaseType
	i, err := injector.NewInjector(tokenMap, vR, packageDir, bTTM, args.UnixLineEndings)
	if err != nil {
		return
	}

	err = i.Execute()
	if err != nil {
		return
	}

	var changelogTitle string
	if pkgMeta.ChangelogTitle != "" {
		changelogTitle = pkgMeta.ChangelogTitle
	} else {
		changelogTitle = projectName
	}

	cl, err := prepareChangeLog(args.SkipChangelog, vR, pkgMeta, changelogTitle, packageDir, topDir)
	if err != nil {
		return
	}
	defer cl.Cleanup()

	if !args.SkipExternals {
		err = pkgMeta.FetchExternals(packageDir, args.ForceExternals)
		if err != nil {
			return
		}
	}

	isNoLib := (args.CreateNoLib || pkgMeta.EnableNoLibCreation) && !args.WatchMode
	if isNoLib && !templateTokens.HasNoLib {
		l.Warn("Provided file and/or label template did not contain %s, but no-lib package requested. Skipping no-lib package since the zip name will not be unique.", tokens.NoLibFlag.NormalizeTemplateToken())
		isNoLib = false
	}

	zipFileName := templateTokens.GetFileName(&tokenMap, flags)
	zipFilePath := filepath.Join(args.ReleaseDir, zipFileName+".zip")
	noLibFlags := make(tokens.FlagMap)
	maps.Copy(noLibFlags, flags)
	noLibFlags[tokens.NoLibFlag] = "-nolib"
	noLibFileName := templateTokens.GetFileName(&tokenMap, noLibFlags)
	noLibZipFilePath := filepath.Join(args.ReleaseDir, noLibFileName+".zip")

	if !args.SkipZip {
		createZipsArgs := createZipsArgs{
			isNoLib:          isNoLib,
			zipFilePath:      zipFilePath,
			noLibZipFilePath: noLibZipFilePath,
			releaseDir:       args.ReleaseDir,
			packageDir:       packageDir,
			topDir:           topDir,
			unixLineEndings:  args.UnixLineEndings,
			pkgMeta:          pkgMeta,
			injector:         i,
		}
		err = createZips(createZipsArgs)
		if err != nil {
			return
		}
	}

	uploadToDistrosArgs := uploadToDistrosArgs{
		zipFilePath:    zipFilePath,
		fileLabel:      templateTokens.GetLabel(&tokenMap, flags),
		tocFiles:       tocFiles,
		pkgMeta:        pkgMeta,
		cl:             cl,
		releaseDir:     args.ReleaseDir,
		releaseType:    releaseType,
		projectName:    projectName,
		projectVersion: tokenMap[tokens.ProjectVersion],
		vR:             vR,
		isNoLib:        isNoLib,
		noLibFileName:  noLibFileName,
	}

	if !args.SkipUpload && !args.WatchMode {
		err = uploadToDistros(uploadToDistrosArgs)
		if err != nil {
			return
		}
	}

	l.TimingSummary()

	l.WarningsEncountered()

	err = ghOutputVars(tokenMap, zipFilePath, isNoLib, noLibZipFilePath)
	if err != nil {
		return
	}

	fmt.Println("")
	successMessage := fmt.Sprintf("%sSuccessfully packaged %s in %s%s", logger.Done, projectName, logger.Time, time.Since(start))
	if args.WatchMode {
		successMessage = fmt.Sprintf("%s at %s%s", successMessage, time.Now().Format("15:04:05"), logger.Watch)
	}
	l.Success("%s", successMessage)
	return nil
}
