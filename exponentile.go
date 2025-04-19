package main

import (
	crand "crypto/rand"
	"fmt"
	"math/rand/v2"
)

type Exponentile struct {
	Size  int
	Board []int

	CollapseTemp [][]*Move

	Rand *rand.Rand

	Printer BoardPrinter
}

type BoardPrinter interface {
	// Print must do any state copying before it returns
	// Exponentile will move on without us
	Print(*Exponentile)
}

type xConsoleBeardPrinter struct{}

func (xConsoleBeardPrinter xConsoleBeardPrinter) Print(et *Exponentile) {
	for y := 0; y < et.Size; y++ {
		for x := 0; x < et.Size; x++ {
			v := et.Board[(y*et.Size)+x]
			fmt.Printf("%6d,", v)
		}
		fmt.Println()
	}
}

var ConsoleBeardPrinter BoardPrinter = &xConsoleBeardPrinter{}

const DefaultExponentileSize = 8

// NewExponentile
// size <= 0 gets replaced with DefaultExponentileSize
func NewExponentile(size int) *Exponentile {
	if size <= 0 {
		size = DefaultExponentileSize
	}
	var seed [32]byte
	crand.Read(seed[:])
	return &Exponentile{
		Size:         size,
		Board:        make([]int, size*size),
		CollapseTemp: make([][]*Move, size*size),
		Rand:         rand.New(rand.NewChaCha8(seed)),
		Printer:      ConsoleBeardPrinter,
	}
}

type Point struct {
	X, Y int
}

type Move struct {
	// Swap tiles at (Xa,Ya) <-> (Xb, Yb),
	// OR, as a collapse (the range from xa,ya to xb,yb)
	Xa, Ya, Xb, Yb int

	// collides refers to other collapse-ranges that overlap
	collides []*Move

	// collapseTouch is a rough count from FindCollapses(false)
	collapseTouch int
}

func (m *Move) Length() int {
	if m.Xa == m.Xb {
		return iabs(m.Ya - m.Yb)
	}
	return iabs(m.Xa - m.Xb)
}

func (m *Move) EnumeratePoints() []Point {
	if m.Xa == m.Xb {
		ymin, ymax := iminmax(m.Ya, m.Yb)
		mlen := (ymax - ymin) + 1
		out := make([]Point, mlen)
		pos := 0
		for y := ymin; y <= ymax; y++ {
			out[pos].X = m.Xa
			out[pos].Y = y
		}
		return out
	} else {
		xmin, xmax := iminmax(m.Xa, m.Xb)
		mlen := (xmax - xmin) + 1
		out := make([]Point, mlen)
		pos := 0
		for x := xmin; x <= xmax; x++ {
			out[pos].X = x
			out[pos].Y = m.Ya
		}
		return out
	}
}

func (m *Move) Contains(x, y int) bool {
	if (m.Xa == x || m.Xb == x) && (m.Ya <= y && y <= m.Yb) {
		return true
	}
	if (m.Ya == y || m.Yb == y) && (m.Xa <= x && x <= m.Xb) {
		return true
	}
	return false
}

func (m *Move) ContainsMove(mov *Move) bool {
	return m.Contains(mov.Xa, mov.Ya) || m.Contains(mov.Xb, mov.Yb)
}

type Collapse struct {
	Ranges []*Move
}

func (cl *Collapse) Total() int {
	total := 0
	for _, m := range cl.Ranges {
		total += m.Length()
	}
	total -= (len(cl.Ranges) - 1)
	return total
}

var collapseMultiplierLut [12]int

func init() {
	collapseMultiplierLut[3] = 2
	for i := 4; i < len(collapseMultiplierLut); i++ {
		collapseMultiplierLut[i] = collapseMultiplierLut[i-1] * 2
	}
}

func (cl *Collapse) Multiplier() int {
	return collapseMultiplierLut[cl.Total()]
}

func (cl *Collapse) AnchorPoint() (int, int) {
	if len(cl.Ranges) == 0 {
		panic("no ranges for Collapse.AnchorPoint()")
	} else if len(cl.Ranges) == 1 {
		return cl.Ranges[0].Xa, cl.Ranges[0].Ya
	} else if len(cl.Ranges) == 2 {
		r1 := cl.Ranges[0]
		r2 := cl.Ranges[1]
		if r1.Xa == r1.Xb {
			// r1 is x-constant, use its x and r2's y
			return r1.Xa, r2.Ya
		} else {
			return r2.Xa, r1.Ya
		}
	} else {
		panic("too many ranges for Collapse.AnchorPoint()")
	}
}

