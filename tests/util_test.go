package tests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

func checkEqualBoard(given, expected []util.Cell) bool {
	givenLen := len(given)
	expectedLen := len(expected)

	if givenLen != expectedLen {
		return false
	}

	visited := make([]bool, expectedLen)
	for i := 0; i < givenLen; i++ {
		element := given[i]
		found := false
		for j := 0; j < expectedLen; j++ {
			if visited[j] {
				continue
			}
			if expected[j] == element {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func boardFail(t *testing.T, given, expected []util.Cell, p gol.Params) bool {
	errorString := fmt.Sprintf("-----------------\n\n  FAILED TEST\n  %vx%v\n  %d Workers\n  %d Turns\n", p.ImageWidth, p.ImageHeight, p.Threads, p.Turns)
	if p.ImageWidth == 16 && p.ImageHeight == 16 {
		errorString = errorString + util.AliveCellsToString(given, expected, p.ImageWidth, p.ImageHeight)
	}
	t.Error(errorString)
	return false
}

func assertEqualBoard(t *testing.T, given, expected []util.Cell, p gol.Params) bool {
	equal := checkEqualBoard(given, expected)

	if !equal {
		boardFail(t, given, expected, p)
	}

	return equal
}

func emptyOutFolder() {
	os.RemoveAll("out")
	_ = os.Mkdir("out", os.ModePerm)
}

func readAliveCells(path string, width, height int) []util.Cell {
	data, ioError := os.ReadFile(path)
	util.Check(ioError)

	fields := strings.Fields(string(data))

	if fields[0] != "P5" {
		panic("Not a pgm file")
	}

	imageWidth, _ := strconv.Atoi(fields[1])
	if imageWidth != width {
		panic("Incorrect width")
	}

	imageHeight, _ := strconv.Atoi(fields[2])
	if imageHeight != height {
		panic("Incorrect height")
	}

	maxval, _ := strconv.Atoi(fields[3])
	if maxval != 255 {
		panic("Incorrect maxval/bit depth")
	}

	image := []byte(fields[4])

	var cells []util.Cell
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := image[0]
			if cell != 0 {
				cells = append(cells, util.Cell{
					X: x,
					Y: y,
				})
			}
			image = image[1:]
		}
	}
	return cells
}

type Tester struct {
	t            *testing.T
	params       gol.Params
	keyPresses   chan<- rune
	events       <-chan gol.Event
	eventWatcher chan gol.Event
	quitting     chan bool
	golDone      <-chan bool
	turn         int
	world        [][]byte
	aliveMap     map[int]int
	testTurn     bool
	sdlSync      chan bool
}

func MakeTester(
	t *testing.T,
	params gol.Params,
	keyPresses chan<- rune,
	events <-chan gol.Event,
	golDone <-chan bool,
) Tester {
	world := make([][]byte, params.ImageHeight)
	for i := range world {
		world[i] = make([]byte, params.ImageWidth)
	}

	clearPixels()
	emptyOutFolder()

	eventWatcher := make(chan gol.Event, 1000)
	return Tester{
		t:            t,
		params:       params,
		keyPresses:   keyPresses,
		events:       events,
		eventWatcher: eventWatcher,
		quitting:     make(chan bool),
		golDone:      golDone,
		turn:         0,
		world:        world,
		aliveMap:     readAliveCounts(params.ImageWidth, params.ImageHeight),
		testTurn:     false,
		sdlSync:      nil,
	}
}

func (tester *Tester) SetTestTurn() {
	tester.testTurn = true
}

func (tester *Tester) SetTestSdl() {
	tester.testTurn = true
	tester.sdlSync = make(chan bool)
}

func (tester *Tester) Loop() {
	defer clearPixels()
	limitedAssert := LimitedAssert{t: tester.t, failed: false, limitHit: false}

	avgTurns := util.NewAvgTurns()

	for {
		select {
		case quitPanic := <-tester.quitting:
			awaitDone := func() {
				for {
					select {
					case <-tester.golDone:
						return
					case <-tester.events:
					}
				}

			}

			if quitPanic {
				timeout(tester.t, 2*time.Second, awaitDone, "Your program has not returned from the gol.Run function\n%v\n%v", "Continuing with other tests, leaving your program executing", "You may get unexpected behaviour")
			} else {
				timeoutWarn(2*time.Second, awaitDone, "Your program has not returned from the gol.Run function\n%v\n%v", "Continuing with other tests, leaving your program executing", "You may get unexpected behaviour")
			}

			limitedAssert.LimitHitMessage("Repeat CellFlipped errors have been hidden")

			return
		case event := <-tester.events:
			switch e := event.(type) {
			case gol.CellFlipped:
				if tester.testTurn {
					limitedAssert.Assert(e.CompletedTurns == tester.turn || e.CompletedTurns == tester.turn+1,
						"Expected completed %v or %v turns for CellFlipped event, got %v instead", tester.turn, tester.turn+1, e.CompletedTurns)
				}
				tester.world[e.Cell.Y][e.Cell.X] = ^tester.world[e.Cell.Y][e.Cell.X]
				flipCell(e.Cell)
			case gol.CellsFlipped:
				if tester.testTurn {
					limitedAssert.Assert(e.CompletedTurns == tester.turn || e.CompletedTurns == tester.turn+1,
						"Expected completed %v or %v turns for CellsFlipped event, got %v instead", tester.turn, tester.turn+1, e.CompletedTurns)
				}
				for _, cell := range e.Cells {
					tester.world[cell.Y][cell.X] = ^tester.world[cell.Y][cell.X]
					flipCell(cell)
				}
			case gol.TurnComplete:
				if tester.testTurn {
					limitedAssert.Reset()
					assert(tester.t, e.CompletedTurns == tester.turn || e.CompletedTurns == tester.turn+1,
						"Expected completed %v or %v turns for TurnComplete event, got %v instead", tester.turn, tester.turn+1, e.CompletedTurns)
				}
				tester.turn++
				refresh()
				if tester.sdlSync != nil {
					tester.sdlSync <- true
					<-tester.sdlSync
				}
			case gol.AliveCellsCount:
				fmt.Printf("Completed Turns %-8v %-20v Avg%+5v turns/sec\n", event.GetCompletedTurns(), event, avgTurns.Get(event.GetCompletedTurns()))
			case gol.ImageOutputComplete:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				tester.HandleEvent(e)
			case gol.FinalTurnComplete:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				tester.HandleEvent(e)
			case gol.StateChange:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				tester.HandleEvent(e)

				if tester.sdlSync != nil && tester.turn == 0 {
					tester.sdlSync <- true
					<-tester.sdlSync
				}
			}
		}
	}
}

func (tester *Tester) HandleEvent(event gol.Event) {
	if len(tester.eventWatcher) >= cap(tester.eventWatcher) {
		fmt.Printf("WARNING: The tester's internal event buffer is full\n%v\n%v\n",
			"Discarding earliest event",
			"Are you sending too many ImageOutputComplete, FinalTurnComplete or StateChange events?")
		<-tester.eventWatcher
	}

	tester.eventWatcher <- event
}

func (tester *Tester) Stop(returnPanic bool) {
	stop := make(chan bool)

	go func() {
		for {
			select {
			case <-tester.sdlSync:
				tester.sdlSync <- true
			case <-stop:
				return
			}
		}
	}()

	tester.quitting <- returnPanic
	stop <- true
}

func (tester *Tester) AwaitSync() (int, bool) {
	success := timeout(tester.t, 2*time.Second, func() {
		<-tester.sdlSync
	}, "No turns completed in 2 seconds. Is your program deadlocked?")
	return tester.turn, success
}

func (tester *Tester) Continue() {
	tester.sdlSync <- true
}

func (tester *Tester) TestAlive() {
	tester.t.Logf("Checking number of alive cells in the SDL window at turn %v", tester.turn)

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
	assert(tester.t, aliveCount == expected,
		"At turn %v expected %v alive cells in the SDL window, got %v instead", tester.turn, expected, aliveCount)
}

func (tester *Tester) TestImage() {
	if tester.turn == 0 || tester.turn == 1 || tester.turn == 100 {
		tester.t.Logf("Checking SDL image at turn %v", tester.turn)

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

		equal := checkEqualBoard(aliveCells, expectedAlive)
		if !equal {
			if tester.turn == 0 {
				tester.t.Error("ERROR: The image displayed in the SDL window is incorrect for turn 0\nHave you sent the correct CellFlipped events before StateChange Executing?")
			} else {
				tester.t.Errorf("ERROR: The image displayed in the SDL window is incorrect for turn %v", tester.turn)
			}
		}
	} else {
		fmt.Printf("WARNING: TestImage called on invalid turn: %v. This call will be ignored\n", tester.turn)
	}
}

func (tester *Tester) TestStartsExecuting() {
	tester.t.Logf("Testing for first StateChange Executing event")
	timeout(tester.t, 2*time.Second, func() {
		e := <-tester.eventWatcher
		if e, ok := e.(gol.StateChange); ok {
			assert(tester.t, e.NewState == gol.Executing,
				"First StateChange event should have a NewState of Executing, not %v", e)
			assert(tester.t, e.CompletedTurns == 0,
				"First StateChange event should have a CompletedTurns of 0, not %v", e.CompletedTurns)
			return
		}

		tester.t.Errorf("ERROR: %v event should not be sent before StateChange Executing", e)

	}, "No StateChange events received in 2 seconds")
}

func (tester *Tester) TestExecutes(turn int) {
	tester.t.Logf("Testing for StateChange Executing event")
	timeout(tester.t, 2*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.StateChange); ok && e.NewState == gol.Executing {
				if e.CompletedTurns != turn && e.CompletedTurns != turn+1 {
					tester.t.Errorf("ERROR: StateChange event should have a CompletedTurns of %v or %v, not %v", turn, turn+1, e.CompletedTurns)
				}

				return
			}
		}
	}, "No StateChange Executing events received in 2 seconds")
}

