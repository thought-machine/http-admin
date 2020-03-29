// Package main implements a simple binary that
package main

import (
	"github.com/peterebden/go-cli-init"
	"github.com/thought-machine/http-admin"
)

var opts = struct{
	Usage string
	Verbosity     cli.Verbosity `short:"v" long:"verbosity" default:"notice" description:"Verbosity of output (higher number = more output)"`
	Admin admin.Opts `group:"Options controlling HTTP admin server" namespace:"admin"`
}{
	Usage: `
This is a simple binary that only serves the admin server, as an example of how
one might set it up and use it.
`,
}

func main() {
	cli.ParseFlagsOrDie("example", &opts)
	info := cli.InitLogging(opts.Verbosity)
	opts.Admin.Logger = cli.MustGetLoggerNamed("github.com.thought-machine.http-admin")
	opts.Admin.LogInfo = info
	admin.Serve(opts.Admin)  // Normally you'd do this in a goroutine so as not to block the rest of the program.
}
