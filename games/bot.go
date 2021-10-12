package games

import (
	"context"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/RobolabGs2/botctl/executil"
)

type Bot struct {
	Name string
	File string
	Cmd  string
}

type BotCmd struct {
	*exec.Cmd
	mut      sync.Mutex
	finished bool
	exitCode error
	wait     chan struct{}
}

func (b *BotCmd) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := b.Cmd.Start(); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		b.mut.Lock()
		defer b.mut.Unlock()
		if !b.finished {
			log.Println("KILLING PROCESS", ctx.Err())
			_ = executil.KillProcess(b.Cmd)
		}
	}()
	go func() {
		err := b.Cmd.Wait()
		b.mut.Lock()
		b.finished = true
		b.exitCode = err
		b.mut.Unlock()
		close(b.wait)
	}()
	return nil
}

func (b *BotCmd) Wait() error {
	<-b.wait // после чтения из канала гарантированно всё изменено как надо
	return b.exitCode
}

func (b *BotCmd) Finish() (GameResult, error) {
	score, err := SummarizeGame(b.Wait())
	if err != nil {
		return 0, err
	}
	return score, nil
}

func NewBot(cmd string, name ...string) (Bot, error) {
	path := strings.Split(cmd, " ")[0]
	filename := filepath.Base(path)
	bot := Bot{
		Name: strings.TrimSuffix(filename, filepath.Ext(filename)),
		File: filename,
		Cmd:  filepath.Clean(cmd),
	}
	if len(name) != 0 {
		bot.Name = name[0]
	}
	return bot, executil.CheckFile(path)
}

func (b Bot) MakeCmd(order TurnOrder) *BotCmd {
	return &BotCmd{
		Cmd:  executil.MakeCmd(append(strings.Split(b.Cmd, " "), order.CmdArg())...),
		wait: make(chan struct{}),
	}
}
