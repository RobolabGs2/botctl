package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"runtime"
	"sync"
	"text/template"
	"time"

	"github.com/RobolabGs2/botctl/cli"
	"github.com/RobolabGs2/botctl/executil"
	"github.com/RobolabGs2/botctl/games"
	"gopkg.in/yaml.v3"
)

type Tournament struct {
	Concurrency int    `name:"c" default:"0" desc:"Количество одновременных соревнований. '0' - количество виртуальных ядер минус 1"`
	Config      string `name:"config" default:"tournament.yaml" desc:"Список ботов для турнира"`
	config      TournamentConfigs
}

func (t *Tournament) Description() string {
	return `Запускает турнир всех со всеми, ведя турнирную таблицу.
Берёт список ботов из файла, задаваемого флагом -config.
В файле ожидается словарь в формате yaml, ключи которого будут использоваться в качестве
имён ботов, а значениями должны быть настройки (пока настройка только одна - команда запуска бота):

bots:
	first bot: # имя бота или его автора
		cmd: bot1.exe # команда запуска без последнего аргумента - номера хода
	Студент Студентович Студентов:
		cmd: bot2.exe
	first bot with args:
		cmd: bot1.exe -d 42
`
}

func (t *Tournament) Run(args []string, streams cli.Streams) error {
	dirName := "."
	if len(args) == 1 {
		dirName = args[0]
	}
	dir := os.DirFS(dirName)
	botsDesc, err := t.readBotDescriptions(dir)
	if err != nil {
		return err
	}
	if len(botsDesc) < 1 {
		return errors.New("not enough bots for tournament")
	}
	bots, err := MakeBots(botsDesc, dirName)
	if err != nil {
		return err
	}
	report := runTournament(bots, botsDesc, streams, t.config.Timeout, t.Concurrency, t.config.Rounds)
	report.Game = t.config.Game
	finished := time.Now()
	filename := fmt.Sprintf("report-%s.html", finished.Format("2006-01-02_15h04m05s"))
	file, err := os.Create(filename)
	log.Println("Отчёт сохранён в файл ", filename)
	return tmpl.Execute(file, report)
}

var tmpl = template.Must(template.New("report").Parse(reportBattleTmpl))

const reportBattleTmpl = `{{- /*gotype: github.com/RobolabGs2/botctl/commands.TournamentReport*/ -}}
<html lang="ru-en">
<head>
	<title>Games report {{.Game}}</title>
	<style>
		body > section {
			margin: 32px 4px;
		}
        body > section > article {
            margin: 4px;
        }
	</style>
</head>
<body>
<script>
    var hideSibling = function (t) {
        t.nextElementSibling.style.display = t.nextElementSibling.style.display === 'none' ? '' : 'none'
    };
</script>
<header>Игра {{.Game}}.</header>
<pre>{{.TotalScore}}</pre>
<section>
	<header>Успешные сражения</header>
    {{ range $i, $battle := .Battles }}
		<article>
			<header style="cursor: pointer"
					onclick="hideSibling(this)"
			>{{$i}}. {{(index $battle.Players 0).Name }} ({{$battle.GameResult 0}})
				против {{(index $battle.Players 1).Name }} ({{$battle.GameResult 1}}). Время: {{$battle.Duration}}
			</header>
			<section style="display:none;">
				<section>
					<header style="cursor: pointer"
							onclick="hideSibling(this)">Логи первого бота
					</header>
					<pre>{{ $battle.Logs 0 }}</pre>
				</section>
				<section>
					<header style="cursor: pointer"
							onclick="hideSibling(this)">Логи второго бота
					</header>
					<pre>{{ $battle.Logs 1 }}</pre>
				</section>
			</section>
		</article>
    {{ end }}
</section>
<section>
	<header>Проблемные сражения</header>
    {{ range $i, $battle := .FailedBattles }}
		<article>
			<header style="cursor: pointer"
					onclick="hideSibling(this)"
			>{{$i}}. {{(index $battle.Players 0).Name }}
				против {{(index $battle.Players 1).Name }}: {{$battle.Err}}
			</header>
			<section style="display:none;">
				<section>
					<header style="cursor: pointer"
							onclick="hideSibling(this)">Логи первого бота
					</header>
					<pre>{{ $battle.Logs 0 }}</pre>
				</section>
				<section>
					<header style="cursor: pointer"
							onclick="hideSibling(this)">Логи второго бота
					</header>
					<pre>{{ $battle.Logs 1 }}</pre>
				</section>
			</section>
		</article>
    {{ end }}
</section>
</body>
</html>`

