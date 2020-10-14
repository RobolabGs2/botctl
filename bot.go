package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Bot struct {
	Name       string
	File       string
	Cmd        string
	Number     int
	TotalScore float32
}

type BotCmd struct {
	*Bot
	*exec.Cmd
	Order TurnOrder
}

func (b BotCmd) Start() error {
	return b.error(b.Cmd.Start())
}

func (b BotCmd) String() string {
	return b.Name
}

func (b BotCmd) WaitChan() chan error {
	errs := make(chan error)
	go func() {
		err := b.Cmd.Wait()
		select {
		case errs <- err:
			break
		default:
			break
		}
		close(errs)
	}()
	return errs
}

type BotGameResult struct {
	Error      error
	GameResult GameResult
	Bot        *Bot
}

func (b BotCmd) Finish(ctx context.Context) (GameResult, string, error) {
	select {
	case <-ctx.Done():
		return 0, "", ctx.Err()
	case res := <-b.WaitChan():
		score, err := SummarizeGame(res)
		if err != nil {
			return 0, "", b.error(err)
		}
		switch score {
		case Win:
			b.TotalScore++
		case Draw:
			b.TotalScore += 0.5
		}
		return score, fmt.Sprintf("%s, ходил %s", score, b.Order), nil
	}
}

func (b Bot) AppendArgs(args []string, order TurnOrder) []string {
	args = append(args, strings.Split(b.Cmd, " ")...)
	return append(args, order.CmdArg())
}

func NewBot(number int, cmd string) (*Bot, error) {
	filename := filepath.Base(strings.Split(cmd, " ")[0])
	bot := Bot{
		Name:   strings.TrimSuffix(filename, filepath.Ext(filename)),
		File:   filename,
		Cmd:    filepath.Clean(cmd),
		Number: number,
	}
	return &bot, bot.checkFile()
}

func (b Bot) checkFile() error {
	if _, err := os.Stat(b.File); err != nil {
		return b.error(err)
	}
	return nil
}

func (b Bot) error(err error) error {
	if err != nil {
		return fmt.Errorf("проблемы с ботом %d: %s: %w", b.Number, b.Name, err)
	}
	return nil
}

func (b *Bot) MakeCmd(order TurnOrder) BotCmd {
	// Платформозависимо - запускаем с помощью cmd
	args := []string{"/C"}
	return BotCmd{
		Bot:   b,
		Cmd:   exec.Command("cmd", b.AppendArgs(args, order)...),
		Order: order,
	}
}
