package translator

import (
	"fmt"
)

type ipGenerator struct {
	index uint32
}

func newIpGenerator() *ipGenerator {
	return &ipGenerator{index: uint32(0)}
}

func (g *ipGenerator) nextIp() string {
	g.index = g.index + 1
	return fmt.Sprintf("240.0.0.%d", g.index)
}
