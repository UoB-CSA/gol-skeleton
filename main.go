package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
	"uk.ac.bris.cs/gameoflife/util"
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

	log.Printf("[Main] %-10v %v", "Threads", params.Threads)
	log.Printf("[Main] %-10v %v", "Width", params.ImageWidth)
	log.Printf("[Main] %-10v %v", "Height", params.ImageHeight)
	log.Printf("[Main] %-10v %v", "Turns", params.Turns)

	keyPresses := make(chan rune, 10)
	events := make(chan gol.Event, 1000)

	go sigint()

	go gol.Run(params, events, keyPresses)
	if !*headless {
		sdl.Run(params, events, keyPresses)
	} else {
		sdl.RunHeadless(events)
	}
}

func sigint() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	var exit atomic.Bool
	for range sigint {
		if exit.Load() {
			log.Printf("[Main] %v Force quit by the user", util.Yellow("WARN"))
			os.Exit(0)
		} else {
			log.Printf("[Main] %v Press Ctrl+C again to force quit", util.Yellow("WARN"))
			exit.Store(true)
			go func() {
				time.Sleep(4 * time.Second)
				exit.Store(false)
			}()
		}
	}
}
