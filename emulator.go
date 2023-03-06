package main

import (
	"fmt"
	"log"
	"os"
)

var fonts = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, //0
	0x20, 0x60, 0x20, 0x20, 0x70, //1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
	0x90, 0x90, 0xF0, 0x10, 0x10, //4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
	0xF0, 0x10, 0x20, 0x40, 0x40, //7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
	0xF0, 0x90, 0xF0, 0x90, 0x90, //A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
	0xF0, 0x80, 0x80, 0x80, 0xF0, //C
	0xE0, 0x90, 0x90, 0x90, 0xE0, //D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
	0xF0, 0x80, 0xF0, 0x80, 0x80, //F
}

type Emulator struct {
	Memory [4096]uint8 // Memory size 4096 / 4k

	PC uint16 // Program Counter
	OC uint16 // Current Opcode
	SP uint16 // stack pointer
	IV uint16 // Index Register

	VX    [16]uint8
	Key   [16]uint8
	Stack [16]uint16

	delayTime uint8

	Display *Display
}

func NewEmulator(display *Display) Emulator {
	emu := Emulator{
		PC:      0x200,
		Display: display,
	}

	for i := 0; i < len(fonts); i++ {
		emu.Memory[i] = fonts[i]
	}

	return emu
}

func (e *Emulator) LoadROM(pathfile string) error {
	file, err := os.OpenFile(pathfile, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	fStat, err := file.Stat()
	if err != nil {
		return err
	}

	if int64(len(e.Memory)-512) < fStat.Size() {
		return fmt.Errorf("ROM size bigger than memory")
	}

	buffer := make([]byte, fStat.Size())
	if _, readErr := file.Read(buffer); readErr != nil {
		return readErr
	}

	for i := 0; i < len(buffer); i++ {
		e.Memory[512+uint16(i)] = buffer[i]
	}

	return nil
}

func (e *Emulator) Loop() {
	e.OC = (uint16(e.Memory[e.PC]) << 8) | uint16(e.Memory[e.PC+1])

	switch e.OC & 0xF000 {
	case 0x0000:
		switch e.OC & 0x000F {
		case 0x0:
			e.ClearScreen()
		case 0x000E:
			e.SP = e.SP - 1
			e.PC = e.Stack[e.SP]
			e.PC = e.PC + 2
		default:
			log.Printf("Invalid opcode %X\n", e.OC)
		}
	case 0x1000:
		e.Jump()
	case 0x2000:
		e.Call()
	case 0x6000:
		e.SetRegisterVX()
	case 0x7000:
		e.AddValueToReg()
	case 0xA000:
		e.SetIndexReg()
	case 0xD000:
		e.Draw()
	default:
		log.Printf("Invalid opcode %X\n", e.OC)
	}

	if e.delayTime > 0 {
		e.delayTime = e.delayTime - 1
	}
}

func (e *Emulator) ClearScreen() {
	log.Println("00E0 - clear screen")
	e.Display.Clear()
	e.Display.Draw()
	e.PC = e.PC + 2
}

func (e *Emulator) Jump() {
	log.Println("1NNN - jump")
	e.PC = e.OC & 0x0FFF
}

func (e *Emulator) Call() {
	log.Println("2NNN - call")
	e.SP = e.SP + 1
	e.Stack[e.SP] = e.PC
	e.PC = e.OC & 0x0FFF
}

func (e *Emulator) SetRegisterVX() {
	log.Println("6XNN - set register VX")
	e.VX[(e.OC&0x0F00)>>8] = uint8(e.OC & 0x00FF)
	e.PC = e.PC + 2
}

func (e *Emulator) AddValueToReg() {
	log.Printf("7XNN - add value to register")
	e.VX[(e.OC&0x0F00)>>8] = e.VX[(e.OC&0xF00)>>8] + uint8(e.OC&0x00FF)
	e.PC = e.PC + 2
}

func (e *Emulator) SetIndexReg() {
	log.Println("ANNN - set index register 1")
	e.IV = e.OC & 0x0FFF
	e.PC = e.PC + 2
}

func (e *Emulator) Draw() {
	log.Println("DXYN - display/draw")
	maxX := e.Display.W - 1
	maxY := e.Display.H - 1
	x := e.VX[(e.OC&0x0F00)>>8] % maxX
	y := e.VX[(e.OC&0x00F0)>>4] % maxY
	h := e.OC & 0x000F

	e.VX[0xF] = 0
	for i := 0; i < int(h); i++ {
		nthByte := e.Memory[e.IV+uint16(i)]
		for j := 7; j >= 0; j-- {
			screenX := x + uint8(7-j)
			screenY := y + uint8(i)

			spritePixel := (nthByte >> j) & 0x01
			screenPixel := e.Display.GetPixel(screenX, screenY)

			if screenX > maxX || screenY > maxY {
				continue
			}

			val := spritePixel ^ screenPixel
			if spritePixel == 1 && screenPixel == 1 {
				e.VX[0xF] = 1
			}
			e.Display.SetPixel(screenX, screenY, val)
		}
	}

	e.Display.Draw()
	e.PC = e.PC + 2
}
