package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

var (
	RECT_SIZE int32 = 10
)

type Drawer interface {
	Clear()
	Set(x, y int32)
	Draw()
	DestroyWindow()
}

type Display struct {
	drawer Drawer
	data   [8192]uint8
	H, W   uint8
}

func NewDisplay() Display {
	d := NewSDLDisplay(64, 32)
	return Display{
		drawer: d,
		H:      32,
		W:      64,
	}
}

func (s *Display) Clear() {
	s.drawer.Clear()
}

func (s *Display) Draw() {
	s.drawer.Clear()

	for i, value := range s.data {
		y := i / int(s.W)
		x := i % int(s.W)
		if value > 0 {
			s.drawer.Set(int32(x), int32(y))
		}
	}

	s.drawer.Draw()
}

func (s *Display) GetPixel(x, y uint8) uint8 {
	i := uint16(s.W)*uint16(y) + uint16(x)
	return s.data[i]
}
func (s *Display) SetPixel(x, y, value uint8) {
	i := uint16(s.W)*uint16(y) + uint16(x)
	s.data[i] = value
}

type SDLDisplay struct {
	window  *sdl.Window
	surface *sdl.Surface
}

func NewSDLDisplay(w, h int32) *SDLDisplay {
	wScreen := w * RECT_SIZE
	hScreen := h * RECT_SIZE
	window, err := sdl.CreateWindow(
		"chip8",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		wScreen,
		hScreen,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	return &SDLDisplay{
		window:  window,
		surface: surface,
	}
}

func (s *SDLDisplay) Clear() {
	s.surface.FillRect(nil, 0)
}

func (s *SDLDisplay) Set(x, y int32) {
	rect := sdl.Rect{
		X: x * RECT_SIZE,
		Y: y * RECT_SIZE,
		W: RECT_SIZE,
		H: RECT_SIZE,
	}
	s.surface.FillRect(&rect, 0xffffffff)
}

func (s *SDLDisplay) Draw() {
	s.window.UpdateSurface()
}

func (s *SDLDisplay) DestroyWindow() {
	s.window.Destroy()
}
