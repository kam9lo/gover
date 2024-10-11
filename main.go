package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kam9lo/gover/internal"
)

var (
	FlagConfigFile    = "./gover.yml"
	FlagCommitMessage = ""
	FlagPreRelease    = ""

	DefaultRepositoryPath = "."
)

func init() {
	flag.StringVar(&FlagConfigFile, "cfg", FlagConfigFile, "configuration file")
	flag.StringVar(&FlagCommitMessage, "msg-file", FlagCommitMessage, "commit message file path")
	flag.StringVar(&FlagPreRelease, "pre", FlagPreRelease, "pre-release version")

	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(args) < 1 || args[0] == "" {
		exit(errors.New("missing command"))
	}

	repositoryPath := DefaultRepositoryPath
	if len(args) > 1 {
		repositoryPath = args[1]
	}

	app, err := internal.NewApp(FlagConfigFile, repositoryPath)
	if err != nil {
		exit(err)
	}

	switch args[0] {
	case "version":
		err = app.Version()
	case "latest":
		err = app.LatestTag()
	case "next":
		err = app.Next(FlagPreRelease)
	case "commit":
		err = app.Commit(FlagCommitMessage)
	case "verify":
		err = app.Verify()
	case "change":
		err = app.Change()
	case "commits":
		err = app.Commits()
	case "tag":
		err = app.Tag(FlagPreRelease)
	default:
		exit(errors.New("invalid command"))
	}
	exit(err)
}

func exit(err error) {
	if err != nil {
		fmt.Println(err.Error())
		fmt.Print(usage)
		os.Exit(1)
	}
	os.Exit(0)
}

const usage = `
Usage:

gover [COMMAND] [PATH] [VERSION]

Commands:
	version Print version
	next 	Print next version based on feature branch commit log.
	latest  Print latest version tag
	commit	Print prompt and generate commit message from template
			and provided values
	verify	Verify commit messages since last
	change	Print type of most important change made since last version
	tag		Tag commit with version based on commits since previous tag

Examples:

Show commit message prompt (ctrl-c to skip):
$ gover commit .

Show next version:
$ gover next .
v0.1.0

Show change type:
$ gover change .
minor

Create next version tag:
$ gover tag .
`
