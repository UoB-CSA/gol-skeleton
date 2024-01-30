package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
	"uk.ac.bris.cs/gameoflife/util"
)

var w *sdl.Window
var refresh chan bool

func TestMain(m *testing.M) {
	runtime.LockOSThread()
	var sdlFlag = flag.Bool(
		"sdl",
		false,
		"Enable the SDL window for testing.")

	flag.Parse()
	done := make(chan int, 1)
	test := func() { done <- m.Run() }
	if !(*sdlFlag) {
		go test()
	} else {
		w = sdl.NewWindow(512, 512)
		refresh = make(chan bool, 1)
		fps := 60
		ticker := time.NewTicker(time.Second / time.Duration(fps))
		dirty := false
		go test()
	loop:
		for {
			select {
			case code := <-done:
				done <- code
				w.Destroy()
				break loop
			case <-ticker.C:
				w.PollEvent()
				if dirty {
					w.RenderFrame()
					dirty = false
				}
			case <-refresh:
				dirty = true
			}
		}
	}
	os.Exit(<-done)
}

// TestSdl tests key presses and events
func TestSdl(t *testing.T) {
	params := gol.Params{
		Turns:       100000000,
		Threads:     8,
		ImageWidth:  512,
		ImageHeight: 512,
	}

	keyPresses := make(chan rune, 10)
	events := make(chan gol.Event, 1000)

	golDone := make(chan bool, 1)
	go func() {
		gol.Run(params, events, keyPresses)
		golDone <- true
	}()
	startTester(t, params, keyPresses, events, golDone)
}

type Tester struct {
	t            *testing.T
	params       gol.Params
	keyPresses   chan<- rune
	events       <-chan gol.Event
	eventWatcher <-chan gol.Event
	turn         int
	world        [][]byte
	aliveMap     map[int]int
}

func startTester(
	t *testing.T,
	params gol.Params,
	keyPresses chan<- rune,
	events <-chan gol.Event,
	golDone <-chan bool,
) {
	world := make([][]byte, params.ImageHeight)
	for i := range world {
		world[i] = make([]byte, params.ImageWidth)
	}

	eventWatcher := make(chan gol.Event, 1000)
	tester := Tester{
		t:            t,
		params:       params,
		keyPresses:   keyPresses,
		events:       events,
		eventWatcher: eventWatcher,
		turn:         0,
		world:        world,
		aliveMap:     readAliveCounts(params.ImageWidth, params.ImageHeight),
	}

	cancelDeadline := deadline(25*time.Second,
		"Your program should complete this test within 20 seconds. Is your program deadlocked?")

	go tester.testPause(3 * time.Second)
	go tester.testOutput(12 * time.Second)
	quitting := make(chan bool, 1)
	go func() {
		tester.testQuitting(16 * time.Second)
		quitting <- true
	}()

	cellFlippedReceived := false
	turnCompleteReceived := false

	avgTurns := util.NewAvgTurns()

loop:
	for {
		select {
		case <-quitting:
			if !cellFlippedReceived {
				panic("No CellFlipped events received")
			}
			if !turnCompleteReceived {
				panic("No TurnComplete events received")
			}
			<-golDone
			cancelDeadline <- true
			break loop
		case event := <-tester.events:
			switch e := event.(type) {
			case gol.CellFlipped:
				cellFlippedReceived = true
				assert(e.CompletedTurns == tester.turn || e.CompletedTurns == tester.turn+1,
					"Expected completed %v turns, got %v instead", tester.turn, e.CompletedTurns)
				tester.world[e.Cell.Y][e.Cell.X] = ^tester.world[e.Cell.Y][e.Cell.X]
				if w != nil {
					w.FlipPixel(e.Cell.X, e.Cell.Y)
				}
			case gol.CellsFlipped:
				cellFlippedReceived = true
				assert(e.CompletedTurns == tester.turn || e.CompletedTurns == tester.turn+1,
					"Expected completed %v turns, got %v instead", tester.turn, e.CompletedTurns)
				for _, cell := range e.Cells {
					tester.world[cell.Y][cell.X] = ^tester.world[cell.Y][cell.X]
					if w != nil {
						w.FlipPixel(cell.X, cell.Y)
					}
				}
			case gol.TurnComplete:
				turnCompleteReceived = true
				tester.turn++
				assert(e.CompletedTurns == tester.turn || e.CompletedTurns == tester.turn+1,
					"Expected completed %v turns, got %v instead", tester.turn, e.CompletedTurns)
				tester.testAlive()
				tester.testGol()
				if refresh != nil {
					refresh <- true
				}
			case gol.AliveCellsCount:
				fmt.Printf("Completed Turns %-8v %-20v Avg%+5v turns/sec\n", event.GetCompletedTurns(), event, avgTurns.Get(event.GetCompletedTurns()))
			case gol.ImageOutputComplete:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				eventWatcher <- e
			case gol.FinalTurnComplete:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				eventWatcher <- e
			case gol.StateChange:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				eventWatcher <- e
			}
		}
	}
}