type TournamentReport struct {
	Game          string
	TotalScore    ScoreTable
	Battles       []*games.Battle
	FailedBattles []*games.Battle
}

func runTournament(bots []games.Bot, botsDesc []BotDescription, streams cli.Streams, battleTimeout time.Duration, concurrency, rounds int) TournamentReport {
	battles := make(chan *games.Battle)
	if rounds <= 0 {
		rounds = 1
	}
	go func() {
		for i, first := range bots {
			for _, second := range bots[i+1:] {
				for j := 0; j < rounds; j++ {
					battles <- &games.Battle{Players: [2]games.Bot{first, second}}
					battles <- &games.Battle{Players: [2]games.Bot{second, first}}
				}
			}
		}
		close(battles)
	}()

	scores := MakeScoreTable(botsDesc)
	output := log.New(streams.Out, "", 0)
	report := TournamentReport{TotalScore: scores}
	for battle := range RunRunners(concurrency, battles, battleTimeout) {
		if err := scores.Update(battle); err != nil {
			report.FailedBattles = append(report.FailedBattles, battle)
			output.Printf(
				"Неудачное сражение между %s vs %s: %s",
				battle.Players[0].Name, battle.Players[1].Name, err)
			continue
		}
		report.Battles = append(report.Battles, battle)
		output.Println(scores)
		output.Println(battle.State(),
			battle.Players[0].Name, battle.GameResult(0),
			battle.Players[1].Name, battle.GameResult(1))
		output.Println()
	}
	return report
}

func testBots(bots []games.Bot, streams cli.Streams, timeout time.Duration, concurrency int) error {
	battles := make(chan *games.Battle)
	go func() {
		for _, bot := range bots {
			battles <- &games.Battle{Players: [2]games.Bot{bot, bot}}
		}
		close(battles)
	}()

	output := log.New(streams.Out, "", 0)
	var problems error
	for battle := range RunRunners(concurrency, battles, timeout) {
		output.Println("Результат тестирования бота", battle.Players[0].Name)
		if err := analizeTestBattle(output, streams, battle); err != nil {
			problems = err
		}
		output.Println()
	}
	return problems
}

func MakeBots(botsDesc []BotDescription, dirName string) ([]games.Bot, error) {
	bots := make([]games.Bot, len(botsDesc))
	for i := range botsDesc {
		var err error
		bots[i], err = games.NewBot(path.Join(dirName, botsDesc[i].Cmd), botsDesc[i].Author)
		if err != nil {
			return nil, err
		}
	}
	return bots, nil
}

type BotsConfig map[string]BotDescription

func (bots BotsConfig) Slice() []BotDescription {
	botsDesc := make([]BotDescription, 0, len(bots))
	for author, bot := range bots {
		bot.Author = author
		botsDesc = append(botsDesc, bot)
	}
	return botsDesc
}

type TournamentConfigs struct {
	Bots    BotsConfig
	Timeout time.Duration
	Game    string
	Rounds  int
}

func (t *Tournament) readBotDescriptions(dir fs.FS) ([]BotDescription, error) {
	if executil.CheckFileFs(dir, t.Config) == nil {
		config, err := dir.Open(t.Config)
		if err != nil {
			return nil, err
		}
		if err := yaml.NewDecoder(config).Decode(&t.config); err != nil {
			return nil, err
		}
		return t.config.Bots.Slice(), nil
	}
	var botsDesc []BotDescription
	files, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if executil.Executable(file) {
			log.Println("Detect bot", file.Name())
			botsDesc = append(botsDesc, BotDescription{
				Author: file.Name(),
				Cmd:    file.Name(),
			})
		}
	}
	return botsDesc, nil
}

func RunRunners(concurrency int, battles chan *games.Battle, battleTimeout time.Duration) chan *games.Battle {
	if concurrency == 0 {
		if concurrency = runtime.NumCPU(); concurrency > 1 {
			concurrency--
		}
	}
	runnersGroup := new(sync.WaitGroup)
	runnersGroup.Add(concurrency)
	finished := make(chan *games.Battle)
	for i := 0; i < concurrency; i++ {
		go games.RunnerWithTimeout(battleTimeout, runnersGroup, battles, finished)
	}
	go func() {
		runnersGroup.Wait()
		close(finished)
	}()
	return finished
}

func (t Tournament) Usage() string {
	return "[folder/with/bots]"
}

type BotDescription struct {
	Author string `yaml:"author"`
	Cmd    string `yaml:"cmd"`
}
