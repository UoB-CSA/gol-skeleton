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
	alive := readAliveCounts(t, p.ImageWidth, p.ImageHeight)
	events := make(chan gol.Event)
	keyPresses := make(chan rune, 2)
	go gol.Run(p, events, keyPresses)
	aliveCountChan := make(chan gol.AliveCellsCount)

	go func() {
		// Forward AliveCellsCount events
		for event := range events {
			switch e := event.(type) {
			case gol.AliveCellsCount:
				aliveCountChan <- e
			}
		}
		close(aliveCountChan)
	}()

	// The first `AliveCellsCount` event must be sent within 4 seconds
	initialTimer := time.After(4 * time.Second)
	// All remaining `AliveCellsCount` events must be sent within 14 seconds
	remainingTimer := time.After(14 * time.Second)
	// Check if reported completed turns are incremental
	const unassignedCompletedTurns int = -1
	lastCompletedTurns := unassignedCompletedTurns
	receivedCount := 0
	for {
		select {
		case <-initialTimer:
			if lastCompletedTurns == unassignedCompletedTurns {
				t.Fatalf("%v No AliveCellsCount events received in 4 seconds", util.Red("ERROR"))
			}
		case <-remainingTimer:
			t.Fatalf(
				`%v Not enough AliveCellsCount events received in 14 seconds
					Make sure AliveCellsCount events are sent every 2 seconds`,
				util.Red("ERROR"),
			)
		case event, ok := <-aliveCountChan:
			var expected int
			if !ok {
				t.Fatalf(
					"%v Not enough AliveCellsCount events received, as distributor exited too early",
					util.Red("ERROR"),
				)
			} else if event.CompletedTurns <= lastCompletedTurns {
				t.Fatalf(
					`%v Reported completed turns should be incremental
					Reported completed turn: %v
					Last completed turn: %v`,
					util.Red("ERROR"),
					event.CompletedTurns,
					lastCompletedTurns,
				)
			} else if event.CompletedTurns <= 10000 {
				expected = alive[event.CompletedTurns]
			} else if event.CompletedTurns%2 == 0 {
				expected = 5565
			} else {
				expected = 5567
			}
			actual := event.CellsCount
			if expected != actual {
				t.Fatalf(
					"%v At turn %v expected %v alive cells, got %v instead",
					util.Red("ERROR"),
					event.CompletedTurns,
					expected,
					actual,
				)
			} else {
				t.Log(event)
				lastCompletedTurns = event.CompletedTurns
				receivedCount++
				if receivedCount >= 5 {
					keyPresses <- 'q'
					return
				}
			}
		}
	}
}

func readAliveCounts(t *testing.T, width, height int) map[int]int {
	f, err := os.Open("check/alive/" + fmt.Sprintf("%vx%v.csv", width, height))
	if err != nil {
		t.Fatalf("%v %v", util.Red("ERROR"), err)
	}
	reader := csv.NewReader(f)
	table, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("%v %v", util.Red("ERROR"), err)
	}
	alive := make(map[int]int)
	for i, row := range table {
		if i == 0 {
			continue
		}
		completedTurns, err := strconv.Atoi(row[0])
		if err != nil {
			t.Fatalf("%v %v", util.Red("ERROR"), err)
		}
		aliveCount, err := strconv.Atoi(row[1])
		if err != nil {
			t.Fatalf("%v %v", util.Red("ERROR"), err)
		}
		alive[completedTurns] = aliveCount
	}
	return alive
}
