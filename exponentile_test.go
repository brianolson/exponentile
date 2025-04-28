package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type arrayIntNSource struct {
	pos    int
	values []int
}

func (ains *arrayIntNSource) IntN(N int) int {
	out := ains.values[ains.pos]
	ains.pos++
	if out >= N {
		panic("bad test IntN")
	}
	return out
}

func TestGravityDown(t *testing.T) {
	et := NewExponentile(4)
	et.Board = []int{
		1, 2, 3, 4,
		5, 6, 0, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}
	et.Rand = &arrayIntNSource{values: []int{0}}
	et.gravityDown()
	assert.Equal(t, []int{
		1, 2, 2, 4,
		5, 6, 3, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}, et.Board)

	et.Board = []int{
		1, 2, 3, 4,
		5, 6, 7, 8,
		1, 2, 0, 4,
		5, 6, 7, 8,
	}
	et.Rand = &arrayIntNSource{values: []int{0}}
	et.gravityDown()
	assert.Equal(t, []int{
		1, 2, 2, 4,
		5, 6, 3, 8,
		1, 2, 7, 4,
		5, 6, 7, 8,
	}, et.Board)

	et.Board = []int{
		1, 2, 3, 4,
		5, 6, 0, 8,
		1, 2, 0, 4,
		5, 6, 7, 8,
	}
	et.Rand = &arrayIntNSource{values: []int{0, 0}}
	et.gravityDown()
	assert.Equal(t, []int{
		1, 2, 2, 4,
		5, 6, 2, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}, et.Board)
}

func TestClearExcept(t *testing.T) {
	size := 4
	et := &Exponentile{
		Size: 4,
		Board: []int{
			1, 2, 3, 4,
			5, 6, 7, 8,
			1, 2, 3, 4,
			5, 6, 7, 8,
		},
		CollapseTemp: make([][]*Move, size*size),
		Rand:         &arrayIntNSource{values: []int{}},
		Printer:      ConsoleBoardPrinter,
	}
	et.Rand = &arrayIntNSource{values: []int{}}
	et.clearExcept(1, 1, 3, 1, 2, 1)
	assert.Equal(t, []int{
		1, 2, 3, 4,
		5, 0, 7, 0,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}, et.Board)

	et.Board = []int{
		1, 2, 3, 4,
		5, 6, 7, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}
	et.clearExcept(1, 1, 1, 3, 1, 2)
	assert.Equal(t, []int{
		1, 2, 3, 4,
		5, 0, 7, 8,
		1, 2, 3, 4,
		5, 0, 7, 8,
	}, et.Board)

	et.Board = []int{
		1, 2, 3, 4,
		5, 6, 7, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}
	et.clearExcept(1, 1, 3, 1, 1, 1)
	assert.Equal(t, []int{
		1, 2, 3, 4,
		5, 6, 0, 0,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}, et.Board)

	et.Board = []int{
		1, 2, 3, 4,
		5, 6, 7, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}
	et.clearExcept(0, 1, 2, 1, 0, 1)
	assert.Equal(t, []int{
		1, 2, 3, 4,
		5, 0, 0, 8,
		1, 2, 3, 4,
		5, 6, 7, 8,
	}, et.Board)
}
