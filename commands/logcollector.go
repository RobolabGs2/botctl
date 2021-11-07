package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/RobolabGs2/botctl/cli"
)

type LogCollector string

func (v LogCollector) String() string {
	return string(v)
}

func (v *LogCollector) Set(s string) error {
	*v = LogCollector(s)
	return nil
}

func (v LogCollector) Prepare() error {
	if v == "" || v == "-" || v == "+" {
		return nil
	}
	return os.MkdirAll(string(v), 0666)
}

func (v LogCollector) GetWriter(round int, botName string, streams cli.Streams) (io.Writer, error) {
	switch v {
	case "":
		return nil, nil
	case "-":
		return streams.Out, nil
	case "+":
		return streams.Err, nil
	default:
		file, err := os.Create(filepath.Join(string(v), fmt.Sprintf("round_%02d_bot_%s.txt", round, botName)))
		if err != nil {
			return nil, fmt.Errorf("failed to create log file for %s: %w", botName, err)
		}
		return file, nil
	}
}
