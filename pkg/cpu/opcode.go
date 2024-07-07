package cpu

// generated with ChatGPT and https://www.masswerk.at/6502/6502_instruction_set.html

type opDat struct {
	addr uint16
	pc   uint16
	mode int
}

// opcode represents a 6502 opcode with its metadata.
type opcode struct {
	Name   string
	Mode   int
	Size   uint16
	Cycles int
	Do     func(opDat) // FbyF is 256 total codes, with 105 illegal opcodes, giving 151 legal codes.
}

// InitializeOpcodeTable initializes the CPU's opcode table.
func (cpu *CPU) initializeOpcodeTable() {
	cpu.opcodes = map[byte]opcode{ // could just be a slice but whatever
		0x00: {Name: "BRK", Mode: implicit, Size: 2, Cycles: 7, Do: cpu.brk}, // https://www.nesdev.org/the%20'B'%20flag%20&%20BRK%20instruction.txt
		0x01: {Name: "ORA", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.ora},
		0x02: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x03: {Name: "SLO", Mode: indirectX, Size: 2, Cycles: 8, Do: cpu.slo}, // illegal
		0x04: {Name: "NOP", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.nop},  // illegal
		0x05: {Name: "ORA", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.ora},
		0x06: {Name: "ASL", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.asl},
		0x07: {Name: "SLO", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.slo}, // illegal
		0x08: {Name: "PHP", Mode: implicit, Size: 1, Cycles: 3, Do: cpu.php},
		0x09: {Name: "ORA", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.ora},
		0x0A: {Name: "ASL", Mode: accumulator, Size: 1, Cycles: 2, Do: cpu.asl},
		0x0B: {Name: "ANC", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.anc}, // illegal
		0x0C: {Name: "NOP", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.nop},  // illegal
		0x0D: {Name: "ORA", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.ora},
		0x0E: {Name: "ASL", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.asl},
		0x0F: {Name: "SLO", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.slo}, // illegal

		0x10: {Name: "BPL", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bpl},
		0x11: {Name: "ORA", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.ora},
		0x12: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x13: {Name: "SLO", Mode: indirectY, Size: 2, Cycles: 8, Do: cpu.slo}, // illegal
		0x14: {Name: "NOP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.nop}, // illegal
		0x15: {Name: "ORA", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.ora},
		0x16: {Name: "ASL", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.asl},
		0x17: {Name: "SLO", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.slo}, // illegal
		0x18: {Name: "CLC", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.clc},
		0x19: {Name: "ORA", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.ora},
		0x1A: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},  // illegal
		0x1B: {Name: "SLO", Mode: absoluteY, Size: 3, Cycles: 7, Do: cpu.slo}, // illegal
		0x1C: {Name: "NOP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.nop}, // illegal
		0x1D: {Name: "ORA", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.ora},
		0x1E: {Name: "ASL", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.asl},
		0x1F: {Name: "SLO", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.slo}, // illegal

		0x20: {Name: "JSR", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.jsr},
		0x21: {Name: "AND", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.and},
		0x22: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x23: {Name: "RLA", Mode: indirectX, Size: 2, Cycles: 8, Do: cpu.rla}, // illegal
		0x24: {Name: "BIT", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.bit},
		0x25: {Name: "AND", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.and},
		0x26: {Name: "ROL", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.rol},
		0x27: {Name: "RLA", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.rla}, // illegal
		0x28: {Name: "PLP", Mode: implicit, Size: 1, Cycles: 4, Do: cpu.plp},
		0x29: {Name: "AND", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.and},
		0x2A: {Name: "ROL", Mode: accumulator, Size: 1, Cycles: 2, Do: cpu.rol},
		0x2B: {Name: "ANC", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.anc}, // illegal
		0x2C: {Name: "BIT", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.bit},
		0x2D: {Name: "AND", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.and},
		0x2E: {Name: "ROL", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.rol},
		0x2F: {Name: "RLA", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.rla}, // illegal

		0x30: {Name: "BMI", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bmi},
		0x31: {Name: "AND", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.and},
		0x32: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x33: {Name: "RLA", Mode: indirectY, Size: 2, Cycles: 8, Do: cpu.rla}, // illegal
		0x34: {Name: "NOP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.nop}, // illegal
		0x35: {Name: "AND", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.and},
		0x36: {Name: "ROL", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.rol},
		0x37: {Name: "RLA", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.rla}, // illegal
		0x38: {Name: "SEC", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.sec},
		0x39: {Name: "AND", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.and},
		0x3A: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},  // illegal
		0x3B: {Name: "RLA", Mode: absoluteY, Size: 3, Cycles: 7, Do: cpu.rla}, // illegal
		0x3C: {Name: "NOP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.nop}, // illegal
		0x3D: {Name: "AND", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.and},
		0x3E: {Name: "ROL", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.rol},
		0x3F: {Name: "RLA", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.rla}, // illegal

		0x40: {Name: "RTI", Mode: implicit, Size: 1, Cycles: 6, Do: cpu.rti},
		0x41: {Name: "EOR", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.eor},
		0x42: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x43: {Name: "SRE", Mode: indirectX, Size: 2, Cycles: 8, Do: cpu.sre}, // illegal
		0x44: {Name: "NOP", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.nop},  // illegal
		0x45: {Name: "EOR", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.eor},
		0x46: {Name: "LSR", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.lsr},
		0x47: {Name: "SRE", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.sre}, // illegal
		0x48: {Name: "PHA", Mode: implicit, Size: 1, Cycles: 3, Do: cpu.pha},
		0x49: {Name: "EOR", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.eor},
		0x4A: {Name: "LSR", Mode: accumulator, Size: 1, Cycles: 2, Do: cpu.lsr},
		0x4B: {Name: "ALR", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.alr}, // illegal
		0x4C: {Name: "JMP", Mode: absolute, Size: 3, Cycles: 3, Do: cpu.jmp},
		0x4D: {Name: "EOR", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.eor},
		0x4E: {Name: "LSR", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.lsr},
		0x4F: {Name: "SRE", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.sre}, // illegal

		0x50: {Name: "BVC", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bvc},
		0x51: {Name: "EOR", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.eor},
		0x52: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x53: {Name: "SRE", Mode: indirectY, Size: 2, Cycles: 8, Do: cpu.sre}, // illegal
		0x54: {Name: "NOP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.nop}, // illegal
		0x55: {Name: "EOR", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.eor},
		0x56: {Name: "LSR", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.lsr},
		0x57: {Name: "SRE", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.sre}, // illegal
		0x58: {Name: "CLI", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.cli},
		0x59: {Name: "EOR", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.eor},
		0x5A: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},  // illegal
		0x5B: {Name: "SRE", Mode: absoluteY, Size: 3, Cycles: 7, Do: cpu.sre}, // illegal
		0x5C: {Name: "NOP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.nop}, // illegal
		0x5D: {Name: "EOR", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.eor},
		0x5E: {Name: "LSR", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.lsr},
		0x5F: {Name: "SRE", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.sre}, // illegal

		0x60: {Name: "RTS", Mode: implicit, Size: 1, Cycles: 6, Do: cpu.rts},
		0x61: {Name: "ADC", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.adc},
		0x62: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x63: {Name: "RRA", Mode: indirectX, Size: 2, Cycles: 8, Do: cpu.rra}, // illegal
		0x64: {Name: "NOP", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.nop},  // illegal
		0x65: {Name: "ADC", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.adc},
		0x66: {Name: "ROR", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.ror},
		0x67: {Name: "RRA", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.rra}, // illegal
		0x68: {Name: "PLA", Mode: implicit, Size: 1, Cycles: 4, Do: cpu.pla},
		0x69: {Name: "ADC", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.adc},
		0x6A: {Name: "ROR", Mode: accumulator, Size: 1, Cycles: 2, Do: cpu.ror},
		0x6B: {Name: "ARR", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.arr}, // illegal
		0x6C: {Name: "JMP", Mode: indirect, Size: 3, Cycles: 5, Do: cpu.jmp},
		0x6D: {Name: "ADC", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.adc},
		0x6E: {Name: "ROR", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.ror},
		0x6F: {Name: "RRA", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.rra}, // illegal

		0x70: {Name: "BVS", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bvs},
		0x71: {Name: "ADC", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.adc},
		0x72: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x73: {Name: "RRA", Mode: indirectY, Size: 2, Cycles: 8, Do: cpu.rra}, // illegal
		0x74: {Name: "NOP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.nop}, // illegal
		0x75: {Name: "ADC", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.adc},
		0x76: {Name: "ROR", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.ror},
		0x77: {Name: "RRA", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.rra}, // illegal
		0x78: {Name: "SEI", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.sei},
		0x79: {Name: "ADC", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.adc},
		0x7A: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},  // illegal
		0x7B: {Name: "RRA", Mode: absoluteY, Size: 3, Cycles: 7, Do: cpu.rra}, // illegal
		0x7C: {Name: "NOP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.nop}, // illegal
		0x7D: {Name: "ADC", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.adc},
		0x7E: {Name: "ROR", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.ror},
		0x7F: {Name: "RRA", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.rra}, // illegal

		0x80: {Name: "NOP", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.nop}, // illegal
		0x81: {Name: "STA", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.sta},
		0x82: {Name: "NOP", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.nop}, // illegal
		0x83: {Name: "SAX", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.sax}, // illegal
		0x84: {Name: "STY", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.sty},
		0x85: {Name: "STA", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.sta},
		0x86: {Name: "STX", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.stx},
		0x87: {Name: "SAX", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.sax}, // illegal
		0x88: {Name: "DEY", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.dey},
		0x89: {Name: "NOP", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.nop}, // illegal
		0x8A: {Name: "TXA", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.txa},
		0x8B: {Name: "ANE", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.ane}, // illegal
		0x8C: {Name: "STY", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.sty},
		0x8D: {Name: "STA", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.sta},
		0x8E: {Name: "STX", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.stx},
		0x8F: {Name: "SAX", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.sax}, // illegal

		0x90: {Name: "BCC", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bcc},
		0x91: {Name: "STA", Mode: indirectY, Size: 2, Cycles: 6, Do: cpu.sta},
		0x92: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0x93: {Name: "SHA", Mode: indirectY, Size: 2, Cycles: 6, Do: cpu.sha}, // illegal
		0x94: {Name: "STY", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.sty},
		0x95: {Name: "STA", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.sta},
		0x96: {Name: "STX", Mode: zeroPageY, Size: 2, Cycles: 4, Do: cpu.stx},
		0x97: {Name: "SAX", Mode: zeroPageY, Size: 2, Cycles: 4, Do: cpu.sax}, // illegal
		0x98: {Name: "TYA", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.tya},
		0x99: {Name: "STA", Mode: absoluteY, Size: 3, Cycles: 5, Do: cpu.sta},
		0x9A: {Name: "TXS", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.txs},
		0x9B: {Name: "TAS", Mode: absoluteY, Size: 3, Cycles: 5, Do: cpu.tas}, // illegal
		0x9C: {Name: "SHY", Mode: absoluteX, Size: 3, Cycles: 5, Do: cpu.shy}, // illegal
		0x9D: {Name: "STA", Mode: absoluteX, Size: 3, Cycles: 5, Do: cpu.sta},
		0x9E: {Name: "SHX", Mode: absoluteY, Size: 3, Cycles: 5, Do: cpu.shx},
		0x9F: {Name: "SHA", Mode: absoluteY, Size: 3, Cycles: 5, Do: cpu.sha}, // illegal

		0xA0: {Name: "LDY", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.ldy},
		0xA1: {Name: "LDA", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.lda},
		0xA2: {Name: "LDX", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.ldx},
		0xA3: {Name: "LAX", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.lax}, // illegal
		0xA4: {Name: "LDY", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.ldy},
		0xA5: {Name: "LDA", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.lda},
		0xA6: {Name: "LDX", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.ldx},
		0xA7: {Name: "LAX", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.lax}, // illegal
		0xA8: {Name: "TAY", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.tay},
		0xA9: {Name: "LDA", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.lda},
		0xAA: {Name: "TAX", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.tax},
		0xAB: {Name: "LXA", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.lxa}, // illegal
		0xAC: {Name: "LDY", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.ldy},
		0xAD: {Name: "LDA", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.lda},
		0xAE: {Name: "LDX", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.ldx},
		0xAF: {Name: "LAX", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.lax}, // illegal

		0xB0: {Name: "BCS", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bcs},
		0xB1: {Name: "LDA", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.lda},
		0xB2: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0xB3: {Name: "LAX", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.lax}, // illegal
		0xB4: {Name: "LDY", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.ldy},
		0xB5: {Name: "LDA", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.lda},
		0xB6: {Name: "LDX", Mode: zeroPageY, Size: 2, Cycles: 4, Do: cpu.ldx},
		0xB7: {Name: "LAX", Mode: zeroPageY, Size: 2, Cycles: 4, Do: cpu.lax}, // illegal
		0xB8: {Name: "CLV", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.clv},
		0xB9: {Name: "LDA", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.lda},
		0xBA: {Name: "TSX", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.tsx},
		0xBB: {Name: "LAS", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.las}, // illegal
		0xBC: {Name: "LDY", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.ldy},
		0xBD: {Name: "LDA", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.lda},
		0xBE: {Name: "LDX", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.ldx},
		0xBF: {Name: "LAX", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.lax}, // illegal

		0xC0: {Name: "CPY", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.cpy},
		0xC1: {Name: "CMP", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.cmp},
		0xC2: {Name: "NOP", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.nop}, // illegal
		0xC3: {Name: "DCP", Mode: indirectX, Size: 2, Cycles: 8, Do: cpu.dcp}, // illegal
		0xC4: {Name: "CPY", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.cpy},
		0xC5: {Name: "CMP", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.cmp},
		0xC6: {Name: "DEC", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.dec},
		0xC7: {Name: "DCP", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.dcp}, // illegal
		0xC8: {Name: "INY", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.iny},
		0xC9: {Name: "CMP", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.cmp},
		0xCA: {Name: "DEX", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.dex},
		0xCB: {Name: "SBX", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.sbx}, // illegal
		0xCC: {Name: "CPY", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.cpy},
		0xCD: {Name: "CMP", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.cmp},
		0xCE: {Name: "DEC", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.dec},
		0xCF: {Name: "DCP", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.dcp}, // illegal

		0xD0: {Name: "BNE", Mode: relative, Size: 2, Cycles: 2, Do: cpu.bne},
		0xD1: {Name: "CMP", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.cmp},
		0xD2: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0xD3: {Name: "DCP", Mode: indirectY, Size: 2, Cycles: 8, Do: cpu.dcp}, // illegal
		0xD4: {Name: "NOP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.nop}, // illegal
		0xD5: {Name: "CMP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.cmp},
		0xD6: {Name: "DEC", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.dec},
		0xD7: {Name: "DCP", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.dcp}, // illegal
		0xD8: {Name: "CLD", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.cld},
		0xD9: {Name: "CMP", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.cmp},
		0xDA: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},  // illegal
		0xDB: {Name: "DCP", Mode: absoluteY, Size: 3, Cycles: 7, Do: cpu.dcp}, // illegal
		0xDC: {Name: "NOP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.nop}, // illegal
		0xDD: {Name: "CMP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.cmp},
		0xDE: {Name: "DEC", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.dec},
		0xDF: {Name: "DCP", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.dcp}, // illegal

		0xE0: {Name: "CPX", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.cpx},
		0xE1: {Name: "SBC", Mode: indirectX, Size: 2, Cycles: 6, Do: cpu.sbc},
		0xE2: {Name: "NOP", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.nop}, // illegal
		0xE3: {Name: "ISC", Mode: indirectX, Size: 2, Cycles: 8, Do: cpu.isc}, // illegal
		0xE4: {Name: "CPX", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.cpx},
		0xE5: {Name: "SBC", Mode: zeroPage, Size: 2, Cycles: 3, Do: cpu.sbc},
		0xE6: {Name: "INC", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.inc},
		0xE7: {Name: "ISC", Mode: zeroPage, Size: 2, Cycles: 5, Do: cpu.isc}, // illegal
		0xE8: {Name: "INX", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.inx},
		0xE9: {Name: "SBC", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.sbc}, // illegal
		0xEA: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},
		0xEB: {Name: "USBC", Mode: immediate, Size: 2, Cycles: 2, Do: cpu.usbc}, // illegal
		0xEC: {Name: "CPX", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.cpx},
		0xED: {Name: "SBC", Mode: absolute, Size: 3, Cycles: 4, Do: cpu.sbc},
		0xEE: {Name: "INC", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.inc},
		0xEF: {Name: "ISC", Mode: absolute, Size: 3, Cycles: 6, Do: cpu.isc}, // illegal

		0xF0: {Name: "BEQ", Mode: relative, Size: 2, Cycles: 2, Do: cpu.beq},
		0xF1: {Name: "SBC", Mode: indirectY, Size: 2, Cycles: 5, Do: cpu.sbc},
		0xF2: {Name: "JAM", Mode: implicit, Size: 1, Cycles: 0, Do: cpu.jam},  // illegal
		0xF3: {Name: "ISC", Mode: indirectY, Size: 2, Cycles: 8, Do: cpu.isc}, // illegal
		0xF4: {Name: "NOP", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.nop}, // illegal
		0xF5: {Name: "SBC", Mode: zeroPageX, Size: 2, Cycles: 4, Do: cpu.sbc},
		0xF6: {Name: "INC", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.inc},
		0xF7: {Name: "ISC", Mode: zeroPageX, Size: 2, Cycles: 6, Do: cpu.isc}, // illegal
		0xF8: {Name: "SED", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.sed},
		0xF9: {Name: "SBC", Mode: absoluteY, Size: 3, Cycles: 4, Do: cpu.sbc},
		0xFA: {Name: "NOP", Mode: implicit, Size: 1, Cycles: 2, Do: cpu.nop},  // illegal
		0xFB: {Name: "ISC", Mode: absoluteY, Size: 3, Cycles: 7, Do: cpu.isc}, // illegal
		0xFC: {Name: "NOP", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.nop}, // illegal
		0xFD: {Name: "SBC", Mode: absoluteX, Size: 3, Cycles: 4, Do: cpu.sbc},
		0xFE: {Name: "INC", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.inc},
		0xFF: {Name: "ISC", Mode: absoluteX, Size: 3, Cycles: 7, Do: cpu.isc}, // illegal
	}
}
