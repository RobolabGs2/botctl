package games

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type BattleState int

var battleStateString = [...]string{"Pending", "Running", "Finished", "Error"}

const (
	BattleAwait BattleState = iota
	BattleRunning
	BattleFinished
	BattleError
)

func (s BattleState) String() string {
	return battleStateString[s]
}

type Battle struct {
	stateMut sync.RWMutex
	state    BattleState
	err      error
	scores   [2]GameResult
	logs     [2]Buffer

	Players    [2]Bot
	StartedAt  time.Time
	FinishedAt time.Time
}

func (b *Battle) Duration() time.Duration {
	return b.FinishedAt.Sub(b.StartedAt)
}

func (battle *Battle) Logs(player int) string {
	return battle.logs[player].ReadAll()
}

func (battle *Battle) State() BattleState {
	battle.stateMut.RLock()
	defer battle.stateMut.RUnlock()
	return battle.state
}

func (battle *Battle) LogScrapper(player int, scrapper io.Writer) error {
	if battle.logs[player].onWrite != nil {
		return errors.New("scrapper already set for player")
	}
	battle.logs[player].onWrite = scrapper
	return nil
}

func (battle *Battle) Err() error {
	battle.stateMut.RLock()
	defer battle.stateMut.RUnlock()
	return battle.err
}

func (battle *Battle) GameResult(player int) GameResult {
	battle.stateMut.RLock()
	defer battle.stateMut.RUnlock()
	if battle.state != BattleFinished {
		panic(errors.New("call GameResult on battle that does not finished"))
	}
	return battle.scores[player]
}

type Buffer struct {
	buf     bytes.Buffer
	mut     sync.RWMutex
	onWrite io.Writer
}

func (b *Buffer) Write(bytes []byte) (int, error) {
	b.mut.Lock()
	defer b.mut.Unlock()
	if b.onWrite != nil {
		_, _ = b.onWrite.Write(bytes)
	}
	return b.buf.Write(bytes)
}

func (b *Buffer) ReadAll() string {
	b.mut.RLock()
	defer b.mut.RUnlock()
	return b.buf.String()
}

func (battle *Battle) Run(ctx context.Context) (err error) {
	battle.stateMut.Lock()
	battle.state = BattleRunning
	battle.StartedAt = time.Now()
	battle.stateMut.Unlock()
	defer func() {
		battle.stateMut.Lock()
		battle.FinishedAt = time.Now()
		battle.stateMut.Unlock()
	}()
	defer func() {
		if err != nil {
			battle.stateMut.Lock()
			battle.state = BattleError
			battle.err = err
			battle.stateMut.Unlock()
		}
	}()
	bot1, bot2 := battle.Players[0].MakeCmd(First), battle.Players[1].MakeCmd(Second)
	bot1.Stderr, _ = bot2.StdinPipe()
	bot2.Stderr, _ = bot1.StdinPipe()
	bot1.Stdout = &battle.logs[0]
	bot2.Stdout = &battle.logs[1]
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := bot1.Start(localCtx); err != nil {
		return err
	}
	if err := bot2.Start(localCtx); err != nil {
		return err
	}
	score1, err1 := bot1.Finish()
	if err1 != nil {
		return err1
	}
	score2, err2 := bot2.Finish()
	if err2 != nil {
		return err2
	}
	if score1+score2 != 0 {
		return &ConflictScore{
			Players: battle.Players,
			Scores:  [2]GameResult{score1, score2},
			Battle:  battle,
		}
	}
	battle.stateMut.Lock()
	battle.scores[0] = score1
	battle.scores[1] = score2
	battle.state = BattleFinished
	battle.stateMut.Unlock()
	return nil
}

type ConflictScore struct {
	Players [2]Bot
	Scores  [2]GameResult
	Battle  *Battle
}

func (c *ConflictScore) Error() string {
	return fmt.Sprintf("conflict scores: %q says what it %s, but %q says what it %s", c.Players[0].Name, c.Scores[0], c.Players[1].Name, c.Scores[1])
}
