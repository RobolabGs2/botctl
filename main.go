package main

import (
	"fmt"
	"os"

	"github.com/RobolabGs2/botctl/cli"
	"github.com/RobolabGs2/botctl/commands"
)

func main() {
	cmds := map[string]cli.Command{
		"duel":       new(commands.DuelConfig),
		"tournament": new(commands.Tournament),
	}
	commandName := "duel"
	helpName := os.Args[0]
	args := os.Args[1:]
	if len(args) > 0 {
		if _, ok := cmds[args[0]]; ok {
			commandName = args[0]
			args = args[1:]
			helpName += " " + commandName
		}
	}
	err := cli.RunCommand(helpName, cmds[commandName], cli.StdStreams, args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
