package build

import (
	"fmt"

	"github.com/McTalian/wow-build-tools/internal/external"
	"github.com/McTalian/wow-build-tools/internal/repo"
)

func prepareRepo(topDir string) (vR repo.VcsRepo, err error) {
	defer func() {
		if err != nil {
			l.Error("Error preparing repo: %v", err)
		}
	}()

	r, err := repo.NewRepo(topDir)
	if err != nil {
		return
	}

	switch r.GetVcsType() {
	case external.Git:
		l.Verbose("Git repository detected")
		vR, err = repo.NewGitRepo(r)
		if err != nil {
			return
		}
	case external.Svn:
		l.Verbose("SVN repository detected")
	case external.Hg:
		l.Verbose("Mercurial repository detected")
	default:
		err = fmt.Errorf("unknown repository type: %s", r.GetVcsType().ToString())
	}
	return
}
