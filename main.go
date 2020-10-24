package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	rounds := flag.Int("r", 1, "Количество раундов")
	bot1Output := LogCollectorFlag("o1", "+", "первого")
	bot2Output := LogCollectorFlag("o2", "", "второго")
	flag.Parse()
	if flag.NArg() != 2 {
		_, _ = fmt.Fprintln(
			os.Stderr,
			`Usage:
botctl [flags] "path/to/mybot1.exe [addition args]" "path/to/mybot2.exe [addition args]"`)
		flag.PrintDefaults()
		return
	}
	bot1, err := NewBot(1, flag.Arg(0))
	logAndExitOnError(err)
	bot2, err := NewBot(2, flag.Arg(1))
	logAndExitOnError(err)
	settings := TournamentSettings{
		Bot1:                   bot1,
		Bot2:                   bot2,
		TotalWriter:            os.Stdout,
		ProcessWriter:          os.Stderr,
		Bot1LogCollectorFabric: GetLogCollectorFabric(*bot1Output, bot1),
		Bot2LogCollectorFabric: GetLogCollectorFabric(*bot2Output, bot2),
	}
	err = settings.Init().Run(*rounds)
	if err != nil {
		log.Println("Что-то пошло не так:", err)
	}
}

func LogCollectorFlag(name, value, numeric string) *string {
	return flag.String(name, value, fmt.Sprint(`Куда перенаправить stdout `, numeric, ` бота 
'-' = stdout, '+' - stderr, '<dirname>' - будет сохранено в папку, файл на раунд, пусто - игнорировать`))
}

func GetLogCollectorFabric(botOutputOption string, bot *Bot) LogCollectorFabric {
	switch botOutputOption {
	case "":
		return IgnoreLogs
	case "-":
		return LogsRedirectTotWriter(os.Stdout)
	case "+":
		return LogsRedirectTotWriter(os.Stderr)
	default:
		logAndExitOnError(os.MkdirAll(botOutputOption, 0666))
		return func(round int) io.Writer {
			file, err := os.Create(filepath.Join(botOutputOption, fmt.Sprintf("round_%02d_bot_%d_%s.txt", round, bot.Number, bot.Name)))
			if err != nil {
				panic(fmt.Errorf("%s: %w", bot.Name, err))
			}
			return file
		}
	}
}

func logAndExitOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