func (tester *Tester) TestPauses() int {
	tester.t.Logf("Testing for StateChange Paused event")

	turn := make(chan int, 1)

	completed := timeout(tester.t, 2*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.StateChange); ok && e.NewState == gol.Paused {
				turn <- e.CompletedTurns
				return
			}
		}
	}, "No StateChange Paused events received in 2 seconds")

	if !completed {
		return -1
	} else {
		return <-turn
	}
}

func (tester *Tester) TestFinishes(allowedTime int) {
	tester.t.Logf("Testing for FinalTurnComplete event")
	timeout(tester.t, time.Duration(allowedTime)*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.FinalTurnComplete); ok {
				assert(tester.t, e.CompletedTurns == tester.params.Turns,
					"FinalTurnComplete should have a CompletedTurns of %v, not %v", tester.params.Turns, e.CompletedTurns)
				return
			}
		}
	}, "No FinalTurnComplete events received in %v seconds", allowedTime)
}

func (tester *Tester) TestTurnCompleteCount() {
	tester.t.Logf("Testing number of TurnComplete events sent")

	if tester.turn > tester.params.Turns {
		tester.t.Errorf("ERROR: Too many TurnComplete events sent. Should be %v, not %v", tester.params.Turns, tester.turn)
	} else if tester.turn < tester.params.Turns {
		tester.t.Errorf("ERROR: Too few TurnComplete events sent. Should be %v, not %v", tester.params.Turns, tester.turn)
	}
}

