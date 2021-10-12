package commands

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/RobolabGs2/botctl/cli"
	"github.com/RobolabGs2/botctl/executil"
	"github.com/RobolabGs2/botctl/games"
	"gopkg.in/yaml.v3"
)

type Tournament struct {
	Concurrency int `name:"c" default:"0" desc:"Количество одновременных соревнований. '0' - количество виртуальных ядер минус 1"`
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
	if len(botsDesc) < 2 {
		return errors.New("not enough bots for tournament")
	}
	bots, err := MakeBots(botsDesc, dirName)
	if err != nil {
		return err
	}
	battles := make(chan *games.Battle)
	go func() {
		for i, first := range bots {
			for _, second := range bots[i+1:] {
				battles <- &games.Battle{Players: [2]games.Bot{first, second}}
				battles <- &games.Battle{Players: [2]games.Bot{second, first}}
			}
		}
		close(battles)
	}()

	scores := MakeScoreTable(botsDesc)
	output := log.New(streams.Out, "", 0)
	for battle := range RunRunners(t.Concurrency, battles) {
		scores.Update(battle)
		log.Println(battle.State(), battle.Players[0].Name, battle.GameResult(0), battle.Players[1].Name, battle.GameResult(1))
		output.Println(scores)
	}
	return nil
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

func (t *Tournament) readBotDescriptions(dir fs.FS) ([]BotDescription, error) {
	var botsDesc []BotDescription
	if executil.CheckFileFs(dir, "tournament.yaml") == nil {
		config, err := dir.Open("tournament.yaml")
		if err != nil {
			return nil, err
		}
		bots := map[string]BotDescription{}
		if err := yaml.NewDecoder(config).Decode(bots); err != nil {
			return nil, err
		}
		for author, bot := range bots {
			bot.Author = author
			botsDesc = append(botsDesc, bot)
		}
	} else {
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
	}
	return botsDesc, nil
}

func RunRunners(concurrency int, battles chan *games.Battle) chan *games.Battle {
	if concurrency == 0 {
		if concurrency = runtime.NumCPU(); concurrency > 1 {
			concurrency--
		}
	}
	runnersGroup := new(sync.WaitGroup)
	runnersGroup.Add(concurrency)
	finished := make(chan *games.Battle)
	for i := 0; i < concurrency; i++ {
		go games.Runner(runnersGroup, battles, finished)
	}
	go func() {
		runnersGroup.Wait()
		close(finished)
	}()
	return finished
}

func (t *Tournament) Usage() string {
	return "Турнир"
}

type BotDescription struct {
	Author string `yaml:"author"`
	Cmd    string `yaml:"cmd"`
}
