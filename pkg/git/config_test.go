// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	internaltesting "github.com/listendev/lstn/internal/testing"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigDoesntWorkForLocalConfig(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	err := stubGitConfig(fs, "/work/example", config.LocalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = local user
		`)
	})
	assert.Nil(t, err)

	conf, err := GetConfig(config.LocalScope)
	assert.Nil(t, err)

	assert.Equal(t, config.NewConfig(), conf)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)
	ccc, err := repo.Config()
	assert.Nil(t, err)
	assert.Equal(t, "local user", ccc.User.Name)
}

func TestGetConfigReturnsEmptyConfigWhenConfigFileNotFound(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	err := stubGitConfig(fs, "/work/example", config.SystemScope, func() string {
		return heredoc.Doc(`
		[user]
		name = system user
		`)
	})
	assert.Nil(t, err)

	conf, err := GetConfig(config.GlobalScope)
	assert.Nil(t, err)

	assert.Equal(t, config.NewConfig(), conf)
}

func TestGetConfigMalformedConfig(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	err := stubGitConfig(fs, "/work/example", config.SystemScope, func() string {
		return heredoc.Doc(`
		[user
		name = system user
		`)
	})
	assert.Nil(t, err)

	conf, err := GetConfig(config.SystemScope)
	assert.Error(t, err)
	assert.Nil(t, conf)
}

func TestGetFinalConfigFromSystemGitConfig(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	err := stubGitConfig(fs, "/work/example", config.SystemScope, func() string {
		return heredoc.Doc(`
		[user]
		name = system user
		`)
	})
	assert.Nil(t, err)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)

	conf, err := GetFinalConfig(repo)
	assert.Nil(t, err)

	assert.Equal(t, "system user", conf.User.Name)
}

func TestGetFinalConfigFromGlobalGitConfig(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	err := stubGitConfig(fs, "/work/example", config.GlobalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = global user
		`)
	})
	assert.Nil(t, err)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)

	conf, err := GetFinalConfig(repo)
	assert.Nil(t, err)

	assert.Equal(t, "global user", conf.User.Name)
}

