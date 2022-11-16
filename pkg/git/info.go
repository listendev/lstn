package git

import (
	"github.com/go-git/go-git/v5"
)

type GitContext struct {
	Remotes []*git.Remote
}

func NewGitContextFrom(path string) (*GitContext, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	remotes, _ := repo.Remotes()
	c := &GitContext{
		Remotes: remotes,
	}

	return c, nil
}
