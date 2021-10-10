package commands

import (
	"fmt"
	"sync"
	"text/tabwriter"

	cli2 "github.com/RobolabGs2/botctl/cli"
	games2 "github.com/RobolabGs2/botctl/games"
)

type DuelConfig struct {
	Rounds   int          `name:"r" default:"1" desc:"Количество раундов"`
	Bot1Logs LogCollector `name:"o1" default:"+" desc:"Куда перенаправить stdout 1 бота '-' = stdout, '+' - stderr, '<dirname>' - будет сохранено в папку, файл на раунд"`
	Bot2Logs LogCollector `name:"o2" desc:"Куда перенаправить stdout 2 бота '-' = stdout, '+' - stderr, '<dirname>' - будет сохранено в папку, файл на раунд"`
	//Parallel int          `name:"p" default:"1" desc:"Параллельный запуск раундов."` TODO
}

func (d *DuelConfig) Usage() string {
	return `"path/to/mybot1.exe [addition args]" "path/to/mybot2.exe [addition args]"`
}

func (d *DuelConfig) Run(args []string, streams cli2.Streams) error {
	if len(args) != 2 {
		return cli2.IncorrectUsageErr{fmt.Errorf("expected args 2, actual %d", len(args))}
	}
	bot1, err := games2.NewBot(args[0])
	if err != nil {
		return err
	}
	if err := d.Bot1Logs.Prepare(); err != nil {
		return err
	}
	bot2, err := games2.NewBot(args[1])
	if err != nil {
		return err
	}
	if err := d.Bot2Logs.Prepare(); err != nil {
		return err
	}
	battlesQueue := make(chan *games2.Battle, d.Rounds)
	finishedBattles := make(chan *games2.Battle, d.Rounds)
	runnersGroup := new(sync.WaitGroup)
	runnersGroup.Add(1)
	battles := make([]games2.Battle, d.Rounds)
	for i := range battles {
		battles[i].Players[i%2] = bot1
		battles[i].Players[(i+1)%2] = bot2
		scrapper, err := d.Bot1Logs.GetWriter(i, fmt.Sprintf("%d_%s", 1, bot1.Name), streams)
		if err == nil {
			err = battles[i].LogScrapper(0, scrapper)
		}
		if err != nil {
			return err
		}
		scrapper, err = d.Bot2Logs.GetWriter(i, fmt.Sprintf("%d_%s", 2, bot1.Name), streams)
		if err == nil {
			err = battles[i].LogScrapper(1, scrapper)
		}
		if err != nil {
			return err
		}
		battlesQueue <- &battles[i]
	}
	close(battlesQueue)
	games2.Runner(runnersGroup, battlesQueue, finishedBattles)
	summary := tabwriter.NewWriter(streams.Out, 0, 4, 4, ' ', 0)
	_, _ = fmt.Fprintf(summary, "Раунд\t%s\t%s\n", bot1.Cmd, bot2.Cmd)
	bot1Score := 0.0
	bot2Score := 0.0
	for i := range battles {
		battle := &battles[i]
		if battle.State() == games2.BattleError {
			return fmt.Errorf("проблемы с раундом %d: %w", i, battle.Err())
		}
		if battle.State() != games2.BattleFinished {
			panic(fmt.Errorf("unexpected battle state: %s", battle.State()))
		}
		first, second := 0, 1
		if battle.Players[first] != bot1 {
			first, second = 1, 0
		}
		result1 := battle.GameResult(first)
		bot1Score += result1.Score()
		result2 := battle.GameResult(second)
		bot2Score += result2.Score()
		_, _ = fmt.Fprintf(summary, "%d\t%s\t%s\n", i,
			fmt.Sprintf("%s, ходил %s", result1, games2.TurnOrder(first)),
			fmt.Sprintf("%s, ходил %s", result2, games2.TurnOrder(second)),
		)
	}
	_, _ = fmt.Fprintf(summary, "Итого\t%.1f\t%.1f\n", bot1Score, bot2Score)
	_ = summary.Flush()
	return err
}
