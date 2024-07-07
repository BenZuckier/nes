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
	memory  [0xFFFF + 1]byte
	opcodes map[byte]opcode
}

func newCPU() *CPU {
	cpu := &CPU{}
	cpu.initializeOpcodeTable()
	return cpu
}

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
	return uint16(cpu.pop()) | uint16(cpu.pop())<<8 // low | hi << 8 // order sensitive
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

// deferrableSetFn makes a cpu.set_α function able to be called with the defer keyword and still work as expected.
//
// It takes a setFn which accepts a byte and allows us to pass a POINTER to the byte
// so the function can be deferred and not make a closure around the value of "result" prematurely.
//
// E.g. if we pass a value and the first line of a function is `defer setZN(cpu.a)` before we do anything and then set `cpu.a = 0`
// the defer will have closed around the previous value of a and we won't properly set the zero flag.
// Instead now we pass the pointer &cpu.a and when it runs the defer it will get the updated value of cpu.a = 0.
//
// TBD if this is a bad idea. Could also genericize.
func deferrableSetFn(setFn func(byte)) func(*byte) {
	return func(v *byte) { setFn(*v) }
}

func (cpu *CPU) setB() {
	cpu.status.B = true
}

// brk - Force Interrupt
//
// The brk instruction forces the generation of an interrupt request.
// The program counter and processor status are pushed on the stack then the IRQ interrupt vector at 0xFFFE/F is loaded into the PC and the break flag in the status set to one.
func (cpu *CPU) brk(dat opDat) {
	cpu.push16(cpu.pc)
	cpu.setB()
	cpu.php(dat)
	// cpu.sei(dat) // TODO: ?
	cpu.pc = cpu.read16(0xFFFE)
}

// lda - Load Accumulator
//
// load the value into register A and set Z and N flags if value is 0 or negative respectively.
func (cpu *CPU) lda(dat opDat) {
	defer deferrableSetFn(cpu.setZN)(&cpu.a)
	cpu.a = cpu.read(dat.addr)
}

// tax - Transfer Accumulator to X
//
// Copies the current contents of the accumulator into the X register and sets the zero and negative flags as appropriate. (transfer a to x)
func (cpu *CPU) tax(opDat) {
	defer deferrableSetFn(cpu.setZN)(&cpu.x)
	// defer cpu.setZN(&cpu.x)
	cpu.x = cpu.a
}

// inx - Increment X Register
//
// Adds one to the X register setting the zero and negative flags as appropriate.
func (cpu *CPU) inx(opDat) {
	defer deferrableSetFn(cpu.setZN)(&cpu.x)
	cpu.x += 1
}

func (cpu *CPU) ora(opDat) {}
func (cpu *CPU) jam(opDat) {}
func (cpu *CPU) slo(opDat) {}
func (cpu *CPU) nop(opDat) {}
func (cpu *CPU) asl(opDat) {}

// php - Push Processor Status
//
// Pushes a copy of the status flags on to the stack.
func (cpu *CPU) php(opDat) {
	cpu.push(cpu.status.Get())
}
func (cpu *CPU) anc(opDat) {}
func (cpu *CPU) bpl(opDat) {}

// clc - Clear Carry Flag
//
// Set the carry flag to zero.
func (cpu *CPU) clc(opDat) {
	cpu.status.C = false
}

// cld - Clear Decimal Mode
//
// Sets the decimal mode flag to zero.
//
// NOTE: decimal mode will not be implemented
func (cpu *CPU) cld(opDat) {
	cpu.status.D = false
}

// cli - Clear Interrupt Disable
//
// Clears the interrupt disable flag allowing normal interrupt requests to be serviced.
func (cpu *CPU) cli(opDat) {
	cpu.status.I = false
}

// clv - Clear Overflow Flag
//
// Clears the overflow flag.
func (cpu *CPU) clv(opDat) {
	cpu.status.V = false
}

func (cpu *CPU) jsr(opDat) {}

