package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	rounds := flag.Int("r", 1, "Количество раундов")
	bot1Output := flag.String("o", "-", "stderr первого бота ('-' = stdout, '+' - stderr, '<filename>' - будет сохранено в файл, пусто - /dev/null)")
	flag.Parse()
	if flag.NArg() != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "Usage:\n  botctl [flags] 'path/to/mybot1.exe [addition args]' 'path/to/mybot2.exe [addition args]'")
		flag.PrintDefaults()
		return
	}
	bot1, err := NewBot(1, flag.Arg(0))
	logAndExitOnError(err)
	bot2, err := NewBot(2, flag.Arg(1))
	logAndExitOnError(err)
	settings := TournamentSettings{
		Bot1:          bot1,
		Bot2:          bot2,
		TotalWriter:   os.Stdout,
		ProcessWriter: os.Stderr,
		Bot1Writer:    nil,
		Bot2Writer:    nil,
	}
	switch *bot1Output {
	case "":
		break
	case "-":
		settings.Bot1Writer = os.Stdout
	case "+":
		settings.Bot1Writer = os.Stderr
	default:
		file, err := os.Create(*bot1Output)
		logAndExitOnError(err)
		defer file.Close()
		settings.Bot1Writer = file
	}
	err = settings.Init().Run(*rounds)
	if err != nil {
		log.Println("Что-то пошло не так:", err)
	}
}

func logAndExitOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
