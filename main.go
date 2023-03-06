package main

import (
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	display := NewDisplay()
	emu := NewEmulator(&display)
	if err := emu.LoadROM("roms/demos/Maze [David Winter, 199x].ch8"); err != nil {
		panic(err)
	}

	for {
		emu.Loop()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch et := event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if et.Type == sdl.KEYUP {
					switch et.Keysym.Sym {
					case sdl.K_1:
						emu.KeyStroke(0x1, false)
					case sdl.K_2:
						emu.KeyStroke(0x2, false)
					case sdl.K_3:
						emu.KeyStroke(0x3, false)
					case sdl.K_4:
						emu.KeyStroke(0xC, false)
					case sdl.K_q:
						emu.KeyStroke(0x4, false)
					case sdl.K_w:
						emu.KeyStroke(0x5, false)
					case sdl.K_e:
						emu.KeyStroke(0x6, false)
					case sdl.K_r:
						emu.KeyStroke(0xD, false)
					case sdl.K_a:
						emu.KeyStroke(0x7, false)
					case sdl.K_s:
						emu.KeyStroke(0x8, false)
					case sdl.K_d:
						emu.KeyStroke(0x9, false)
					case sdl.K_f:
						emu.KeyStroke(0xE, false)
					case sdl.K_z:
						emu.KeyStroke(0xA, false)
					case sdl.K_x:
						emu.KeyStroke(0x0, false)
					case sdl.K_c:
						emu.KeyStroke(0xB, false)
					case sdl.K_v:
						emu.KeyStroke(0xF, false)
					}
				} else if et.Type == sdl.KEYDOWN {
					switch et.Keysym.Sym {
					case sdl.K_1:
						emu.KeyStroke(0x1, true)
					case sdl.K_2:
						emu.KeyStroke(0x2, true)
					case sdl.K_3:
						emu.KeyStroke(0x3, true)
					case sdl.K_4:
						emu.KeyStroke(0xC, true)
					case sdl.K_q:
						emu.KeyStroke(0x4, true)
					case sdl.K_w:
						emu.KeyStroke(0x5, true)
					case sdl.K_e:
						emu.KeyStroke(0x6, true)
					case sdl.K_r:
						emu.KeyStroke(0xD, true)
					case sdl.K_a:
						emu.KeyStroke(0x7, true)
					case sdl.K_s:
						emu.KeyStroke(0x8, true)
					case sdl.K_d:
						emu.KeyStroke(0x9, true)
					case sdl.K_f:
						emu.KeyStroke(0xE, true)
					case sdl.K_z:
						emu.KeyStroke(0xA, true)
					case sdl.K_x:
						emu.KeyStroke(0x0, true)
					case sdl.K_c:
						emu.KeyStroke(0xB, true)
					case sdl.K_v:
						emu.KeyStroke(0xF, true)
					}
				}
			}
		}
		sdl.Delay(1000 / 60)
	}

}
