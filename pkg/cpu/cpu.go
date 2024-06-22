package cpu

import (
	"math"
)

type CPU struct {
	pc      uint16
	a, x, y byte
	status  byte // NV_BDIZC, aka P
	s       byte // stack pointer
	memory  [0xFFFF]byte
}

const ( // status flag masks
	c byte = 1 << iota
	z
	i
	_ // decimal would go here but not supporting it
	b
	_
	v
	n
)

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
	cpu.status = 0

	cpu.pc = cpu.read16(0xFFFC)

}

// sets the zero flag if result was 0
func (cpu *CPU) setZ(result byte) {
	if result == 0 {
		cpu.status |= z
	} else {
		cpu.status &= ^z //
	}
}

// The negative flag is set if the result of the last operation had bit 7 set to a one (negative 2s complement).
func (cpu *CPU) setN(result byte) {
	if result&0x80 != 0 {
		cpu.status |= n
	} else {
		cpu.status &= ^n
	}
}

func (cpu *CPU) setB() {
	cpu.status |= b
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
	for cpu.status&b == 0 {
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
