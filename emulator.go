package main

import (
	"fmt"
	"log"
	"math/rand"
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

	delayTimer uint8

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

func (e *Emulator) KeyStroke(i uint8, isDown bool) {
	if isDown {
		e.Key[i] = 1
	} else {
		e.Key[i] = 0
	}
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

	if int64(len(e.Memory)-0x200) < fStat.Size() {
		return fmt.Errorf("ROM size bigger than memory")
	}

	buffer := make([]byte, fStat.Size())
	if _, readErr := file.Read(buffer); readErr != nil {
		return readErr
	}

	for i := 0; i < len(buffer); i++ {
		e.Memory[0x200+uint16(i)] = buffer[i]
	}

	return nil
}

func (e *Emulator) Loop() {
	e.OC = (uint16(e.Memory[e.PC]) << 8) | uint16(e.Memory[e.PC+1])

	vx := e.VX[(e.OC&0x0F00)>>8]
	vy := e.VX[(e.OC&0x00F0)>>4]

	switch e.OC & 0xF000 {
	case 0x0000:
		e.ClearScreen()
	case 0x1000:
		e.Jump()
	case 0x2000:
		e.Call()
	case 0x3000:
		e.SkipIfVXEqualNN(vx)
	case 0x4000:
		e.SkipIfNotVXEqualNN(vx)
	case 0x5000:
		e.SkipIfVXEqualVY(vx, vy)
	case 0x6000:
		e.SetRegisterVX()
	case 0x7000:
		e.AddValueToReg(vx)
	case 0x8000:
		e.LogicalAndArithmatic(vx, vy)
	case 0x9000:
		e.SkipIfVXNotEqualVY(vx, vy)
	case 0xA000:
		e.SetIndexReg()
	case 0xB000:
		e.JumpPlusV0()
	case 0xC000:
		e.SetVX()
	case 0xD000:
		e.Draw(vx, vy)
	case 0xE000:
		e.SkipByKey(vx)
	case 0xF000:
		e.Instructions(vx, vy)
	default:
		log.Panicf("Invalid opcode %X\n", e.OC)
	}

	if e.delayTimer > 0 {
		e.delayTimer = e.delayTimer - 1
	}
}

func (e *Emulator) next() {
	e.PC = e.PC + 2
}

func (e *Emulator) skipnext() {
	e.PC = e.PC + 4
}

func (e *Emulator) ClearScreen() {
	switch e.OC & 0x000F {
	case 0x0000:
		// 00E0 - clear screen
		e.Display.Clear()
		e.Display.Draw()
		e.next()
	case 0x000E:
		e.PC = e.Stack[e.SP]
		e.SP -= 1
		e.next()
	default:
		log.Panicf("Invalid opcode %X\n", e.OC)
	}
}

func (e *Emulator) Jump() {
	// 1NNN - jump
	e.PC = e.OC & 0x0FFF
}

func (e *Emulator) Call() {
	// 2NNN - call
	e.SP = e.SP + 1
	e.Stack[e.SP] = e.PC
	e.PC = e.OC & 0x0FFF
}

func (e *Emulator) SkipIfVXEqualNN(vx uint8) {
	// 3NNN - skip if VX == NN
	if uint16(vx) == e.OC&0x00FF {
		e.skipnext() // skip
	} else {
		e.next()
	}
}

func (e *Emulator) SkipIfNotVXEqualNN(vx uint8) {
	// 4NNN - skip if VX != NN
	if uint16(vx) != e.OC&0x00FF {
		e.skipnext() // skip
	} else {
		e.next()
	}
}

func (e *Emulator) SkipIfVXEqualVY(vx, vy uint8) {
	// 5NNN - skip if VX == VY
	if vx == vy {
		e.skipnext() // skip
	} else {
		e.next()
	}
}

func (e *Emulator) SetRegisterVX() {
	// 6XNN - set register VX
	e.VX[(e.OC&0x0F00)>>8] = uint8(e.OC & 0x00FF)
	e.next()
}

func (e *Emulator) AddValueToReg(vx uint8) {
	// 7XNN - add value to register
	e.VX[(e.OC&0x0F00)>>8] = vx + uint8(e.OC&0x00FF)
	e.next()
}

func (e *Emulator) LogicalAndArithmatic(vx, vy uint8) {
	// 8XNN - logical and arithmatic
	vxAddr := (e.OC & 0x0F00) >> 8
	switch e.OC & 0x000F {
	case 0x0000:
		// set VX to value of VY
		vx = vy
		e.next()
	case 0x0001:
		// set vx to vx or vy
		vx = vx | vy
		e.next()
	case 0x0002:
		// set vx to vx and vy
		vx = vx & vy
		e.next()
	case 0x0003:
		// set vx to vx xor vy
		vx = vx ^ vy
		e.next()
	case 0x0004:
		// add vy to vx, vf is set to 1 when no carry, and 0 for otherwise
		if vy > 0xFF-vx {
			e.VX[0xF] = 1
		} else {
			e.VX[0xF] = 0
		}
		e.VX[vxAddr] = vx + vy
		e.next()
	case 0x0005:
		// substract vy from vx, check the borrow
		if vy > vx {
			e.VX[0xF] = 0
		} else {
			e.VX[0xF] = 1
		}
		e.VX[vxAddr] = vx - vy
		e.next()
	case 0x0006:
		// right by 1 and store res to VX
		e.VX[0xF] = vx & 0x1
		e.VX[vxAddr] = vx >> 1
		e.next()
	case 0x0007:
		// set vx to vy minus vx
		if vx > vy {
			e.VX[0xF] = 0
		} else {
			e.VX[0xF] = 1
		}
		e.VX[vxAddr] = vy - vx
		e.next()
	case 0x000E:
		// shift vy left one an copy to vx
		e.VX[0xF] = vx >> 7
		e.VX[vxAddr] = vx << 1
		e.next()
	default:
		log.Panicf("Invalid opcode %X\n", e.OC)
	}
}

func (e *Emulator) SkipIfVXNotEqualVY(vx, vy uint8) {
	// 9NNN - skip if VX != VY
	if vx != vy {
		e.skipnext() // skip
	} else {
		e.next()
	}
}

func (e *Emulator) SetIndexReg() {
	// ANNN - set index register 1
	e.IV = e.OC & 0x0FFF
	e.next()
}

func (e *Emulator) JumpPlusV0() {
	// BNNN - jump to address NNN + v0
	e.PC = e.OC&0x0FFF + uint16(e.VX[0x0])
}

func (e *Emulator) SetVX() {
	// CXNN - sets vx
	e.VX[(e.OC&0x0F00)>>8] = uint8(rand.Intn(256)) & uint8(e.OC&0x00FF)
	e.next()
}

func (e *Emulator) Draw(vx, vy uint8) {
	// DXYN - display/draw
	maxX := e.Display.W - 1
	maxY := e.Display.H - 1
	x := vx % maxX
	y := vy % maxY
	h := e.OC & 0x000F

	e.VX[0xF] = 0
	for i := 0; i < int(h); i++ {
		nthByte := e.Memory[e.IV+uint16(i)]
		for j := 7; j >= 0; j-- {
			screenX := x + uint8(7-j)
			screenY := y + uint8(i)

			spritePixel := ((nthByte >> j) & 0x01) == 1
			screenPixel := e.Display.GetPixel(screenX, screenY)

			if screenX > maxX || screenY > maxY {
				continue
			}

			val := spritePixel || screenPixel
			if spritePixel && screenPixel {
				e.VX[0xF] = 1
			}
			e.Display.SetPixel(screenX, screenY, val)
		}
	}

	e.Display.Draw()
	e.next()
}

func (e *Emulator) SkipByKey(vx uint8) {
	switch e.OC & 0x00FF {
	case 0x009E:
		// skip if key in vx pressed
		if e.Key[vx] == 1 {
			e.skipnext()
		} else {
			e.next()
		}
	case 0x00A1:
		// skip if key in vs not pressed
		if e.Key[vx] == 0 {
			e.skipnext()
		} else {
			e.next()
		}
	default:
		log.Panicf("Invalid opcode %X\n", e.OC)
	}
}

func (e *Emulator) Instructions(vx, vy uint8) {
	vxAddr := (e.OC & 0x0F00) >> 8
	switch e.OC & 0x00FF {
	case 0x0007:
		// 0xFX07 set to delay timer
		e.VX[vxAddr] = e.delayTimer
		e.next()
	case 0x0015:
		// 0xFX15 set delay timer to vx
		e.delayTimer = vx
		e.next()
	case 0x0018:
		// 0xFX18 set delay timer to vy
		e.delayTimer = vy
		e.next()
	case 0x000A:
		// 0xFX0A  get key from keypad
		isPressed := false
		for i := 0; i < len(e.Key); i++ {
			if e.Key[i] != 0 {
				e.VX[vxAddr] = uint8(i)
				isPressed = true
			}
		}
		if !isPressed {
			return
		}
		e.next()
	case 0x001E:
		// 0xFX1E add vx to index
		if e.IV+uint16(vx) > 0xFFF {
			e.VX[0xF] = 1
		} else {
			e.VX[0xF] = 0
		}
		e.IV = e.IV + uint16(vx)
		e.next()
	case 0x0029:
		// 0xFX29 font character
		e.IV = uint16(vx) * 0x5
		e.next()
	case 0x0033:
		// 0xFX33 binary coded decimal conversion
		e.Memory[e.IV] = vx / 100
		e.Memory[e.IV+1] = (vx / 10) % 10
		e.Memory[e.IV+2] = (vx % 100) / 10
		e.next()
	case 0x0055:
		// 0xFX55 store v0 to vx
		for i := 0; i < int(vxAddr)+1; i++ {
			e.Memory[uint16(i)+e.IV] = e.VX[i]
		}
		e.IV = uint16(vx + 1)
		e.next()
	case 0x0065:
		// 0xFX65 fill v0 to vx
		for i := 0; i < int(vxAddr)+1; i++ {
			e.VX[i] = e.Memory[e.IV+uint16(i)]
		}
		e.IV = uint16(vx + 1)
		e.next()
	default:
		log.Panicf("Invalid opcode %X\n", e.OC)
	}

}
