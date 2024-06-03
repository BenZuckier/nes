package cpu

import "math"

type CPU struct {
	pc      uint16
	a, x, y byte
	status  byte // _NVBDIZC
}

const ( // status flag masks
	c byte = 1 << iota
	z
	i
	d
	b
	v
	n
	_
)

// if code was A9 then next byte is parameter for lda

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

// load the value into register A and set Z and N flags if value is 0 or negative respectively.
func (cpu *CPU) LDA(value byte) {
	cpu.a = value
	cpu.setZ(cpu.a)
	cpu.setN(cpu.a)
}

func (cpu *CPU) TAX() {
	cpu.x = cpu.a
	cpu.setZ(cpu.x)
	cpu.setN(cpu.x)
}

func (cpu *CPU) BRK() {
	cpu.setB()
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
		}

	}
}
