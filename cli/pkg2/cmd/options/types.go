package options

import (
	"context"
)

type Options struct {
	Top Top
}

type Top struct {
	Ctx         context.Context
	Interactive bool
}
