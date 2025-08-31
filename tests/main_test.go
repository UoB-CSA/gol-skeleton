package tests

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"uk.ac.bris.cs/gameoflife/sdl"
	"uk.ac.bris.cs/gameoflife/util"
)

var w *sdl.Window
var flipCellChan chan util.Cell
var refreshChan chan struct{}
var clearPixelsChan chan struct{}

func TestMain(m *testing.M) {
	runtime.LockOSThread()

	// Switch the working directory from the `tests` subdirectory
	// to the root directory to ensure correct access to test resources
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filename))
	if err := os.Chdir(rootDir); err != nil {
		panic(err)
	}

	// Disable normal log, use test log only
	log.SetOutput(io.Discard)

	var sdlFlag = flag.Bool(
		"sdl",
		false,
		"Enable the SDL window for testing.")

	flag.Parse()
	done := make(chan int, 1)
	test := func() { done <- m.Run() }
	if !*sdlFlag {
		go test()
	} else {
		w = sdl.NewWindow(512, 512)
		flipCellChan = make(chan util.Cell, 1000)
		refreshChan = make(chan struct{}, 1)
		clearPixelsChan = make(chan struct{}, 1)
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
			case cell := <-flipCellChan:
				w.FlipPixel(cell.X, cell.Y)
			case <-clearPixelsChan:
				w.ClearPixels()
				w.RenderFrame()
			case <-refreshChan:
				dirty = true
			}
		}
	}
	os.Exit(<-done)
}

func flipCell(cell util.Cell) {
	if flipCellChan != nil {
		flipCellChan <- cell
	}
}

func refresh() {
	if refreshChan != nil {
		refreshChan <- struct{}{}
	}
}

func clearPixels() {
	if clearPixelsChan != nil {
		clearPixelsChan <- struct{}{}
	}
}