// and - Logical AND
//
// A,Z,N = A&M
//
// A logical AND is performed, bit by bit, on the accumulator contents using the contents of a byte of memory.
func (cpu *CPU) and(dat opDat) {
	defer deferrableSetFn(cpu.setZN)(&cpu.a)
	cpu.a &= cpu.read(dat.addr)
}
func (cpu *CPU) rla(opDat) {}
func (cpu *CPU) bit(opDat) {}
func (cpu *CPU) rol(opDat) {}
func (cpu *CPU) plp(opDat) {}
func (cpu *CPU) bmi(opDat) {}
func (cpu *CPU) sec(opDat) {}
func (cpu *CPU) rti(opDat) {}
func (cpu *CPU) eor(opDat) {}
func (cpu *CPU) sre(opDat) {}
func (cpu *CPU) lsr(opDat) {}
func (cpu *CPU) pha(opDat) {}
func (cpu *CPU) alr(opDat) {}
func (cpu *CPU) jmp(opDat) {}
func (cpu *CPU) bvc(opDat) {}
func (cpu *CPU) rts(opDat) {}
func (cpu *CPU) adc(opDat) {}
func (cpu *CPU) rra(opDat) {}
func (cpu *CPU) ror(opDat) {}
func (cpu *CPU) pla(opDat) {}
func (cpu *CPU) arr(opDat) {}
func (cpu *CPU) bvs(opDat) {}

// SEI - Set Interrupt Disable
//
// I = 1
//
// Set the interrupt disable flag to one.
func (cpu *CPU) sei(opDat) {
	cpu.status.I = true
}
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

	copy(cpu.memory[:], program) // TODO: actually load

	for !cpu.status.B {
		op, nPC := cpu.opcodes[cpu.read(cpu.pc)], cpu.pc+1
		dat := opDat{mode: op.Mode}
		switch op.Mode {
		case implicit:
		case accumulator:
		case immediate:
			dat.addr = nPC
		case zeroPage:
			dat.addr = uint16(cpu.read(nPC))
		case zeroPageX:
			dat.addr = uint16(cpu.read(nPC) + cpu.x)
		case zeroPageY: // This mode can only be used with the LDX and STX instructions.
			dat.addr = uint16(cpu.read(nPC) + cpu.y)
		case relative:
			dat.addr = nPC + 1 + uint16(int8(cpu.read(nPC))) // the "byte" read is really a signed int8. Interpret as int8 then cast to unsigned 2s complement and account for the instruction length.
		case absolute:
			dat.addr = cpu.read16(nPC)
		case absoluteX:
			dat.addr = cpu.read16(nPC) + uint16(cpu.x)
		case absoluteY:
			dat.addr = cpu.read16(nPC) + uint16(cpu.y)
		// JMP is the only 6502 instruction to support indirection.
		// The instruction contains a 16 bit address which identifies the location of the least significant byte of another 16 bit memory address which is the real target of the instruction.
		case indirect:
			dat.addr = cpu.read16(cpu.read16(nPC)) // TODO: apparently there's a bug ?
		// Indexed indirect addressing is normally used in conjunction with a table of address held on zero page.
		// The address of the table is taken from the instruction and the X register added to it (with zero page wrap around) to give the location of the least significant byte of the target address.
		case indirectX:
			dat.addr = cpu.read16(
				uint16(cpu.read(nPC)) + uint16(cpu.x),
			)
		// Indirect indirect addressing is the most common indirection mode used on the 6502.
		// In instruction contains the zero page location of the least significant byte of 16 bit address. The Y register is dynamically added to this value to generated the actual target address for operation.
		case indirectY:
			dat.addr = cpu.read16(
				uint16(cpu.read(nPC)),
			) + uint16(cpu.y)
		default:
			panic(fmt.Errorf("unknown mode for op: %+v, cpu: %+v", op, cpu))
		}
		cpu.pc += op.Size
		// TODO: count cycles and page crossings
		dat.pc = cpu.pc

		op.Do(dat)

	}
	fmt.Printf("Time to take a BRK. bye :)\n")
}
