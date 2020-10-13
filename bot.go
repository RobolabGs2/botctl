package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Bot struct {
	Name       string
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

func (b BotCmd) Finish() (GameResult, string, error) {
	score, err := SummarizeGame(b.Wait())
	if err != nil {
		return 0, "", err
	}
	switch score {
	case Win:
		b.TotalScore++
	case Draw:
		b.TotalScore += 0.5
	}
	return score, fmt.Sprintf("%s, ходил %s", score, b.Order), nil
}

func (b Bot) AppendArgs(args []string, order TurnOrder) []string {
	args = append(args, strings.Split(b.Cmd, " ")...)
	return append(args, order.CmdArg())
}

func NewBot(number int, cmd string) (*Bot, error) {
	bot := Bot{
		Name:   strings.Split(cmd, " ")[0],
		Cmd:    filepath.Clean(cmd),
		Number: number,
	}
	return &bot, bot.checkFile()
}

func (b Bot) checkFile() error {
	if _, err := os.Stat(b.Name); err != nil {
		return b.error(err)
	}
	return nil
}

func (b Bot) error(err error) error {
	if err != nil {
		return fmt.Errorf("проблемы с ботом %d.%s:%w", b.Number, b.Name, err)
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
