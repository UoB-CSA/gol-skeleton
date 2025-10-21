package gol

import (
	"fmt"

	"uk.ac.bris.cs/gameoflife/util"
)

// Event represents any Game of Life event that needs to be communicated to the user.
type Event interface {
	// Stringer allows each event to be printed by the GUI
	fmt.Stringer
	// GetCompletedTurns should return the number of fully completed turns.
	// If the 0th turn is finished, this should return 1.
	GetCompletedTurns() int
}

// `AliveCellsCount` is an Event notifying the user about the number of currently alive cells.
// This Event should be sent every 2s.
type AliveCellsCount struct { // implements Event
	CompletedTurns int
	CellsCount     int
}

// `ImageOutputComplete` is an Event notifying the user about the completion of output.
// This Event should be sent every time an image has been saved.
type ImageOutputComplete struct { // implements Event
	CompletedTurns int
	Filename       string
}

// State represents a change in the state of execution.
type State int

const (
	Paused State = iota
	Executing
	Quitting
)

// `StateChange` is an Event notifying the user about the change of state of execution.
// This Event should be sent every time the execution is paused, resumed or quit.
type StateChange struct { // implements Event
	CompletedTurns int
	NewState       State
}

// `CellFlipped` is an Event notifying the GUI about a change of state of a single cell.
// This event should be sent every time a cell changes state.
// Make sure to send this event for all cells that are alive when the image is loaded in.
type CellFlipped struct { // implements Event
	CompletedTurns int
	Cell           util.Cell
}

// `CellsFlipped` is an Event notifying the GUI about a change of state of many cells.
// You can collect many flipped cells and send `CellsFlipped` at a time instead of sending `CellFlipped` for every flipped cell.
// You can send many times of `CellsFlipped` event in a turn, i.e., each worker could send `CellsFlipped`.
// **Please be careful not to send `CellFlipped` and `CellsFlipped` at the same time, as they may conflict.**
// Choose one of them.
type CellsFlipped struct { // implements Event
	CompletedTurns int
	Cells          []util.Cell
}

// `TurnComplete` is an Event notifying the GUI about turn completion.
// SDL will render a frame when this event is sent.
// All `CellFlipped` or `CellsFlipped` events must be sent *before* `TurnComplete`.
type TurnComplete struct { // implements Event
	CompletedTurns int
}

// `FinalTurnComplete` is an Event notifying the testing framework about the new world state after execution finished.
// The data included with this Event is used directly by the tests.
// SDL closes the window when this Event is sent.
type FinalTurnComplete struct {
	CompletedTurns int
	Alive          []util.Cell
}

// String methods allow the different types of Events and States to be printed.

func (state State) String() string {
	switch state {
	case Paused:
		return "Paused"
	case Executing:
		return "Executing"
	case Quitting:
		return "Quitting"
	default:
		return "Incorrect State"
	}
}

func (event StateChange) String() string {
	return fmt.Sprintf("%v", event.NewState)
}

func (event StateChange) GetCompletedTurns() int {
	return event.CompletedTurns
}

func (event AliveCellsCount) String() string {
	return fmt.Sprintf("Alive Cells %v", event.CellsCount)
}

func (event AliveCellsCount) GetCompletedTurns() int {
	return event.CompletedTurns
}

func (event ImageOutputComplete) String() string {
	return fmt.Sprintf("File %v.pgm output done", event.Filename)
}

func (event ImageOutputComplete) GetCompletedTurns() int {
	return event.CompletedTurns
}

func (event CellFlipped) String() string {
	return ""
}

func (event CellFlipped) GetCompletedTurns() int {
	return event.CompletedTurns
}

func (event CellsFlipped) String() string {
	return ""
}

func (event CellsFlipped) GetCompletedTurns() int {
	return event.CompletedTurns
}

func (event TurnComplete) String() string {
	return ""
}

func (event TurnComplete) GetCompletedTurns() int {
	return event.CompletedTurns
}

func (event FinalTurnComplete) String() string {
	return "Final Turn Complete"
}

func (event FinalTurnComplete) GetCompletedTurns() int {
	return event.CompletedTurns
}

// This might all seem like weird syntax to you...
// You have however seen something similar to it before in first year.

// In the Go code an Interface called Event is created, this provides a set of methods that
// need to be defined for something to have the type Event.

// This is a similar concept to typeclasses in Haskell. A typeclass called Event could be defined.
// It would require two methods to be implemented: string and getCompletedTurns. Note the
// similarities between the type signatures of the Go and Haskell functions.

/*
> class Event event where
>   string :: event -> String
>   getCompletedTurns :: event -> Int
*/

// A new data type called ImageOutputComplete can then be created, just like in Go.

/*
> data ImageOutputComplete = ImageOutputComplete Int String
*/

// Now in the Go code extension methods are created for the ImageOutputComplete so that it
// provides the methods required for the Event Inteface. Similarly in Haskell, an instance
// of the typeclass Event can be created.

/*
> instance Event ImageOutputComplete where
>   string (ImageOutputComplete t f) = concat ["Turn ", show t, " - File ", f, " output complete"]
>   getCompletedTurns (ImageOutputComplete t f) = t
*/
