package build

import (
	"github.com/McTalian/wow-build-tools/internal/github"
	"github.com/McTalian/wow-build-tools/internal/tokens"
)

func ghOutputVars(tokenMap tokens.SimpleTokenMap, zipFilePath string, isNoLib bool, noLibZipFilePath string) (err error) {
	defer func() {
		if err != nil {
			l.Error("Error setting GitHub Action output variables: %v", err)
		}
	}()

	err = github.Output(string(tokens.PackageName), tokenMap[tokens.PackageName])
	if err != nil {
		return
	}
	err = github.Output(string(tokens.ProjectVersion), tokenMap[tokens.ProjectVersion])
	if err != nil {
		return
	}
	err = github.Output("main-zip-path", zipFilePath)
	if err != nil {
		return
	}
	if isNoLib {
		err = github.Output("nolib-zip-path", noLibZipFilePath)
		if err != nil {
			return
		}
	}

	return
}