func (tester *Tester) TestQuits() {
	tester.t.Logf("Testing for StateChange Quitting event")
	timeout(tester.t, 2*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.StateChange); ok && e.NewState == gol.Quitting {
				return
			}
		}
	}, "No StateChange Quitting events received in 2 seconds")
}

func (tester *Tester) TestNoStateChange(ddl time.Duration) {
	change := make(chan gol.StateChange, 1)
	stop := make(chan bool)

	go func() {
		for {
			select {
			case e := <-tester.eventWatcher:
				if e, ok := e.(gol.StateChange); ok {
					change <- e
					return
				}
			case <-stop:
				return
			}
		}
	}()

	select {
	case <-time.After(ddl):
		stop <- true
	case e := <-change:
		tester.t.Errorf("ERROR: Recieved unexpected StateChange event %v", e)
	}
}

func (tester *Tester) TestOutput() {
	width, height := tester.params.ImageWidth, tester.params.ImageHeight
	tester.t.Logf("Testing image output")

	turn := make(chan int, 1)

	completed := timeout(tester.t, 4*time.Second, func() {
		for e := range tester.eventWatcher {
			if e, ok := e.(gol.ImageOutputComplete); ok {
				assert(tester.t, e.Filename == fmt.Sprintf("%vx%vx%v", width, height, e.CompletedTurns),
					"Filename is not correct")
				turn <- e.CompletedTurns
				return
			}
		}
	}, "No ImageOutput events received in 4 seconds\n%v",
		"If this test is running in WSL2, please make sure the test is located within WSL2 file system rather than Windows! i.e. Your path must not start with /mnt/...")

	if !completed {
		return
	}

	eventTurn := <-turn

	expected := 0
	if eventTurn <= 10000 {
		expected = tester.aliveMap[eventTurn]
	} else if eventTurn%2 == 0 {
		expected = 5565
	} else {
		expected = 5567
	}

	path := fmt.Sprintf("out/%vx%vx%v.pgm", width, height, eventTurn)

	defer func() {
		if r := recover(); r != nil {
			tester.t.Errorf("ERROR: Failed to read image file. Make sure you do ioCheckIdle before sending the ImageOutputComplete\n%v", r)
		}
	}()
	alive := readAliveCells(path, width, height)

	assert(tester.t, len(alive) == expected, "At turn %v expected %v alive cells in output PGM image, got %v instead", eventTurn, expected, len(alive))
}

