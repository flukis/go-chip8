package main

import (
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	display := NewDisplay()
	emu := NewEmulator(&display)
	if err := emu.LoadROM("samples/IBM Logo.ch8"); err != nil {
		panic(err)
	}

	for {
		emu.Loop()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			}
		}
		sdl.Delay(1000 / 60)
	}

}
