package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/cachecashproject/go-cachecash/ranger"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.HideVersion = true
	app.ArgsUsage = "[options] [YAML definition] [file prefix]"
	// XXX small fix for usage text in a subcommand-less zone
	app.UsageText = fmt.Sprintf("%s %s", path.Base(os.Args[0]), app.ArgsUsage)
	app.Usage = "Generate byte range marshaling code from an object specification"

	app.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name:  "code, c",
			Usage: "Generate source code for marshaling",
		},
		cli.BoolTFlag{
			Name:  "test, t",
			Usage: "Generate test code for marshaling",
		},
		cli.BoolTFlag{
			Name:  "fuzz, f",
			Usage: "Generate code suitable for fuzz testing with go-fuzz",
		},
	}

	app.HideHelp = true
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return errors.New("invalid arguments - try --help")
	}

	rf, err := ranger.ParseFile(ctx.Args()[0])
	if err != nil {
		return errors.Wrap(err, "could not parse YAML definition")
	}

	p := ctx.Args()[1]

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return errors.Wrap(err, "could not create parent directories")
	}

	if ctx.BoolT("code") {
		content, err := rf.GenerateCode()
		if err != nil {
			return errors.Wrap(err, "could not generate code")
		}

		if err := ioutil.WriteFile(fmt.Sprintf("%s.go", p), content, 0644); err != nil {
			return errors.Wrap(err, "could not write code segment")
		}
	}

	if ctx.BoolT("test") {
		content, err := rf.GenerateTest()
		if err != nil {
			return errors.Wrap(err, "could not generate tests")
		}

		if err := ioutil.WriteFile(fmt.Sprintf("%s_test.go", p), content, 0644); err != nil {
			return errors.Wrap(err, "could not write tests segment")
		}
	}

	if ctx.BoolT("fuzz") {
		content, err := rf.GenerateFuzz()
		if err != nil {
			return errors.Wrap(err, "could not generate fuzz tests")
		}

		if err := ioutil.WriteFile(fmt.Sprintf("%s_fuzz.go", p), content, 0644); err != nil {
			return errors.Wrap(err, "could not write fuzz tests segment")
		}
	}
	return nil
}