func timeout(t *testing.T, ddl time.Duration, f func(), msg string, a ...interface{}) bool {
	done := make(chan bool, 1)
	go func() {
		f()
		done <- true
	}()
	select {
	case <-time.After(ddl):
		t.Errorf("ERROR: %v", fmt.Sprintf(msg, a...))
		return false
	case <-done:
		return true
	}
}

func timeoutWarn(ddl time.Duration, f func(), msg string, a ...interface{}) {
	done := make(chan bool, 1)
	go func() {
		f()
		done <- true
	}()
	select {
	case <-time.After(ddl):
		fmt.Printf("WARNING: %v\n", fmt.Sprintf(msg, a...))
	case <-done:
		return
	}
}

func assert(t *testing.T, predicate bool, msg string, a ...interface{}) {
	if !predicate {
		t.Errorf("ERROR: %v", fmt.Sprintf(msg, a...))
	}
}

type LimitedAssert struct {
	t        *testing.T
	failed   bool
	limitHit bool
}

func (l *LimitedAssert) Assert(predicate bool, msg string, a ...interface{}) {
	if !predicate {
		if l.failed {
			l.limitHit = true
		} else {
			l.t.Errorf("ERROR: %v", fmt.Sprintf(msg, a...))
			l.failed = true
		}
	}
}

func (l *LimitedAssert) Reset() {
	l.failed = false
}

func (l *LimitedAssert) LimitHitMessage(msg string) {
	if l.limitHit {
		l.t.Log(msg)
	}
}
