package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/RobolabGs2/flagconfig"
)

// Command - CLI команда
type Command interface {
	Run(args []string, streams Streams) error
	Usage() string
}

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

var StdStreams = Streams{
	In:  os.Stdin,
	Out: os.Stdout,
	Err: os.Stderr,
}

type IncorrectUsageErr struct {
	Reason error
}

func (e IncorrectUsageErr) Error() string {
	return fmt.Sprintf("incorrect usage: %s", e.Reason)
}

func (e IncorrectUsageErr) Unwrap() error {
	return e.Reason
}

// err - влияет на exit code? Вроде бы всё обработали
func RunCommand(name string, cmd Command, streams Streams, args []string) error {
	flags, err := flagconfig.MakeFlags(cmd, "", flag.ContinueOnError)
	if err != nil {
		return err
	}
	flags.Usage = func() {}
	if err := flags.Parse(args); err != nil {
		PrintUsage(name, cmd, streams.Err, flags)
		return err
	}
	err = cmd.Run(flags.Args(), streams)
	if badUsage := new(IncorrectUsageErr); errors.As(err, badUsage) {
		_, _ = fmt.Fprintln(streams.Err, badUsage.Reason)
		PrintUsage(name, cmd, streams.Err, flags)
	}
	return err
}

func PrintUsage(name string, cmd Command, writer io.Writer, flags *flag.FlagSet) {
	_, _ = fmt.Fprintf(writer, "Usage:\n %s [flags] %s\n", name, cmd.Usage())
	flags.SetOutput(writer)
	flags.PrintDefaults()
}
