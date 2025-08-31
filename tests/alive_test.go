package tests

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

// TestAlive will automatically check the 512x512 cell counts for the first 5 messages.
// You can manually check your counts by looking at CSVs provided in check/alive
func TestAlive(t *testing.T) {
	p := gol.Params{
		Turns:       100000000,
		Threads:     8,
		ImageWidth:  512,
		ImageHeight: 512,
	}
	alive := readAliveCounts(p.ImageWidth, p.ImageHeight)
	events := make(chan gol.Event)
	keyPresses := make(chan rune, 2)
	go gol.Run(p, events, keyPresses)

	implemented := false
	eventsClosed := make(chan bool)
	aliveCellCounts := make(chan gol.AliveCellsCount)

	go func() {
		for event := range events {
			switch e := event.(type) {
			case gol.AliveCellsCount:
				aliveCellCounts <- e
			}
		}
		eventsClosed <- true
	}()

	timer := time.After(5 * time.Second)

	i := 0
	for {
		select {
		case e := <-aliveCellCounts:
			var expected int
			if e.CompletedTurns == 0 {
				t.Error("ERROR: Count reported for turn 0, should have a delay.")
			} else if e.CompletedTurns <= 10000 {
				expected = alive[e.CompletedTurns]
			} else if e.CompletedTurns%2 == 0 {
				expected = 5565
			} else {
				expected = 5567
			}
			actual := e.CellsCount
			if expected != actual {
				t.Fatalf("ERROR: At turn %v expected %v alive cells, got %v instead", e.CompletedTurns, expected, actual)
			} else {
				t.Log(e)
				implemented = true
				i++

				if i >= 5 {
					keyPresses <- 'q'
					return
				}
			}
		case <-timer:
			if !implemented {
				t.Fatal("ERROR: No AliveCellsCount events received in 5 seconds")
			}
		case <-eventsClosed:
			t.Fatal("ERROR: Not enough AliveCellsCount events received")
		}
	}
}

func readAliveCounts(width, height int) map[int]int {
	f, err := os.Open("check/alive/" + fmt.Sprintf("%vx%v.csv", width, height))
	util.Check(err)
	reader := csv.NewReader(f)
	table, err := reader.ReadAll()
	util.Check(err)
	alive := make(map[int]int)
	for i, row := range table {
		if i == 0 {
			continue
		}
		completedTurns, err := strconv.Atoi(row[0])
		util.Check(err)
		aliveCount, err := strconv.Atoi(row[1])
		util.Check(err)
		alive[completedTurns] = aliveCount
	}
	return alive
}
