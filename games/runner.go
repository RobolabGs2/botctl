package games

import (
	"context"
	"sync"
)

func Runner(group *sync.WaitGroup, battles <-chan *Battle, finished chan<- *Battle) {
	defer group.Done()
	for battle := range battles {
		// battle contains error, this error should be handled by a receiver of finished chan
		_ = battle.Run(context.TODO())
		finished <- battle
	}
}
