package tests

import (
	"sync"
	"testing"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

// TestKeyboard tests key presses and events
func TestKeyboard(t *testing.T) {
	t.Run("p", testKeyboardP)
	t.Run("s", testKeyboardS)
	t.Run("q", testKeyboardQ)
	t.Run("p+s", testKeyboardPS)
	t.Run("p+q", testKeyboardPQ)
}

func testKeyboardP(t *testing.T) {
	params := gol.Params{
		Turns:       20,
		Threads:     8,
		ImageWidth:  512,
		ImageHeight: 512,
	}

	keyPresses := make(chan rune, 10)
	events := make(chan gol.Event, 1000)

	allowDone := false
	allowDoneMutex := sync.Mutex{}

	golDone := make(chan bool, 1)

	keyPresses <- 'p'

	go func() {
		gol.Run(params, events, keyPresses)
		golDone <- true

		allowDoneMutex.Lock()
		if !allowDone {
			t.Errorf(
				"%v Your program has returned from the gol.Run function before it was unpaused",
				util.Red("ERROR"),
			)
		}
		allowDoneMutex.Unlock()
	}()

	tester := MakeTester(t, params, keyPresses, events, golDone)

	go func() {
		tester.TestStartsExecuting()

		turn := tester.TestPauses()

		tester.TestNoStateChange(5 * time.Second)

		allowDoneMutex.Lock()
		allowDone = true
		allowDoneMutex.Unlock()
		keyPresses <- 'p'

		tester.TestExecutes(turn)

		tester.Stop(false)
	}()

	tester.Loop()
}

func testKeyboardS(t *testing.T) {
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

	tester := MakeTester(t, params, keyPresses, events, golDone)

	go func() {
		tester.TestStartsExecuting()

		time.Sleep(500 * time.Millisecond)
		keyPresses <- 's'
		tester.TestOutput()

		keyPresses <- 'q'
		tester.Stop(false)
	}()

	tester.Loop()
}

func testKeyboardQ(t *testing.T) {
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

	tester := MakeTester(t, params, keyPresses, events, golDone)

	go func() {
		tester.TestStartsExecuting()

		time.Sleep(500 * time.Millisecond)

		keyPresses <- 'q'
		tester.TestOutput()
		tester.TestQuits()
		tester.Stop(true)
	}()

	tester.Loop()
}

func testKeyboardPS(t *testing.T) {
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

	tester := MakeTester(t, params, keyPresses, events, golDone)

	go func() {
		tester.TestStartsExecuting()

		time.Sleep(500 * time.Millisecond)

		keyPresses <- 'p'
		tester.TestPauses()

		keyPresses <- 's'
		tester.TestOutput()

		keyPresses <- 'p'
		keyPresses <- 'q'
		tester.Stop(false)
	}()

	tester.Loop()
}

func testKeyboardPQ(t *testing.T) {
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

	tester := MakeTester(t, params, keyPresses, events, golDone)

	go func() {
		tester.TestStartsExecuting()

		time.Sleep(500 * time.Millisecond)

		keyPresses <- 'p'
		tester.TestPauses()

		keyPresses <- 'q'
		tester.TestOutput()
		tester.TestQuits()
		tester.Stop(true)
	}()

	tester.Loop()
}
