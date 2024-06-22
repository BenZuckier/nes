package cpu

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBitBang(t *testing.T) {
	Convey("Bit bangs", t, func() {
		c := CPU{}

		So(z, ShouldEqual, 0x02)
		So(^z, ShouldEqual, 0xfd)

		Convey("with status all ones should set and unset zero flag", func() {
			// set to 11..11
			c.status = 0xff

			c.setZ(0) // set the zero flag
			So(c.status, ShouldEqual, 0xff)

			c.setZ(1)                              // should unset the zero flag
			So(c.status, ShouldEqual, 0b1111_1101) // 0xfd

			c.setZ(0) // set the zero flag
			So(c.status, ShouldEqual, 0xff)
		})

		Convey("with status all zeroes should set and unset zero flag", func() {
			c.status = 0x00 // all zeroes

			c.setZ(1)                       // unset zero flag
			So(c.status, ShouldEqual, 0x00) // 0xfd

			c.setZ(0)                              // set zero flag
			So(c.status, ShouldEqual, 0b0000_0010) // zero flag set

			c.setZ(1)                       // unset zero flag
			So(c.status, ShouldEqual, 0x00) // 0xfd

		})

		Convey("only affects the zero flag", func() {
			initialStatus := byte(0xfd)
			c.status = initialStatus

			c.setZ(1) // unset zero flag
			fmt.Printf("%b\n", c.status)
			So(c.status&^z, ShouldEqual, initialStatus)
			c.setZ(0) // set zero flag
			fmt.Printf("%b\n", c.status)
			So(c.status&^z, ShouldEqual, initialStatus)
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
				So(cpu.status&z, ShouldEqual, 0)
				So(cpu.status&n, ShouldEqual, 0)
			})

			Convey("test neg 1 to zero", func() {
				val := byte(0xff) // -1
				cpu.x = val
				cpu.Hotloop([]byte{0xe8, 0x00})

				So(cpu.x, ShouldEqual, val+1)
				So(cpu.status&z, ShouldNotEqual, 0)
				So(cpu.status&n, ShouldEqual, 0)
			})

		})
	})

}
