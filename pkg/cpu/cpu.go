package cpu

import (
	"math"
)

type Status struct {
	N, V, P_, B, D, I, Z, C bool
}

const ( // status flag masks
	posC byte = 1 << iota
	posZ
	posI
	posD // decimal would go here but not supporting it
	posB
	pos_ // unused
	posV
	posN
)

// func (status *Status) Clear() {
// 	status = &Status{}
// }

func (status *Status) Get() (b byte) {
	if status.N {
		b |= posN
	}
	if status.V {
		b |= posV
	}
	if status.P_ {
		b |= pos_
	}
	if status.B {
		b |= posB
	}
	if status.D {
		b |= posD
	}
	if status.I {
		b |= posI
	}
	if status.Z {
		b |= posZ
	}
	if status.C {
		b |= posC
	}
	return b
}

func (status *Status) Set(b byte) {
	status.N = (b&posN != 0)
	status.V = (b&posV != 0)
	status.P_ = (b&pos_ != 0)
	status.B = (b&posB != 0)
	status.D = (b&posD != 0)
	status.I = (b&posI != 0)
	status.Z = (b&posZ != 0)
	status.C = (b&posC != 0)
}

type CPU struct {
	pc      uint16
	a, x, y byte
	// status  byte // NV_BDIZC, aka P
	status Status
	s      byte // stack pointer
	memory [0xFFFF]byte
}

// func newCPU() CPU {
// 	return CPU{status: &Status{}}
// }

const (
	implicit = iota // implied?
	accumulator
	immediate
	zeroPage
	zeroPageX
	zeroPageY
	relative
	absolute
	absoluteX
	absoluteY
	indirect
	indexedIndirect
	indirectIndexed
)

// implementing 6502 as described in https://wdc65xx.com/Programming-Manual/ eyes/lichty
// lots of the descriptions taken from https://www.nesdev.org/obelisk-6502-guide/reference.html

func (cpu *CPU) read16(position uint16) uint16 {
	hi, lo := uint16(cpu.memory[position]), uint16(cpu.memory[position+1])
	// a := binary.LittleEndian.Uint16(cpu.memory[position])
	// a2 := int16(a)
	return hi<<8 | lo
}

func (cpu *CPU) reset() {
	cpu.a, cpu.x, cpu.y = 0, 0, 0
	cpu.status = Status{}

	cpu.pc = cpu.read16(0xFFFC)

}

// sets the zero flag if result was 0
func (cpu *CPU) setZ(result byte) {
	cpu.status.Z = result == 0
}

// The negative flag is set if the result of the last operation had bit 7 set to a one (negative 2s complement).
func (cpu *CPU) setN(result byte) {
	cpu.status.N = result&0x80 != 0
}

func (cpu *CPU) setB() {
	cpu.status.B = true
}

// Force Interrupt
//
// The BRK instruction forces the generation of an interrupt request.
// The program counter and processor status are pushed on the stack then the IRQ interrupt vector at $FFFE/F is loaded into the PC and the break flag in the status set to one.
func (cpu *CPU) BRK() {
	cpu.setB()

	// TODO: push PC and status onto the stack
	// TOOD: load IRQ interrup vector at $FFFE/F into the PC
}

// Load Accumulator
//
// load the value into register A and set Z and N flags if value is 0 or negative respectively.
func (cpu *CPU) LDA(value byte) {
	cpu.a = value
	cpu.setZ(cpu.a)
	cpu.setN(cpu.a)
}

//	Transfer Accumulator to X
//
// Copies the current contents of the accumulator into the X register and sets the zero and negative flags as appropriate. (transfer a to x)
func (cpu *CPU) TAX() {
	cpu.x = cpu.a
	cpu.setZ(cpu.x)
	cpu.setN(cpu.x)
}

// Increment X Register
//
// Adds one to the X register setting the zero and negative flags as appropriate.
func (cpu *CPU) INX() {
	cpu.x += 1
	cpu.setZ(cpu.x)
	cpu.setN(cpu.x)
}

func (cpu *CPU) Hotloop(program []byte) {
	if len(program) > math.MaxUint16 {
		return
	}

	cpu.pc = 0
	for !cpu.status.B {
		op := program[cpu.pc]
		cpu.pc += 1
		switch op {
		case 0x00:
			cpu.BRK()
		case 0xa9:
			value := program[cpu.pc]
			cpu.pc += 1
			cpu.LDA(value)
		case 0xaa:
			cpu.TAX()
		case 0xe8:
			cpu.INX()
		}

	}
}
