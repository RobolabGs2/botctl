package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	if (len(os.Args)) != 3 {
		log.Fatalf("bot-connector [bot1].exe [bot2].exe")
	}
	bot1ExeName := os.Args[1]
	bot2ExeName := os.Args[2]
	bot1, bot2 := MakeCmd(bot1ExeName, "0"), MakeCmd(bot2ExeName, "1")
	bot1.Stdout, _ = bot2.StdinPipe()
	bot2.Stdout, _ = bot1.StdinPipe()
	bot1.Stderr = os.Stdout
	logAndExitOnError(bot1.Start(), bot1ExeName)
	logAndExitOnError(bot2.Start(), bot2ExeName)
	logAndExitOnError(bot1.Wait(), bot1ExeName)
	logAndExitOnError(bot2.Wait(), bot2ExeName)
}

func logAndExitOnError(err error, bot1ExeName string) {
	if err != nil {
		log.Fatalln(bot1ExeName, err)
	}
}

func MakeCmd(botExeName string, order string) *exec.Cmd {
	// Платформозависимо - запускаем с помощью cmd
	return exec.Command("cmd", "/C", botExeName, order)
}
