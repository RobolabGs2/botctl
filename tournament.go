package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"text/tabwriter"
)

type TournamentSettings struct {
	Bot1, Bot2                                     *Bot
	TotalWriter                                    io.Writer
	ProcessWriter                                  io.Writer
	Bot1LogCollectorFabric, Bot2LogCollectorFabric LogCollectorFabric
}

type LogCollectorFabric func(round int) io.Writer

func IgnoreLogs(_ int) io.Writer {
	return nil
}

func LogsRedirectTotWriter(writer io.Writer) func(_ int) io.Writer {
	return func(_ int) io.Writer {
		return writer
	}
}

type Tournament struct {
	TournamentSettings
	summary *tabwriter.Writer
	logger  *log.Logger
}

func (s TournamentSettings) Init() *Tournament {
	return &Tournament{
		s,
		tabwriter.NewWriter(s.TotalWriter, 0, 4, 4, ' ', 0),
		log.New(s.ProcessWriter, "botctl: ", log.Ltime|log.Lmsgprefix),
	}
}

func (t *Tournament) Run(rounds int) error {
	fmt.Fprintf(t.summary, "Раунд\t%s\t%s\n", t.Bot1.Cmd, t.Bot2.Cmd)
	for i := 0; i < rounds; i++ {
		t.logger.Println("Раунд", i)
		res, err := t.Round(i)
		if err != nil {
			return err
		}
		t.logger.Println(res)
	}
	fmt.Fprintf(t.summary, "Итого\t%.1f\t%.1f\n", t.Bot1.TotalScore, t.Bot2.TotalScore)
	return t.summary.Flush()
}

func (t *Tournament) Round(i int) (string, error) {
	bot1TurnOrder := TurnOrder(i%2 == 0)
	bot1, bot2 := t.Bot1.MakeCmd(bot1TurnOrder), t.Bot2.MakeCmd(bot1TurnOrder.Opponent())
	bot1.Stderr, _ = bot2.StdinPipe()
	bot2.Stderr, _ = bot1.StdinPipe()
	bot1.Stdout = t.Bot1LogCollectorFabric(i)
	if closer, ok := bot1.Stdout.(io.WriteCloser); ok {
		defer closer.Close()
	}
	bot2.Stdout = t.Bot2LogCollectorFabric(i)
	if closer, ok := bot2.Stdout.(io.WriteCloser); ok {
		defer closer.Close()
	}
	if err := bot1.Start(); err != nil {
		return "", err
	}
	defer bot1.Process.Kill()
	if err := bot2.Start(); err != nil {
		return "", err
	}
	defer bot2.Process.Kill()
	// TODO
	score1, score1Desc, err1 := bot1.Finish(context.Background())
	score2, score2Desc, err2 := bot2.Finish(context.Background())
	if err1 != nil {
		return "", err1
	}
	if err2 != nil {
		return "", err2
	}
	fmt.Fprintf(t.summary, "%d\t%s\t%s\n", i, score1Desc, score2Desc)
	if score1+score2 != Draw {
		return "", fmt.Errorf("конфликт показаний: %s (%s) против %s (%s)", bot1.Name, score1, bot2.Name, score2)
	}
	if score1 == score2 {
		return score1.String(), nil
	} else if score1 == Win {
		return fmt.Sprint(bot1.Name, " ", score1Desc), nil
	} else {
		return fmt.Sprint(bot2.Name, " ", score2Desc), nil
	}
}
