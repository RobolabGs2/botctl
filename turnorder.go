package main

import "fmt"

type TurnOrder bool

const (
	First  TurnOrder = true
	Second TurnOrder = false
)

func (order TurnOrder) String() string {
	switch order {
	case First:
		return "первым"
	case Second:
		return "вторым"
	default:
		panic(fmt.Sprintf("Неизвестая очередь хода %#v", order))
	}
}

func (order TurnOrder) Opponent() TurnOrder {
	return !order
}
