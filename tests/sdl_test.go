package tests

import (
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

// TestSdl tests for correct visualisation in the SDL window
func TestSdl(t *testing.T) {
	t.Run("turn", testSdlTurn)
	t.Run("images", testSdlImages)
	t.Run("alive", testSdlAlive)
}

func testSdlTurn(t *testing.T) {
	params := gol.Params{
		Turns:       100,
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
	tester.SetTestTurn()

	go func() {
		tester.TestFinishes(20)
		tester.TestTurnCompleteCount()
		tester.Stop(false)
	}()

	tester.Loop()
}

func testSdlImages(t *testing.T) {
	params := gol.Params{
		Turns:       100,
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
	tester.SetTestSdl()

	go func() {
		tester.TestStartsExecuting()

		turn, success := tester.AwaitSync() // Before first turn
		if !success {
			tester.Stop(false)
			return
		}

		assert(tester.t, turn == 0, "Turn number should be 0 after first sync, not %v", turn)

		tester.TestImage()
		tester.Continue()

		for turn < 100 {
			turn, success = tester.AwaitSync() // After each turn
			if !success {
				tester.Stop(false)
				return
			}

			if turn == 1 || turn == 100 {
				tester.TestImage()
			}

			tester.Continue()
		}

		tester.Stop(false)
	}()

	tester.Loop()
}

func testSdlAlive(t *testing.T) {
	params := gol.Params{
		Turns:       100,
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
	tester.SetTestSdl()

	go func() {
		tester.TestStartsExecuting()

		turn, success := tester.AwaitSync() // Before first turn
		if !success {
			tester.Stop(false)
			return
		}
		assert(tester.t, turn == 0, "Turn number should be 0 after first sync, not %v", turn)
		tester.Continue()

		for turn < 100 {
			turn, success = tester.AwaitSync() // After each turn
			if !success {
				tester.Stop(false)
				return
			}
			tester.TestAlive()
			tester.Continue()
		}

		tester.Stop(false)
	}()

	tester.Loop()
}
