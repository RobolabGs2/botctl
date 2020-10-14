package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
)

type GameResult int

const (
	Win  GameResult = 1
	Lose GameResult = -1
	Draw GameResult = 0
)

func (order TurnOrder) CmdArg() string {
	switch order {
	case First:
		return "0"
	case Second:
		return "1"
	default:
		panic(fmt.Sprintf("Неизвестая очередь хода %#v", order))
	}
}

func (res GameResult) String() string {
	switch res {
	case Win:
		return "Победа"
	case Lose:
		return "Поражение"
	case Draw:
		return "Ничья"
	default:
		return strconv.Itoa(int(res))
	}
}

func SummarizeGame(err error) (GameResult, error) {
	if err == nil {
		return Win, nil
	}
	if exit := new(exec.ExitError); errors.As(err, &exit) {
		switch exit.ExitCode() {
		case 3:
			return Lose, nil
		case 4:
			return Draw, nil
		}
	}
	return 0, err
}
