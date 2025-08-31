package tests

import (
	"os"
	"runtime/trace"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

// TestTrace is a special test to be used to generate traces - not a real test
func TestTrace(t *testing.T) {
	traceParams := gol.Params{
		Turns:       10,
		Threads:     4,
		ImageWidth:  64,
		ImageHeight: 64,
	}
	f, _ := os.Create("trace.out")
	events := make(chan gol.Event)
	err := trace.Start(f)
	if err != nil {
		t.Fatalf("%v %v", util.Red("ERROR"), err)
	}
	go gol.Run(traceParams, events, nil)
	for range events {
	}
	trace.Stop()
	err = f.Close()
	if err != nil {
		t.Fatalf("%v %v", util.Red("ERROR"), err)
	}
}
