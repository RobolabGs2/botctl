package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/RobolabGs2/botctl/cli"
	"github.com/RobolabGs2/botctl/commands"
	"github.com/RobolabGs2/flagconfig"
)

func main() {
	cmds := map[string]cli.Command{
		"duel":       new(commands.DuelConfig),
		"tournament": new(commands.Tournament),
		"test":       new(commands.TestBot),
	}
	commandName := "duel"
	helpName := os.Args[0]
	args := os.Args[1:]
	if len(args) > 0 {
		if _, ok := cmds[args[0]]; ok {
			commandName = args[0]
			args = args[1:]
			helpName += " " + commandName
		} else if args[0] == "help" {
			writer := os.Stdout
			if len(args) == 1 {
				list := make([]string, 0, len(cmds))
				for key := range cmds {
					list = append(list, key)
				}
				sort.Strings(list)
				fmt.Fprintln(writer, "Доступные команды:")
				helpTable := tabwriter.NewWriter(writer, 0, 8, 8, ' ', 0)
				for _, cmdName := range list {
					cmd := cmds[cmdName]
					_, _ = fmt.Fprintf(helpTable, "\t%s\t%s\n", cmdName, strings.Split(cmd.Description(), "\n")[0])
				}
				helpTable.Flush()
				fmt.Fprintf(writer, "\nИспользуйте \"%s help <command>\" чтобы больше узнать о конкретной команде.\n",
					helpName)
				return
			}
			commandName = args[1]
			helpName += " " + commandName
			command := cmds[commandName]
			flags, _ := flagconfig.MakeFlags(command, "", flag.ContinueOnError)
			_, _ = fmt.Fprintf(writer, "использование:\n %s [flags] %s\n", helpName, command.Usage())
			flags.SetOutput(writer)
			flags.PrintDefaults()
			_, _ = fmt.Fprintln(writer, "\n", command.Description())
			return
		}
	}
	err := cli.RunCommand(helpName, cmds[commandName], cli.StdStreams, args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
