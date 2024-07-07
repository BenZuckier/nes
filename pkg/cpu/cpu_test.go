package cpu

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBitBang(t *testing.T) {
	Convey("Bit bangs", t, func() {
		c := CPU{}
		// c := newCPU()

		So(posZ, ShouldEqual, 0x02)
		So(^posZ, ShouldEqual, 0xfd)

		Convey("with status all ones should set and unset zero flag", func() {
			// set to 11..11
			c.status.Set(0xff)

			fmt.Printf("status %v", c.status)

			c.setZ(0) // set the zero flag
			So(c.status.Get(), ShouldEqual, 0xff)

			c.setZ(1)                                    // should unset the zero flag
			So(c.status.Get(), ShouldEqual, 0b1111_1101) // 0xfd

			c.setZ(0) // set the zero flag
			So(c.status.Get(), ShouldEqual, 0xff)
		})

		Convey("with status all zeroes should set and unset zero flag", func() {
			c.status.Set(0x00) // all zeroes

			c.setZ(1)                             // unset zero flag
			So(c.status.Get(), ShouldEqual, 0x00) // 0xfd

			c.setZ(0)                                    // set zero flag
			So(c.status.Get(), ShouldEqual, 0b0000_0010) // zero flag set

			c.setZ(1)                             // unset zero flag
			So(c.status.Get(), ShouldEqual, 0x00) // 0xfd

		})

		Convey("only affects the zero flag", func() {
			initialStatus := byte(0xfd)
			c.status.Set(initialStatus)

			c.setZ(1) // unset zero flag
			fmt.Printf("%b\n", c.status.Get())
			So(c.status.Get()&^posZ, ShouldEqual, initialStatus)
			c.setZ(0) // set zero flag
			fmt.Printf("%b\n", c.status.Get())
			So(c.status.Get()&^posZ, ShouldEqual, initialStatus)
		})

		Convey("relative mode adds signed int to uint?", func() {
			// when checking my work against some random on the internet i saw their calculation for relative mode was weird and complex whereas mine was simple.
			// This tests checks every possible byte value against my impl and their impl and shows they're the same.
			myFunc := func(num byte) uint16 { return 2 + uint16(int8(num)) }
			internetFunc := func(num byte) uint16 {
				offset := uint16(num)
				if offset < 0x80 {
					return 2 + offset
				} else {
					return 2 + offset - 0x100
				}
			}
			for i := 0; i < 256; i++ {
				num := byte(i)
				So(myFunc(num), ShouldEqual, internetFunc(num))
			}

			// more proof that casting a byte to int8 and back to uint16 gives the proper value
			neg1 := byte(0xff)
			x := int8(neg1)
			ux := uint16(x)

			Printf("\nneg1: %0x, x: %0x, ux: %0x\n", neg1, x, ux)

			So(int(x), ShouldEqual, -1)
			So(ux, ShouldEqual, 0xffff)

		})

	})
}

func TestOpcodes(t *testing.T) {
	Convey("should test LDA and BRK", t, func() {
		cpu := newCPU()
		cpu.write16(0xFFFE, 0x1234) // put 1234 at the IRQ interrupt vector

		Convey("should load test value with LDA and BRK", func() {
			val := byte(0x69)

			cpu.Hotloop([]byte{0xa9, val, 0x00, 0x00}) // LDA, val, BRK, ignored brk value
			So(cpu.a, ShouldEqual, val)

			// pc gets set to the val at 0xFFFE after brk (IRQ interrupt vector)
			So(cpu.pc, ShouldEqual, 0x1234)

			oldStatus, oldPC := cpu.pop(), cpu.pop16() // get the status and program counter from the stack
			So(oldStatus, ShouldEqual, 0x00|posB)
			So(oldPC, ShouldEqual, 4)

		})

		Convey("test TAX and BRK", func() {
			val := byte(0x42)

			cpu.a = val
			cpu.Hotloop([]byte{0xaa, 0x00, 0x00}) // TAX, BRK, ignored brk val

			So(cpu.x, ShouldEqual, val)

			So(cpu.pc, ShouldEqual, 0x1234)

			oldStatus, oldPC := cpu.pop(), cpu.pop16() // get the status and program counter from the stack
			So(oldStatus, ShouldEqual, 0x00|posB)
			So(oldPC, ShouldEqual, 3)
		})

		Convey("add one to x inx", func() {

			Convey("test random value is incremented", func() {
				val := byte(0x69)
				cpu.x = val

				cpu.Hotloop([]byte{0xe8, 0x00})

				So(cpu.x, ShouldEqual, val+1)
				So(cpu.status.Get()&posZ, ShouldEqual, 0)
				So(cpu.status.Get()&posN, ShouldEqual, 0)
			})

			Convey("test neg 1 to zero", func() {
				val := byte(0xff) // -1
				cpu.x = val
				cpu.Hotloop([]byte{0xe8, 0x00})

				So(cpu.x, ShouldEqual, val+1)
				So(cpu.status.Get()&posZ, ShouldNotEqual, 0)
				So(cpu.status.Get()&posN, ShouldEqual, 0)
			})

		})

		Convey("simple programs", func() {

			Convey("p1", func() {
				cpu.Hotloop([]byte{0xa9, 0xc0, 0xaa, 0xe8, 0x00})
				So(cpu.x, ShouldEqual, 0xc1)

				oldStatus, oldPC := cpu.pop(), cpu.pop16() // get the status and program counter from the stack
				So(oldStatus, ShouldEqual, 0x00|posB|posN)
				So(oldPC, ShouldEqual, 6)
			})

			Convey("p2", func() {
				cpu.x = 0xff
				cpu.Hotloop([]byte{0xe8, 0xe8, 0x00})
				So(cpu.x, ShouldEqual, 1)
			})

		})

	})

}

func TestMemory(t *testing.T) {
	Convey("should test memory", t, func() {
		cpu := CPU{}

		Convey("test read and write", func() {
			// write 16B 0x8000 to addr 0x9000, little endian so it's 0x0080 in memory
			pos := uint16(0x9000)
			expected := uint16(0x8000)
			cpu.memory[pos] = 0x00
			cpu.memory[pos+1] = 0x80

			dat := cpu.read16(pos)
			fmt.Printf("dat is %04x", dat)
			So(dat, ShouldEqual, expected)

			// test writing back zero then 0x8000 again
			cpu.write16(pos, 0)
			So(cpu.read16(pos), ShouldEqual, 0)
			cpu.write16(pos, expected)
			So(cpu.read16(pos), ShouldEqual, expected)

			So(true, ShouldBeFalse) // fail
		})
	})
}
