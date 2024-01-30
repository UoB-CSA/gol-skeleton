package sdl

import (
	"fmt"
	"time"
	"github.com/veandco/go-sdl2/sdl"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

const FPS = 60

func Run(p gol.Params, events <-chan gol.Event, keyPresses chan<- rune) {
	w := NewWindow(int32(p.ImageWidth), int32(p.ImageHeight))
	defer w.Destroy()
	dirty := false
	refreshTicker := time.NewTicker(time.Second / time.Duration(FPS))
	avgTurns := util.NewAvgTurns()

sdl:
	for {
		select {
		case <-refreshTicker.C:
			event := w.PollEvent()
			if event != nil {
				switch e := event.(type) {
				case *sdl.QuitEvent:
					keyPresses <- 'q'
				case *sdl.KeyboardEvent:
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						keyPresses <- 'q'
					case sdl.K_p:
						keyPresses <- 'p'
					case sdl.K_s:
						keyPresses <- 's'
					case sdl.K_q:
						keyPresses <- 'q'
					case sdl.K_k:
						keyPresses <- 'k'
					}
				}
			}
			if dirty {
				w.RenderFrame()
				dirty = false
			}

		case event, ok := <-events:
			if !ok {
				break sdl
			}
			switch e := event.(type) {
			case gol.CellFlipped:
				w.FlipPixel(e.Cell.X, e.Cell.Y)
			case gol.CellsFlipped:
				for _, cell := range e.Cells {
					w.FlipPixel(cell.X, cell.Y) 
				}
			case gol.TurnComplete:
				dirty = true
			case gol.AliveCellsCount:
				fmt.Printf("Completed Turns %-8v %-20v Avg%+5v turns/sec\n", event.GetCompletedTurns(), event, avgTurns.Get(event.GetCompletedTurns()))
			case gol.FinalTurnComplete:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
			case gol.ImageOutputComplete:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
			case gol.StateChange:
				fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
				if e.NewState == gol.Quitting {
					break sdl
				}
			}
		}
	}
}

func RunHeadless(events <-chan gol.Event) {
	avgTurns := util.NewAvgTurns()
	for event := range events {
		switch e := event.(type) {
		case gol.AliveCellsCount:
			fmt.Printf("Completed Turns %-8v %-20v Avg%+5v turns/sec\n", event.GetCompletedTurns(), event, avgTurns.Get(event.GetCompletedTurns()))
		case gol.FinalTurnComplete:
			fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), "Final Turn Complete")
		case gol.ImageOutputComplete:
			fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
		case gol.StateChange:
			fmt.Printf("Completed Turns %-8v %v\n", event.GetCompletedTurns(), event)
			if e.NewState == gol.Quitting {
				break
			}
		}
	}
}
