package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func main() {
	rounds := flag.Int("r", 1, "Количество раундов")
	verbosity := flag.Int("v", 1, "0 - логи ботов не выводятся, 1 - выводятся логи первого, 2 - обоих ботов")
	flag.Parse()
	if flag.NArg() != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "Usage:\n  botctl [flags] path/to/mybot1.exe path/to/mybot2.exe")
		flag.PrintDefaults()
		return
	}
	bot1ExeName := filepath.Clean(flag.Arg(0))
	bot2ExeName := filepath.Clean(flag.Arg(1))
	checkFile(bot1ExeName)
	checkFile(bot2ExeName)
	totalScore1, totalScore2 := GameResult(0), GameResult(0)
	for i := 0; i < *rounds; i++ {
		bot1, bot2 := MakeCmd(bot1ExeName, i%2), MakeCmd(bot2ExeName, 1-i%2)
		bot1.Stdout, _ = bot2.StdinPipe()
		bot2.Stdout, _ = bot1.StdinPipe()
		switch *verbosity {
		case 2:
			bot2.Stderr = os.Stderr
			fallthrough
		case 1:
			bot1.Stderr = os.Stdout
		}
		log.Println("Раунд:", i)
		logAndExitOnError(bot1.Start(), bot1ExeName)
		logAndExitOnErrorAndAction(bot2.Start(), bot2ExeName, bot1.Process.Kill)
		score1, err := SummarizeGame(bot1.Wait())
		if err != nil {
			log.Println(bot1ExeName, ": ", err)
		}
		score2, err := SummarizeGame(bot2.Wait())
		if err != nil {
			log.Println(bot2ExeName, ": ", err)
		}
		if score1+score2 != Draw {
			log.Println("Кто-то из ботов мухлюет")
		}
		totalScore1 += score1
		totalScore2 += score2
		log.Println(score1, ":", score2)
	}
	if *rounds > 1 {
		log.Println("Итого:")
		log.Println(totalScore1, ":", totalScore2)
	}
}

func checkFile(filename string) {
	if _, err := os.Stat(filename); err != nil {
		log.Fatalln("Проблемы с ", filename, ":", err)
	}
}

func logAndExitOnError(err error, botExeName string) {
	if err != nil {
		log.Fatalln(botExeName, err)
	}
}
func logAndExitOnErrorAndAction(err error, botExeName string, action func() error) {
	if err != nil {
		_ = action()
		log.Fatalln(botExeName, err)
	}
}

type GameResult int

const (
	Win  GameResult = 1
	Lose GameResult = -1
	Draw GameResult = 0
)

func SummarizeGame(err error) (GameResult, error) {
	if err == nil {
		return Win, nil
	}
	if exit := new(exec.ExitError); errors.As(err, &exit) {
		switch exit.ExitCode() {
		case 1:
			return Lose, nil
		case 2:
			return Draw, nil
		}
	}
	return 0, err
}

func MakeCmd(botExeName string, order int) *exec.Cmd {
	// Платформозависимо - запускаем с помощью cmd
	return exec.Command("cmd", "/C", botExeName, strconv.Itoa(order))
}
