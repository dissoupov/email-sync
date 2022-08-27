package main

import (
	"io"
	"os"

	"github.com/ableorg/email-sync/internal/version"
	"github.com/ableorg/email-sync/pkg/cli"
	"github.com/alecthomas/kong"
	"github.com/effective-security/xpki/x/ctl"
)

type app struct {
	cli.Cli

	Auth cli.AuthCmd `cmd:"" help:"Authentication commands"`
}

func main() {
	realMain(os.Args, os.Stdout, os.Stderr, os.Exit)
}

func realMain(args []string, out io.Writer, errout io.Writer, exit func(int)) {
	cl := app{
		Cli: cli.Cli{
			Version: ctl.VersionFlag("0.0.1"),
		},
	}
	cl.Cli.WithErrWriter(errout).
		WithWriter(out)

	parser, err := kong.New(&cl,
		kong.Name("email-sync"),
		kong.Description("CLI tool for email-sync service"),
		//kong.UsageOnError(),
		kong.Writers(out, errout),
		kong.Exit(exit),
		ctl.BoolPtrMapper,
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": version.Current().String(),
		})
	if err != nil {
		panic(err)
	}

	ctx, err := parser.Parse(args[1:])
	parser.FatalIfErrorf(err)

	if ctx != nil {
		err = ctx.Run(&cl.Cli)
		ctx.FatalIfErrorf(err)
	}
}
