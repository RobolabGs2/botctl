package commands

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/RobolabGs2/botctl/games"
)

type ScoreTable []ScoreLine

type ScoreLine struct {
	Author string
	Score  float64
	Games  int
}

func MakeScoreTable(players []BotDescription) ScoreTable {
	scores := make(ScoreTable, len(players))
	for i, bot := range players {
		scores[i].Author = bot.Author
	}
	return scores
}

func (t ScoreTable) Update(battle *games.Battle) error {
	if state := battle.State(); state != games.BattleFinished {
		return fmt.Errorf("battle was not finished (%s): %w", state, battle.Err())
	}
	t.updatePlayer(battle.Players[0].Name, battle.GameResult(0).Score())
	t.updatePlayer(battle.Players[1].Name, battle.GameResult(1).Score())
	return nil
}

func (t ScoreTable) String() string {
	buf := bytes.Buffer{}
	if _, err := t.WriteTo(&buf); err != nil {
		panic(fmt.Errorf("can't stringlify score table: %w", err))
	}
	return buf.String()
}

func (t ScoreTable) WriteTo(writer io.Writer) (int64, error) {
	summary := tabwriter.NewWriter(writer, 0, 4, 4, ' ', 0)
	n, err := fmt.Fprint(summary, "Место\tАвтор\tСчёт\tИгр всего\n")
	if err != nil {
		return int64(n), err
	}
	total := int64(n)
	for i, line := range t {
		n, err := fmt.Fprintf(summary, "%d\t%s\t%.1f\t%d\n", i+1, line.Author, line.Score, line.Games)
		if err != nil {
			return int64(n), err
		}
		total += int64(n)
	}
	return total, summary.Flush()
}

func (t ScoreTable) updatePlayer(player string, score float64) {
	for i := range t {
		if t[i].Author == player {
			t[i].Games++
			t[i].Score += score
			t.fixFrom(i)
			return
		}
	}
	panic("unreachable")
}

// fixFrom всплывает пузырёк с игроком, увеличившим счёт
func (t ScoreTable) fixFrom(from int) {
	for i := from; i > 0; i-- {
		if t[i-1].Score >= t[i].Score {
			break
		}
		t[i], t[i-1] = t[i-1], t[i]
	}
}
