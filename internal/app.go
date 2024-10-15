package internal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/kam9lo/gover/internal/config"
	"github.com/kam9lo/gover/internal/prompt"
	"github.com/kam9lo/gover/internal/repository"
	"github.com/kam9lo/gover/internal/version"
)

var (
	Name    string
	Version string
)

// App is a main application's structure.
type App struct {
	cfg  *config.Config
	repo *repository.Repository
}

// NewApp returns new instance of application.
func NewApp(cfgPath, repoPath string) (*App, error) {
	cfg, err := config.NewFromFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("new config from file: %w", err)
	}
	repo, err := repository.Open(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repository: %w", err)
	}
	return &App{
		cfg:  cfg,
		repo: repo,
	}, nil
}

// Commit displays configured in configuration file prompt with values to fill
// commit message template. If msgFile argument is non-empty, generated text is
// written directly into the file.
func (a *App) Commit(msgFile string) (err error) {
	mp := map[string]string{}
	for _, arg := range a.cfg.Args {
		if len(arg.Options) != 0 {
			mp[arg.Name], err = prompt.Select(arg.Name, arg.Options)
			if err != nil {
				return
			}
		} else {
			mp[arg.Name], err = prompt.TextInput(arg.Name, arg.Required)
			if err != nil {
				return
			}
		}
	}

	tmpl, err := template.New("commit message").Parse(a.cfg.Templates.Commit)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	buff := bytes.NewBuffer(nil)
	if err = tmpl.Execute(buff, mp); err != nil {
		return fmt.Errorf(
			"couldn't execute template\n%s\nwith args:\n%v",
			a.cfg.Templates.Commit, mp,
		)
	}
	if msgFile != "" {
		return os.WriteFile(msgFile, buff.Bytes(), 0o644)
	}

	return nil
}

// Next is the next version based on configuration and current branch commit
// messages.
func (a *App) Next(pre string) error {
	change, err := a.change(true)
	if err != nil {
		return err
	}
	version, err := a.version(change.String(), pre)
	if err != nil {
		return err
	}

	fmt.Println(version)

	return nil
}

// Verify iterates over current branch commits up to latest tagged and
// validates their message format and template matching.
func (a *App) Verify() error {
	_, err := a.change(false)
	return err
}

// Change displays resolved from current branch change type.
func (a *App) Change() error {
	change, err := a.change(true)
	if err != nil {
		return err
	}

	fmt.Println(change)

	return nil
}

func (a *App) Commits() error {
	commits, err := a.repo.FeatureCommits()
	if err != nil {
		return fmt.Errorf("feature commits: %w", err)
	}
	for _, commit := range commits {
		fmt.Println(commit)
	}
	return nil
}

func (a *App) Version() error {
	fmt.Printf("%s %s\n", Name, Version)

	return nil
}

func (a *App) Changelog() error {
	if a.cfg.Templates.Changelog == "" {
		return a.Commits()
	}

	commits, err := a.featureCommits()
	if err != nil {
		return err
	}

	sortedMessages := map[string]map[string][]repository.Message{}
	for _, commit := range commits {
		for field, value := range commit {
			if sortedMessages[field] == nil {
				sortedMessages[field] = map[string][]repository.Message{}
			}
			sortedMessages[field][value] = append(sortedMessages[field][value], commit)
		}
	}

	tmpl, err := template.New("changelog").Parse(a.cfg.Templates.Changelog)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	buff := bytes.NewBuffer(nil)
	if err = tmpl.Execute(buff, sortedMessages); err != nil {
		return fmt.Errorf(
			"couldn't execute template\n%s\nwith args:\n%v\n%w",
			a.cfg.Templates.Changelog, sortedMessages,
			err,
		)
	}

	fmt.Println(buff.String())

	return nil
}

func (a *App) LatestTag() error {
	tag, err := a.repo.LatestTag()
	if err != nil {
		return err
	}

	fmt.Println(tag)

	return nil
}

// Tag creates new version tag on last commit.
func (a *App) Tag(pre string) error {
	change, err := a.change(true)
	if err != nil {
		return err
	}
	version, err := a.version(change.String(), pre)
	if err != nil {
		return err
	}
	if err := a.repo.CreateTag(version); err != nil {
		return err
	}
	return nil
}

func (a *App) featureCommits() ([]repository.Message, error) {
	commits, err := a.repo.FeatureCommits()
	if err != nil {
		return nil, err
	}

	messages := make([]repository.Message, 0, len(commits))
	for _, commmit := range commits {
		msg, err := repository.ParseMessage(
			a.cfg.Templates.Commit,
			commmit,
			a.cfg.RequiredArgs()...,
		)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (a *App) version(change string, pre string) (string, error) {
	ct := version.ChangeTypeNone
	ct.Parse(change)

	latestTag, err := a.repo.LatestTag()
	if err != nil {
		if errors.Is(err, repository.ErrCommitNotFound) {
			return "", nil
		}
		return "", err
	}

	tag, err := version.New(latestTag)
	if err != nil {
		return "", err
	}

	return tag.Next(ct, pre).String(), nil
}

func (a *App) change(allowMismatch bool) (version.ChangeType, error) {
	msgs, err := a.repo.FeatureCommits()
	if err != nil {
		if errors.Is(err, repository.ErrCommitNotFound) {
			return version.ChangeTypeNone, nil
		}
		return version.ChangeTypeNone, err
	}

	versionTypes := map[argName]map[optName]version.ChangeType{}
	for _, arg := range a.cfg.Args {
		versionTypes[arg.Name] = argOptionsTypes(arg.Options)
	}

	change := version.ChangeTypeNone
	for _, msg := range msgs {
		m, err := repository.ParseMessage(
			a.cfg.Templates.Commit,
			msg,
			a.cfg.RequiredArgs()...,
		)
		if err != nil && !allowMismatch {
			return version.ChangeTypeNone, err
		}
		for tmpl, val := range m {
			msgChange := versionTypes[tmpl][val]
			if msgChange > change {
				change = msgChange
			}
		}
	}

	return change, nil
}

func argOptionsTypes(opts []config.Option) map[string]version.ChangeType {
	m := map[string]version.ChangeType{}
	for _, opt := range opts {
		ac := version.ChangeTypeNone
		ac.Parse(opt.Version)

		if ac != version.ChangeTypeNone {
			m[opt.Value] = ac
		}
	}
	return m
}

type (
	argName = string
	optName = string
)

const invalidCommitErrMsg = `invalid commit message:

%s
-----------------------------------
template:
%s
-----------------------------------
options for "%s":
%s
-----------------------------------
got: "%s"`