func (et *Exponentile) trySwap(xa, ya, xb, yb int, possibleMoves []Move) []Move {
	olda := et.Board[(ya*et.Size)+xa]
	oldb := et.Board[(yb*et.Size)+xb]
	et.Board[(ya*et.Size)+xa] = oldb
	et.Board[(yb*et.Size)+xb] = olda
	tmoves := et.FindCollapses(false)
	touched := 0
	for _, m := range tmoves {
		touched += m.Length()
	}
	if touched > 0 {
		possibleMoves = append(possibleMoves, Move{
			Xa:            xa,
			Ya:            ya,
			Xb:            xb,
			Yb:            yb,
			collapseTouch: touched,
		})
	}

	// undo swap
	et.Board[(ya*et.Size)+xa] = olda
	et.Board[(yb*et.Size)+xb] = oldb
	return possibleMoves
}

func (et *Exponentile) FindMoves() []Move {
	var possibleMoves []Move
	for ya := 0; ya < et.Size; ya++ {
		for xa := 0; xa < et.Size; xa++ {
			// test x+1
			xb := xa + 1
			if xb < et.Size {
				yb := ya
				possibleMoves = et.trySwap(xa, ya, xb, yb, possibleMoves)
			}

			// test y+1
			yb := ya + 1
			if yb < et.Size {
				xb = xa
				possibleMoves = et.trySwap(xa, ya, xb, yb, possibleMoves)
			}
		}
	}
	return possibleMoves
}

func (et *Exponentile) clearCollapseTemp() {
	for y := 0; y < et.Size; y++ {
		for x := 0; x < et.Size; x++ {
			ct := et.CollapseTemp[(y*et.Size)+x]
			if ct != nil {
				et.CollapseTemp[(y*et.Size)+x] = ct[:0]
			}
		}
	}
}

func (et *Exponentile) setCollapseTemp(x, y int, mov *Move) {
	ct := et.CollapseTemp[(y*et.Size)+x]
	et.CollapseTemp[(y*et.Size)+x] = append(ct, mov)
}

// FindCollapses finds all runs >=3 long along any x or y line
// setTemp=false for testing a move, setTemp=true for applying a move
func (et *Exponentile) FindCollapses(setTemp bool) []*Move {
	if setTemp {
		et.clearCollapseTemp()
	}
	var moves []*Move
	for y := 0; y < et.Size; y++ {
		prev := -1
		runlen := 0
		startx := -1
		for x := 0; x < et.Size; x++ {
			cur := et.Board[(y*et.Size)+x]
			if cur == prev {
				runlen++
			} else {
				if runlen >= 3 {
					nmov := &Move{
						Xa: startx,
						Ya: y,
						Xb: x - 1,
						Yb: y,
					}
					if setTemp {
						for tx := nmov.Xa; tx <= nmov.Xb; tx++ {
							et.setCollapseTemp(tx, nmov.Ya, nmov)
						}
					}
					moves = append(moves, nmov)
				}
				runlen = 1
				prev = cur
				startx = x
			}
		}
		if runlen >= 3 {
			nmov := &Move{
				Xa: startx,
				Ya: y,
				Xb: et.Size - 1,
				Yb: y,
			}
			if setTemp {
				for tx := nmov.Xa; tx <= nmov.Xb; tx++ {
					et.setCollapseTemp(tx, nmov.Ya, nmov)
				}
			}
			moves = append(moves, nmov)
		}
	}
	for x := 0; x < et.Size; x++ {
		prev := -1
		runlen := 0
		starty := -1
		for y := 0; y < et.Size; y++ {
			cur := et.Board[(y*et.Size)+x]
			if cur == prev {
				runlen++
			} else {
				if runlen >= 3 {
					nmov := &Move{
						Xa: x,
						Ya: starty,
						Xb: x,
						Yb: y - 1,
					}
					if setTemp {
						for ty := nmov.Ya; ty <= nmov.Yb; ty++ {
							et.setCollapseTemp(nmov.Xa, ty, nmov)
						}
					}
					moves = append(moves, nmov)
				}
				runlen = 1
				prev = cur
				starty = y
			}
		}
		if runlen >= 3 {
			nmov := &Move{
				Xa: x,
				Ya: starty,
				Xb: x,
				Yb: et.Size - 1,
			}
			if setTemp {
				for ty := nmov.Ya; ty <= nmov.Yb; ty++ {
					et.setCollapseTemp(nmov.Xa, ty, nmov)
				}
			}
			moves = append(moves, nmov)
		}
	}
	return moves
}

func movesToCollapses(moves []*Move) []Collapse {
	out := make([]Collapse, 0, len(moves))
	for i, mov := range moves {
		var nc Collapse
		nc.Ranges = append(nc.Ranges, mov)
		for _, collideMov := range mov.collides {
			for j := i + 1; j < len(moves); j++ {
				if moves[j] == collideMov {
					moves[j] = nil
				}
			}
			nc.Ranges = append(nc.Ranges, collideMov)
		}
		out = append(out, nc)
	}
	return out
}

