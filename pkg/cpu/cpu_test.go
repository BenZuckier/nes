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

	})
}

func TestOpcodes(t *testing.T) {
	Convey("should test LDA and BRK", t, func() {
		cpu := CPU{}

		Convey("should load test value with LDA and BRK", func() {
			val := byte(0x69)
			cpu.Hotloop([]byte{0xa9, val, 0x00}) // LDA, val, BRK
			So(cpu.a, ShouldEqual, val)

			So(cpu.pc, ShouldEqual, 3)

		})

		Convey("test TAX and BRK", func() {
			val := byte(0x42)

			cpu.a = val
			cpu.Hotloop([]byte{0xaa, 0x00}) // TAX, BRK

			So(cpu.x, ShouldEqual, val)

			So(cpu.pc, ShouldEqual, 2)
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
		})
	})
}
