package games

import "fmt"

type TurnOrder int

const (
	First  TurnOrder = 0
	Second TurnOrder = 1
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
	return 1 - order
}