func (et *Exponentile) clearExcept(xa, ya, xb, yb, keepx, keepy int) {
	if xa == xb {
		for ty := ya; ty <= yb; ty++ {
			if xa == keepx && ty == keepy {
				continue
			}
			et.Board[(ty*et.Size)+xa] = 0
		}
	} else { // ya == yb
		for tx := xa; tx <= xb; tx++ {
			if tx == keepx && ya == keepy {
				continue
			}
			et.Board[(ya*et.Size)+tx] = 0
		}
	}
}

func (et *Exponentile) randTile() int {
	i := et.Rand.IntN(3)
	switch i {
	case 0:
		return 2
	case 1:
		return 4
	case 2:
		return 8
	default:
		return 2
	}
}

func (et *Exponentile) gravityDown() {
	// for each column, compact down, replace empties at top
	for x := 0; x < et.Size; x++ {
		yout := et.Size - 1
		yin := yout - 1
		for yout >= 0 {
			if et.Board[(yout*et.Size)+x] == 0 {
				if yin >= 0 {
					et.Board[(yout*et.Size)+x] = et.Board[(yin*et.Size)+x]
					et.Board[(yin*et.Size)+x] = 0
					yin--
				} else {
					et.Board[(yout*et.Size)+x] = et.randTile()
				}
			}
			yout--
			yin--
		}
	}
}

func (et *Exponentile) ApplyMove(mov Move) {
	collapseMoves := et.FindCollapses(true)
	collapses := movesToCollapses(collapseMoves)
	for _, collapse := range collapses {
		if len(collapse.Ranges) == 1 {
			tmov := collapse.Ranges[0]
			if tmov.Contains(mov.Xa, mov.Ya) {
				// pin result to where the move touched it
				et.clearExcept(tmov.Xa, tmov.Ya, tmov.Xb, tmov.Yb, mov.Xa, mov.Ya)
				et.Board[(mov.Ya*et.Size)+mov.Xa] *= collapse.Multiplier()
			} else if tmov.Contains(mov.Xb, mov.Yb) {
				// pin result to where the move touched it
				et.clearExcept(tmov.Xa, tmov.Ya, tmov.Xb, tmov.Yb, mov.Xb, mov.Yb)
				et.Board[(mov.Yb*et.Size)+mov.Xb] *= collapse.Multiplier()
			} else {
				// collapse to leftmost/topmost
				et.clearExcept(tmov.Xa, tmov.Ya, tmov.Xb, tmov.Yb, tmov.Xa, tmov.Ya)
				et.Board[(tmov.Ya*et.Size)+tmov.Xa] *= collapse.Multiplier()
			}
		} else if len(collapse.Ranges) == 2 {
			// collapse to intersection point
			ax, ay := collapse.AnchorPoint()
			for _, cr := range collapse.Ranges {
				et.clearExcept(cr.Xa, cr.Ya, cr.Xb, cr.Yb, ax, ay)
			}
			et.Board[(ay*et.Size)+ax] *= collapse.Multiplier()
		} else {
			panic(fmt.Sprintf("len(collapse.Ranges)=%d", len(collapse.Ranges)))
		}
	}
	et.gravityDown()
	et.Printer.Print(et)
	// TODO: post update

	for {
		// process chain reactions
		ncollapseMoves := et.FindCollapses(false)
		if len(ncollapseMoves) == 0 {
			break
		}
		ncollapses := movesToCollapses(ncollapseMoves)
		for _, collapse := range ncollapses {
			if len(collapse.Ranges) == 1 {
				tmov := collapse.Ranges[0]
				// collapse to leftmost/topmost
				et.clearExcept(tmov.Xa, tmov.Ya, tmov.Xb, tmov.Yb, tmov.Xa, tmov.Ya)
				et.Board[(tmov.Ya*et.Size)+tmov.Xa] *= collapse.Multiplier()
			} else if len(collapse.Ranges) == 2 {
				// collapse to intersection point
				ax, ay := collapse.AnchorPoint()
				for _, cr := range collapse.Ranges {
					et.clearExcept(cr.Xa, cr.Ya, cr.Xb, cr.Yb, ax, ay)
				}
				et.Board[(ay*et.Size)+ax] *= collapse.Multiplier()
			} else {
				panic(fmt.Sprintf("len(collapse.Ranges)=%d", len(collapse.Ranges)))
			}
		}
		et.gravityDown()
		// TODO: post update
	}
}

func iabs(a int) int {
	if a < 0 {
		return a * -1
	}
	return a
}

func iminmax(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}
