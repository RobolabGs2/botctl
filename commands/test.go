package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/RobolabGs2/botctl/cli"
	"github.com/RobolabGs2/botctl/games"
)

type TestBot struct {
	Bot1Logs LogCollector  `name:"o1" default:"-" desc:"Перенаправить stdout бота за первый ход '-' = stdout, '+' - stderr, '<dirname>' - будет сохранено в папку, файл на раунд"`
	Bot2Logs LogCollector  `name:"o2" desc:"Перенаправить stdout бота за второй ход '-' = stdout, '+' - stderr, '<dirname>' - будет сохранено в папку, файл на раунд"`
	Timeout  time.Duration `name:"t" default:"30m" desc:"Таймаут, после которого считать бота зависшим и прекратить тестирование"`
}

func (d TestBot) Usage() string {
	return `"path/to/mybot.exe [addition args]"`
}

func (d TestBot) Description() string {
	return "Тестирует бота, запуская игру с самим собой"
}

func (d *TestBot) Run(args []string, streams cli.Streams) error {
	if len(args) != 1 {
		return cli.IncorrectUsageErr{Reason: fmt.Errorf("expected args 1, actual %d", len(args))}
	}
	bot1, err := games.NewBot(args[0])
	if err != nil {
		return err
	}
	if err := d.Bot1Logs.Prepare(); err != nil {
		return err
	}
	if err := d.Bot2Logs.Prepare(); err != nil {
		return err
	}
	battle := games.Battle{Players: [2]games.Bot{bot1, bot1}}
	scrapper, err := d.Bot1Logs.GetWriter(0, fmt.Sprintf("%d_%s", 1, bot1.Name), streams)
	if err == nil {
		err = battle.LogScrapper(0, scrapper)
	}
	if err != nil {
		return err
	}
	scrapper, err = d.Bot2Logs.GetWriter(0, fmt.Sprintf("%d_%s", 2, bot1.Name), streams)
	if err == nil {
		err = battle.LogScrapper(1, scrapper)
	}
	if err != nil {
		return err
	}
	timeout, cancelFunc := context.WithTimeout(context.Background(), d.Timeout)
	err = battle.Run(timeout)
	cancelFunc()
	output := log.New(streams.Out, "", 0)
	if err != nil {
		if conflict := new(games.ConflictScore); errors.As(err, &conflict) {
			output.Printf(`Либо бот упал с ошибкой, либо он некорректно возвращает результат игры.
Exit code должен быть: победа - 0, поражение - 3, ничья - 4.
Сейчас же бот ходивший первым утверждает %q, а вторым: %q
`, conflict.Scores[0].String(), conflict.Scores[1].String())
			return err
		}
		if errors.Is(timeout.Err(), context.DeadlineExceeded) {
			output.Println("Бот слишком долго играл, возможно, завис")
			return timeout.Err()
		}
		_, _ = fmt.Fprintln(streams.Out, "Проблемы с ботом", err)
		return err
	}
	_, _ = fmt.Fprintf(streams.Out, `Тестирование прошло успешно, бот способен играть сам с собой.
Результат ходившего первым: %s
`, battle.GameResult(0))
	return nil
}
