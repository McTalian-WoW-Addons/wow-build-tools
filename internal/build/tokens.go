package build

import (
	"strings"
	"time"

	"github.com/McTalian/wow-build-tools/internal/repo"
	"github.com/McTalian/wow-build-tools/internal/toc"
	"github.com/McTalian/wow-build-tools/internal/tokens"
)

type populateTokensResult struct {
	tokenMap    tokens.SimpleTokenMap
	flags       tokens.FlagMap
	bTTM        tokens.BuildTypeTokenMap
	releaseType string
}

type populateTokensArgs struct {
	projectName       string
	buildTimestampStr string
	buildDate         string
	buildDateIso      string
	buildDateInteger  string
	buildYear         string
	forceAlpha        bool
	forceBeta         bool
	forceDev          bool
	vR                repo.VcsRepo
	packageDir        string
	unixLineEndings   bool
}

func populateTokens(args populateTokensArgs) (result populateTokensResult, err error) {
	defer func() {
		if err != nil {
			l.Error("Error populating tokens: %v", err)
		}
	}()

	tokenMap := tokens.SimpleTokenMap{
		tokens.PackageName:      args.projectName,
		tokens.BuildTimestamp:   args.buildTimestampStr,
		tokens.BuildDate:        args.buildDate,
		tokens.BuildDateIso:     args.buildDateIso,
		tokens.BuildDateInteger: args.buildDateInteger,
		tokens.BuildYear:        args.buildYear,
	}

	flags := tokens.FlagMap{
		tokens.NoLibFlag:   "",
		tokens.AlphaFlag:   "",
		tokens.BetaFlag:    "",
		tokens.ClassicFlag: "",
	}

	var releaseType string
	bTTM := tokens.BuildTypeTokenMap{
		tokens.Alpha:         false,
		tokens.Beta:          false,
		tokens.Classic:       false,
		tokens.Debug:         false,
		tokens.Retail:        false,
		tokens.VersionRetail: false,
		tokens.VersionBcc:    false,
		tokens.VersionWrath:  false,
		tokens.VersionCata:   false,
	}

	preGetInjectionValues := time.Now()
	if err = args.vR.GetInjectionValues(&tokenMap); err != nil {
		return
	}
	forcedBuildType := ""
	if args.forceAlpha {
		forcedBuildType = "alpha"
	} else if args.forceBeta {
		forcedBuildType = "beta"
	} else if args.forceDev {
		forcedBuildType = "dev"
	}

	if forcedBuildType != "" {
		l.Warn("Forced build mode active: %s (git/tag auto-detection disabled)", forcedBuildType)
		switch forcedBuildType {
		case "alpha":
			flags[tokens.AlphaFlag] = "-alpha"
			bTTM[tokens.Alpha] = true
			bTTM[tokens.Beta] = false
			bTTM[tokens.Debug] = false
			releaseType = "alpha"
		case "beta":
			flags[tokens.BetaFlag] = "-beta"
			bTTM[tokens.Alpha] = false
			bTTM[tokens.Beta] = true
			bTTM[tokens.Debug] = false
			releaseType = "beta"
		case "dev":
			flags[tokens.AlphaFlag] = "-alpha"
			bTTM[tokens.Alpha] = true
			bTTM[tokens.Beta] = false
			bTTM[tokens.Debug] = true
			releaseType = "alpha"
		}
	} else {

		tag := args.vR.GetCurrentTag()
		l.Verbose("Current Tag: %s", tag)
		if tag != "" {
			if strings.Contains(tag, "alpha") {
				flags[tokens.AlphaFlag] = "-alpha"
				bTTM[tokens.Alpha] = true
				bTTM[tokens.Beta] = false
				releaseType = "alpha"
			} else if strings.Contains(tag, "beta") {
				flags[tokens.BetaFlag] = "-beta"
				bTTM[tokens.Alpha] = false
				bTTM[tokens.Beta] = true
				releaseType = "beta"
			} else {
				bTTM[tokens.Alpha] = false
				bTTM[tokens.Beta] = false
				releaseType = "release"
			}
		} else {
			flags[tokens.AlphaFlag] = "-alpha"
			bTTM[tokens.Alpha] = true
			bTTM[tokens.Beta] = false
			releaseType = "alpha"
		}
	}
	l.Verbose("Release Type: %s", releaseType)
	flavors := toc.GetGameFlavors()
	if len(flavors) == 1 {
		switch flavors[0] {
		case toc.Retail:
			bTTM[tokens.Retail] = true
			bTTM[tokens.VersionRetail] = true
		case toc.ClassicEra:
			flags[tokens.ClassicFlag] = "-classic"
			bTTM[tokens.Classic] = true
		case toc.TbcClassic:
			bTTM[tokens.VersionBcc] = true
		case toc.WotlkClassic:
			bTTM[tokens.VersionWrath] = true
		case toc.CataClassic:
			bTTM[tokens.VersionCata] = true
		case toc.MistsClassic:
			bTTM[tokens.VersionMists] = true
		case toc.WodClassic:
			bTTM[tokens.VersionWod] = true
		case toc.LegionClassic:
			bTTM[tokens.VersionLegion] = true
		case toc.BfaClassic:
			bTTM[tokens.VersionBfa] = true
		case toc.SlClassic:
			bTTM[tokens.VersionSl] = true
		case toc.DfClassic:
			bTTM[tokens.VersionDf] = true
		default:
			bTTM[tokens.Retail] = true
		}
	}
	// TODO: Handle multiple game versions

	l.Verbose("%s", tokenMap.String())
	l.Timing("Getting Injection Values took %s", time.Since(preGetInjectionValues))

	result.tokenMap = tokenMap
	result.flags = flags
	result.bTTM = bTTM
	result.releaseType = releaseType

	return
}