func TestGetFinalConfigFromLocalGitConfig(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	err := stubGitConfig(fs, "/work/example", config.LocalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = local user
		`)
	})
	assert.Nil(t, err)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)

	conf, err := GetFinalConfig(repo)
	assert.Nil(t, err)

	assert.Equal(t, "local user", conf.User.Name)
}

func TestGetFinalConfigMergesAllConfigs(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	var err error
	err = stubGitConfig(fs, "/work/example", config.SystemScope, func() string {
		return heredoc.Doc(`
		[user]
		name = system user
		[author]
		name = leodido
		email = some@email.net
		`)
	})
	assert.Nil(t, err)

	err = stubGitConfig(fs, "/work/example", config.GlobalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = global user
		email = global@example.com
		`)
	})
	assert.Nil(t, err)

	err = stubGitConfig(fs, "/work/example", config.LocalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = local user
		`)
	})
	assert.Nil(t, err)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)

	conf, err := GetFinalConfig(repo)
	assert.Nil(t, err)

	assert.Equal(t, "local user", conf.User.Name)
	assert.Equal(t, "global@example.com", conf.User.Email)
	assert.Equal(t, "leodido", conf.Author.Name)
	assert.Equal(t, "some@email.net", conf.Author.Email)
}

func TestGetFinalConfigAppliesURLRules(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	var err error
	err = stubGitConfig(fs, "/work/example", config.SystemScope, func() string {
		return heredoc.Doc(`
		[user]
		name = system user
		[author]
		name = leodido
		email = some@email.net
		[url "git@bitbucket.org:"]
 		insteadOf = bb:
		[url "git@github.com:"]
		insteadOf	= gh:
		[url "git@gitlab.com:"]
		insteadOf = gl:
		`)
	})
	assert.Nil(t, err)

	err = stubGitConfig(fs, "/work/example", config.GlobalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = global user
		email = global@example.com
		`)
	})
	assert.Nil(t, err)

	err = stubGitConfig(fs, "/work/example", config.LocalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = local user
		[remote "origin"]
		url = gh:listendev/lstn
		fetch = +refs/heads/*:refs/remotes/origin/*
		`)
	})
	assert.Nil(t, err)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)

	conf, err := GetFinalConfig(repo)
	assert.Nil(t, err)

	assert.Equal(t, "local user", conf.User.Name)
	assert.Equal(t, "global@example.com", conf.User.Email)
	assert.Equal(t, "leodido", conf.Author.Name)
	assert.Equal(t, "some@email.net", conf.Author.Email)
	assert.Equal(t, "origin", conf.Remotes["origin"].Name)
	assert.Equal(t, []string{"git@github.com:listendev/lstn"}, conf.Remotes["origin"].URLs)
}

func TestGetFinalConfigWithPartiallyValidLocalConfig(t *testing.T) {
	fs := memfs.New()

	activeFS = fs
	defer func() { activeFS = defaultFS() }()

	var err error
	err = stubGitConfig(fs, "/work/example", config.SystemScope, func() string {
		return heredoc.Doc(`
		[user]
		name = system user
		[author]
		name = leodido
		email = some@email.net
		`)
	})
	assert.Nil(t, err)

	err = stubGitConfig(fs, "/work/example", config.GlobalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = global user
		email = global@example.com
		`)
	})
	assert.Nil(t, err)

	err = stubGitConfig(fs, "/work/example", config.LocalScope, func() string {
		return heredoc.Doc(`
		[user]
		name = local user
		[remote "origin"]
		url = gh:listendev/lstn
		fetch = +refs/heads/*:refs/remotes/origin/*
		[alias]
		allalias = "!f(){ \
						sed -n '/^\\[alias\\]$/,/^\\[/p' ~/.gitconfig | \
						sed 's/..//' | \
						sed '1d;$d' | \
						tr -s ' '; \
						}; \
						f"
		`)
	})
	assert.Nil(t, err)

	repo, err := gitInit(fs, "/work/example")
	assert.Nil(t, err)

	conf, err := GetFinalConfig(repo)
	assert.Nil(t, err)

	assert.Equal(t, "local user", conf.User.Name)
	assert.Equal(t, "global@example.com", conf.User.Email)
	assert.Equal(t, "leodido", conf.Author.Name)
	assert.Equal(t, "some@email.net", conf.Author.Email)
	assert.Equal(t, "origin", conf.Remotes["origin"].Name)
	assert.Equal(t, []string{"gh:listendev/lstn"}, conf.Remotes["origin"].URLs)
}

// stubGitConfig writes into the given fs a Git config
// at the local, global, or system Git config levels.
func stubGitConfig(fs billy.Filesystem, worktreeDir string, s config.Scope, content func() string) error {
	if s == config.SystemScope {
		// stub system level config
		paths, err := config.Paths(config.SystemScope)
		if len(paths) < 1 || err != nil {
			return fmt.Errorf("couldn't get system level Git config path")
		}
		if err := internaltesting.WriteFileContent(fs, paths[0], content(), false); err != nil {
			return err
		}
	}

	if s == config.GlobalScope {
		// stub global level config
		paths, err := config.Paths(config.GlobalScope)
		if len(paths) < 1 || err != nil {
			return fmt.Errorf("couldn't get system level Git config path")
		}
		if err := internaltesting.WriteFileContent(fs, paths[0], content(), false); err != nil {
			return err
		}
	}

	if s == config.LocalScope {
		// stub local level config
		path := filepath.Join(worktreeDir, ".git", "config")
		if err := internaltesting.WriteFileContent(fs, path, content(), false); err != nil {
			return err
		}
	}

	return nil
}

// gitInit initializes a Git repo pointing at the working tree path
// and at the nested .git/ directory inside the given fs.
func gitInit(fs billy.Filesystem, workingTreePath string) (*git.Repository, error) {
	workTreeFS, err := fs.Chroot(workingTreePath)
	if err != nil {
		return nil, err
	}

	gitDir := filepath.Join(workingTreePath, ".git")
	err = fs.MkdirAll(gitDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	gitDirFS, err := fs.Chroot(gitDir)
	if err != nil {
		return nil, err
	}

	s := filesystem.NewStorage(gitDirFS, cache.NewObjectLRUDefault())

	return git.Init(s, workTreeFS)
}
