package regs

import (
	"fmt"
)

type R byte

const (
	R0 = R(0)
	R1 = R(1)
	R2 = R(2)
)

func (reg R) String() string {
	return fmt.Sprintf("r%d", reg)
}
