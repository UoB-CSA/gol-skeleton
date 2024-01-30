package main

import (
	"flag"
	"fmt"
	"runtime"
	"os"
	"os/signal"
	"syscall"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
)

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread()
	var params gol.Params

	flag.IntVar(
		&params.Threads,
		"t",
		8,
		"Specify the number of worker threads to use. Defaults to 8.")

	flag.IntVar(
		&params.ImageWidth,
		"w",
		512,
		"Specify the width of the image. Defaults to 512.")

	flag.IntVar(
		&params.ImageHeight,
		"h",
		512,
		"Specify the height of the image. Defaults to 512.")

	flag.IntVar(
		&params.Turns,
		"turns",
		10000000000,
		"Specify the number of turns to process. Defaults to 10000000000.")

	headless := flag.Bool(
		"headless",
		false,
		"Disable the SDL window for running in a headless environment.")

	flag.Parse()

	fmt.Printf("%-10v %v\n", "Threads", params.Threads)
	fmt.Printf("%-10v %v\n", "Width", params.ImageWidth)
	fmt.Printf("%-10v %v\n", "Height", params.ImageHeight)
	fmt.Printf("%-10v %v\n", "Turns", params.Turns)

	keyPresses := make(chan rune, 10)
	events := make(chan gol.Event, 1000)

	go sigterm(keyPresses)

	go gol.Run(params, events, keyPresses)
	if !(*headless) {
		sdl.Run(params, events, keyPresses)
	} else {
		sdl.RunHeadless(events)
	}
}

func sigterm(keyPresses chan<- rune) {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)
	<-sigterm
	keyPresses <- 'q'
}