func (tester *Tester) testAlive() {
	aliveCount := 0
	for _, row := range tester.world {
		for _, cell := range row {
			if cell == 0xFF {
				aliveCount++
			}
		}
	}
	expected := 0
	if tester.turn <= 10000 {
		expected = tester.aliveMap[tester.turn]
	} else if tester.turn%2 == 0 {
		expected = 5565
	} else {
		expected = 5567
	}
	assert(aliveCount == expected,
		"At turn %v expected %v alive cells, got %v instead", tester.turn, expected, aliveCount)
}

func (tester *Tester) testGol() {
	if tester.turn == 0 || tester.turn == 1 || tester.turn == 100 {
		width, height := tester.params.ImageWidth, tester.params.ImageHeight
		path := fmt.Sprintf("check/images/%vx%vx%v.pgm", width, height, tester.turn)
		expectedAlive := readAliveCells(path, width, height)
		aliveCells := make([]util.Cell, 0, width*height)
		for y := range tester.world {
			for x, cell := range tester.world[y] {
				if cell == 255 {
					aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
				}
			}
		}
		assertEqualBoard(tester.t, aliveCells, expectedAlive, tester.params)
	}
}

func (tester *Tester) testOutput(delay time.Duration) {
	time.Sleep(delay)
	width, height := tester.params.ImageWidth, tester.params.ImageHeight
	tester.t.Logf("Testing image output")
	for len(tester.eventWatcher) > 0 {
		<-tester.eventWatcher
	}
	tester.keyPresses <- 's'
	timeout(4*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.ImageOutputComplete); ok {
				assert(e.Filename == fmt.Sprintf("%vx%vx%v", width, height, e.CompletedTurns),
					"Filename is not correct")
				break
			}
		}
	}, "No ImageOutput events received in 4 seconds\n%v",
		"If this test is running in WSL2, please make sure the test is located within WSL2 file system rather than Windows! i.e. Your path must not start with /mnt/...")
}

func (tester *Tester) testPause(delay time.Duration) {
	time.Sleep(delay)
	tester.t.Logf("Testing Pause key pressed")
	for len(tester.eventWatcher) > 0 {
		<-tester.eventWatcher
	}
	tester.keyPresses <- 'p'
	timeout(2*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.StateChange); ok && e.NewState == gol.Paused {
				break
			}
		}
	}, "No Pause events received in 2 seconds")

	tester.testOutput(2 * time.Second)

	time.Sleep(2 * time.Second)
	tester.t.Logf("Testing Pause key pressed again")
	tester.keyPresses <- 'p'
	timeout(2*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.StateChange); ok && e.NewState == gol.Executing {
				break
			}
		}
	}, "No Executing events received in 2 seconds")
}

func (tester *Tester) testQuitting(delay time.Duration) {
	time.Sleep(delay)
	tester.t.Logf("Testing Quit key pressed")
	for len(tester.eventWatcher) > 0 {
		<-tester.eventWatcher
	}
	tester.keyPresses <- 'q'
	timeout(2*time.Second, func() {
		for e := range tester.eventWatcher {
			if _, ok := e.(gol.FinalTurnComplete); ok {
				break
			}
		}
	}, "No FinalTurnComplete events received in 2 seconds")

	timeout(4*time.Second, func() {
		for e := range tester.eventWatcher {
			if _, ok := e.(gol.ImageOutputComplete); ok {
				break
			}
		}
	}, "No ImageOutput events received in 4 seconds")

	timeout(2*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.StateChange); ok && e.NewState == gol.Quitting {
				break
			}
		}
	}, "No Quitting events received in 2 seconds")
}

func deadline(ddl time.Duration, msg string) chan<- bool {
	done := make(chan bool, 1)
	go func() {
		select {
		case <-time.After(ddl):
			panic(msg)
		case <-done:
			return
		}
	}()
	return done
}

func timeout(ddl time.Duration, f func(), msg string, a ...interface{}) {
	done := make(chan bool, 1)
	go func() {
		f()
		done <- true
	}()
	select {
	case <-time.After(ddl):
		panic(fmt.Sprintf(msg, a...))
	case <-done:
		break
	}
}

func assert(predicate bool, msg string, a ...interface{}) {
	if !predicate {
		panic(fmt.Sprintf(msg, a...))
	}
}
