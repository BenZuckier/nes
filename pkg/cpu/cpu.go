package cpu

import (
	"fmt"
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

func (status *Status) Clear() {
	status.Set(0)
}

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
	status  Status
	// stack pointer. really uint16 but the high byte is always 0x01 so the effective addr is `0x01 | CPU.s`
	s byte
	// 0x0000-0x00FF is zero page (first page)
	//
	// 0x0100-0x01FF is system stack (second page)
	//
	// last 6 bytes 0xFFFA-0xFFFF must be programmed with the addresses of
	// the non-maskable interrupt handler (0xFFFA/B),
	// the power on reset location (0xFFFC/D)
	// and the BRK/interrupt request handler (0xFFFE/F).
	//
	// https://www.nesdev.org/obelisk-6502-guide/architecture.html
	memory  [0xFFFF]byte
	opcodes map[byte]opcode
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
	indirectX
	indirectY
)

var modeNames = map[int]string{
	implicit:    "implicit",
	accumulator: "accumulator",
	immediate:   "immediate",
	zeroPage:    "zeroPage",
	zeroPageX:   "zeroPageX",
	zeroPageY:   "zeroPageY",
	relative:    "relative",
	absolute:    "absolute",
	absoluteX:   "absoluteX",
	absoluteY:   "absoluteY",
	indirect:    "indirect",
	indirectX:   "indirectX", // indexedIndirect, (Indirect,X)
	indirectY:   "indirectY", // indirectIndexed, (Indirect),Y
}

// implementing 6502 as described in https://wdc65xx.com/Programming-Manual/ eyes/lichty
// lots of the descriptions taken from https://www.nesdev.org/obelisk-6502-guide/reference.html

// read returns the byte stored at the 16 bit position in memory
func (cpu *CPU) read(pos uint16) byte {
	return cpu.memory[pos]
}

// write stores the given byte `dat` into the 16 bit position in memory
func (cpu *CPU) write(pos uint16, dat byte) {
	cpu.memory[pos] = dat
}

func (cpu *CPU) read16(pos uint16) uint16 {

	lo, hi := uint16(cpu.memory[pos]), uint16(cpu.memory[pos+1])
	return hi<<8 | lo

}

func (cpu *CPU) write16(pos, dat uint16) {

	cpu.write(pos, byte(dat))
	cpu.write(pos+1, byte(dat>>8))
}

// push writes the byte dat to the position pointed to by [CPU.s] and grows the stack downwards.
func (cpu *CPU) push(dat byte) {
	cpu.write(cpu.effStack(), dat)
	cpu.s -= 1
}

// pop increments the stack pointer and gets the byte at the top of the stack.
func (cpu *CPU) pop() byte {
	cpu.s += 1
	return cpu.read(cpu.effStack())
}

// push16 writes the uint16 dat to the position pointed to by [CPU.s] and grows the stack downwards twice.
func (cpu *CPU) push16(dat uint16) {
	cpu.push(byte(dat >> 8)) // hi
	cpu.push(byte(dat))
}

// pop16 increments the stack pointer twice and gets the uint16 at the top of the stack.
func (cpu *CPU) pop16() uint16 {
	return uint16(cpu.pop()<<8) | uint16(cpu.pop()) // hi << 8
}

// effStack gets the effective stack pointer into memory by adding 0x0100 as the high byte to the supplied low byte stack pointer.
func (cpu *CPU) effStack() uint16 {
	return stackOffset | uint16(cpu.s)
}

const stackOffset = uint16(0x0100)
const pcInitAddr = uint16(0xFFFC)
const stackInit = byte(0xff - 2) // simulates stack pointer being at ff and "secretly" pushing PC and pushing P which increments p by 2 like BRK or IRQ

func (cpu *CPU) reset() {
	cpu.a, cpu.x, cpu.y = 0, 0, 0
	cpu.status.Set(0x24) // clear and set B and I

	cpu.pc = cpu.read16(pcInitAddr)
	cpu.s = stackInit

}

// sets the zero flag if result was 0
func (cpu *CPU) setZ(result byte) {
	cpu.status.Z = result == 0
}

// The negative flag is set if the result of the last operation had bit 7 set to a one (negative 2s complement).
func (cpu *CPU) setN(result byte) {
	cpu.status.N = result&0x80 != 0
}

// setZN sets the zero flag [CPU.setZ] if result was 0 and the negative flag [CPU.setN] if the result of the last operation had bit 7 set to a one.
func (cpu *CPU) setZN(result byte) {
	cpu.setZ(result)
	cpu.setN(result)
}

func (cpu *CPU) setB() {
	cpu.status.B = true
}

// Force Interrupt
//
// The brk instruction forces the generation of an interrupt request.
// The program counter and processor status are pushed on the stack then the IRQ interrupt vector at 0xFFFE/F is loaded into the PC and the break flag in the status set to one.
func (cpu *CPU) brk(opDat) {
	cpu.push16(cpu.pc)
	cpu.setB()

	// TODO: push PC and status onto the stack
	// TODO: load IRQ interrupt vector at 0xFFFE/F into the PC
}

// Load Accumulator
//
// load the value into register A and set Z and N flags if value is 0 or negative respectively.
func (cpu *CPU) lda(dat opDat) {
	cpu.a = cpu.read(dat.addr)
	cpu.setZN(cpu.a)
}

