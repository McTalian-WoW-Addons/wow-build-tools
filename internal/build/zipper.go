package build

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/McTalian/wow-build-tools/internal/injector"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/McTalian/wow-build-tools/internal/pkg"
	"github.com/McTalian/wow-build-tools/internal/tokens"
)

type Zipper struct {
	pkgDir          string
	releaseDir      string
	topDir          string
	logGroup        *logger.LogGroup
	unixLineEndings bool
}

func (z *Zipper) Complete() {
	z.logGroup.Flush(true)
}

func (z *Zipper) ZipFiles(outputPath string, noLibArgs ...[]string) (err error) {
	defer func() {
		if err != nil {
			z.logGroup.Error("Error creating zip file: %v", err)
		}
	}()

	srcPath := z.pkgDir
	z.logGroup.Info("%sCreating %s", logger.ZipFile, outputPath)
	dirsToExclude := []string{}
	noLibStripPaths := []string{}
	if len(noLibArgs) > 0 {
		dirsToExclude = noLibArgs[0]
	}
	if len(noLibArgs) > 1 {
		noLibStripPaths = noLibArgs[1]
	}

	// Delete the destination file if it already exists
	if _, err = os.Stat(outputPath); err == nil {
		z.logGroup.Verbose("Removing existing file: %s", outputPath)
		err = os.Remove(outputPath)
		if err != nil {
			return
		}
	}

	// Create the zip file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return
	}
	defer func() { _ = zipFile.Close() }()

	// Initialize the zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	// Walk the source directory
	err = filepath.Walk(srcPath, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// Check if the directory should be excluded
		if info.IsDir() {
			for _, dir := range dirsToExclude {
				if path == dir {
					return filepath.SkipDir
				}
			}
		}

		// Create a header based on the file info
		header, e := zip.FileInfoHeader(info)
		if e != nil {
			return e
		}

		// Use a relative path so that the files are not stored with full system paths
		relPath, e := filepath.Rel(filepath.Dir(srcPath), path)
		if e != nil {
			return e
		}
		header.Name = relPath

		// If it's a directory, we need to end the header name with a "/"
		if info.IsDir() {
			header.Name += "/"
		} else {
			// Use deflate compression for files
			header.Method = zip.Deflate
		}

		// Create writer for this file/directory header
		writer, e := zipWriter.CreateHeader(header)
		if e != nil {
			return e
		}

		// For directories, no need to copy file content
		if info.IsDir() {
			return nil
		}

		// Check file size and warn if it seems too large
		if info.Size() > 1000000 {
			abbrevSize := float64(info.Size()) / 1000000.0
			trimmedPath := strings.ReplaceAll(path, z.pkgDir, z.topDir)
			trimmedDestPath := strings.TrimPrefix(outputPath, z.releaseDir+string(os.PathSeparator))
			z.logGroup.Warn("%s: %s is large (%f MB), consider adding it to ignores", trimmedDestPath, trimmedPath, abbrevSize)
		}

		if len(noLibStripPaths) > 0 {
			noLibStripVariants := tokens.NoLibStrip.GetVariants()
			for _, noLibStripPath := range noLibStripPaths {
				if path == noLibStripPath {
					// TODO, read the file and comment out the lib strip line
					contents, e := os.ReadFile(path)
					if e != nil {
						return e
					}

					contentsStr := string(contents)
					// Comment out the lib strip line
					var lineEnding string
					if z.unixLineEndings {
						lineEnding = "\n"
					} else {
						lineEnding = "\r\n"
					}
					contentsLines := strings.Split(contentsStr, lineEnding)
					var lineStart = -1
					var newContents []string
					for i, line := range contentsLines {
						if strings.Contains(line, fmt.Sprintf("@%s@", noLibStripVariants.Standard)) {
							lineStart = i
							continue
						}
						if strings.Contains(line, fmt.Sprintf("@%s@", noLibStripVariants.StandardEnd)) {
							lineStart = -1
							continue
						}
						if lineStart != -1 && i > lineStart {
							continue
						}
						newContents = append(newContents, line)
					}
					contentsStr = strings.Join(newContents, lineEnding)

					_, e = writer.Write([]byte(contentsStr))
					if e != nil {
						return e
					}
					return nil
				}
			}
		}

		// Open the file to be added
		file, e := os.Open(path)
		if e != nil {
			return e
		}
		defer func() { _ = file.Close() }()

		// Copy the file content into the zip writer
		_, e = io.Copy(writer, file)
		return e
	})

	return
}

func NewZipper(pkgDir string, releaseDir string, topDir string, unixLineEndings bool) *Zipper {
	logGroup := logger.NewLogGroup(fmt.Sprintf("%sCreating Zip File(s)", logger.Zip))
	return &Zipper{
		pkgDir:          pkgDir,
		releaseDir:      releaseDir,
		topDir:          topDir,
		logGroup:        logGroup,
		unixLineEndings: unixLineEndings,
	}
}

type createZipsArgs struct {
	isNoLib          bool
	zipFilePath      string
	noLibZipFilePath string
	releaseDir       string
	packageDir       string
	topDir           string
	unixLineEndings  bool
	pkgMeta          *pkg.PkgMeta
	injector         *injector.Injector
}

func createZips(args createZipsArgs) (err error) {
	defer func() {
		if err != nil {
			l.Error("Error creating zips: %v", err)
		}
	}()

	isNoLib := args.isNoLib
	zipFilePath := args.zipFilePath
	noLibZipFilePath := args.noLibZipFilePath
	releaseDir := args.releaseDir
	packageDir := args.packageDir
	topDir := args.topDir
	unixLineEndings := args.unixLineEndings
	pkgMeta := args.pkgMeta
	i := args.injector

	zipsToCreate := 1
	if isNoLib {
		zipsToCreate++
	}
	var zipWGroup sync.WaitGroup
	zipErrChan := make(chan error, zipsToCreate)

	z := NewZipper(packageDir, releaseDir, topDir, unixLineEndings)
	zipWGroup.Add(1)
	go func() {
		defer zipWGroup.Done()
		zipPath := zipFilePath
		zipErrChan <- z.ZipFiles(zipPath)
	}()

	if isNoLib {
		dirsToExclude := pkgMeta.GetNoLibDirs(packageDir)
		zipWGroup.Add(1)
		go func() {
			defer zipWGroup.Done()
			zipPath := noLibZipFilePath
			zipErrChan <- z.ZipFiles(zipPath, dirsToExclude, i.NoLibStripFiles)
		}()
	}

	zipWGroup.Wait()
	close(zipErrChan)
	z.Complete()

	// Collect errors
	errsEncountered := 0
	errStr := ""
	for e := range zipErrChan {
		if e != nil {
			errsEncountered++
			errStr += fmt.Sprintf("\n  - %s", e.Error())
		}
	}

	if errsEncountered > 0 {
		err = fmt.Errorf("encountered %d errors while creating zip files:%s", errsEncountered, errStr)
	}

	return
}