//	Transfer Accumulator to X
//
// Copies the current contents of the accumulator into the X register and sets the zero and negative flags as appropriate. (transfer a to x)
func (cpu *CPU) tax(opDat) {
	cpu.x = cpu.a
	cpu.setZN(cpu.x)
}

// Increment X Register
//
// Adds one to the X register setting the zero and negative flags as appropriate.
func (cpu *CPU) inx(opDat) {
	cpu.x += 1
	cpu.setZN(cpu.x)
}

func (cpu *CPU) ora(opDat) {}
func (cpu *CPU) jam(opDat) {}
func (cpu *CPU) slo(opDat) {}
func (cpu *CPU) nop(opDat) {}
func (cpu *CPU) asl(opDat) {}
func (cpu *CPU) php(opDat) {

}
func (cpu *CPU) anc(opDat)  {}
func (cpu *CPU) bpl(opDat)  {}
func (cpu *CPU) clc(opDat)  {}
func (cpu *CPU) jsr(opDat)  {}
func (cpu *CPU) and(opDat)  {}
func (cpu *CPU) rla(opDat)  {}
func (cpu *CPU) bit(opDat)  {}
func (cpu *CPU) rol(opDat)  {}
func (cpu *CPU) plp(opDat)  {}
func (cpu *CPU) bmi(opDat)  {}
func (cpu *CPU) sec(opDat)  {}
func (cpu *CPU) rti(opDat)  {}
func (cpu *CPU) eor(opDat)  {}
func (cpu *CPU) sre(opDat)  {}
func (cpu *CPU) lsr(opDat)  {}
func (cpu *CPU) pha(opDat)  {}
func (cpu *CPU) alr(opDat)  {}
func (cpu *CPU) jmp(opDat)  {}
func (cpu *CPU) bvc(opDat)  {}
func (cpu *CPU) cli(opDat)  {}
func (cpu *CPU) rts(opDat)  {}
func (cpu *CPU) adc(opDat)  {}
func (cpu *CPU) rra(opDat)  {}
func (cpu *CPU) ror(opDat)  {}
func (cpu *CPU) pla(opDat)  {}
func (cpu *CPU) arr(opDat)  {}
func (cpu *CPU) bvs(opDat)  {}
func (cpu *CPU) sei(opDat)  {}
func (cpu *CPU) sta(opDat)  {}
func (cpu *CPU) sax(opDat)  {}
func (cpu *CPU) sty(opDat)  {}
func (cpu *CPU) stx(opDat)  {}
func (cpu *CPU) dey(opDat)  {}
func (cpu *CPU) txa(opDat)  {}
func (cpu *CPU) ane(opDat)  {}
func (cpu *CPU) bcc(opDat)  {}
func (cpu *CPU) sha(opDat)  {}
func (cpu *CPU) tya(opDat)  {}
func (cpu *CPU) txs(opDat)  {}
func (cpu *CPU) tas(opDat)  {}
func (cpu *CPU) shy(opDat)  {}
func (cpu *CPU) shx(opDat)  {}
func (cpu *CPU) ldy(opDat)  {}
func (cpu *CPU) ldx(opDat)  {}
func (cpu *CPU) lax(opDat)  {}
func (cpu *CPU) tay(opDat)  {}
func (cpu *CPU) lxa(opDat)  {}
func (cpu *CPU) bcs(opDat)  {}
func (cpu *CPU) clv(opDat)  {}
func (cpu *CPU) tsx(opDat)  {}
func (cpu *CPU) las(opDat)  {}
func (cpu *CPU) cpy(opDat)  {}
func (cpu *CPU) cmp(opDat)  {}
func (cpu *CPU) dcp(opDat)  {}
func (cpu *CPU) dec(opDat)  {}
func (cpu *CPU) iny(opDat)  {}
func (cpu *CPU) dex(opDat)  {}
func (cpu *CPU) sbx(opDat)  {}
func (cpu *CPU) bne(opDat)  {}
func (cpu *CPU) cld(opDat)  {}
func (cpu *CPU) cpx(opDat)  {}
func (cpu *CPU) sbc(opDat)  {}
func (cpu *CPU) isc(opDat)  {}
func (cpu *CPU) inc(opDat)  {}
func (cpu *CPU) usbc(opDat) {}
func (cpu *CPU) beq(opDat)  {}
func (cpu *CPU) sed(opDat)  {}

// 4 + 72 = 76
// 23 illegals

func (cpu *CPU) Hotloop(program []byte) {
	if len(program) > math.MaxUint16 {
		panic(fmt.Errorf("len of program %v greater than max %v", len(program), math.MaxUint16))
	}

	cpu.pc = 0
	// for !cpu.status.B {
	// 	op := program[cpu.pc]
	// 	cpu.pc += 1
	// 	switch op {
	// 	case 0x00:
	// 		cpu.brk()
	// 	case 0xa9:
	// 		// value := program[cpu.pc]
	// 		cpu.lda(opDat{addr: })
	// 		cpu.pc += 1
	// 	case 0xaa:
	// 		cpu.tax()
	// 	case 0xe8:
	// 		cpu.inx()
	// 	}

	// }
}
