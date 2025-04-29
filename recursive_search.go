package main

import (
	crand "crypto/rand"
	"math/rand/v2"
)

type RecursiveSearch struct {
	Depth int
	// TODO: Oversampling int // how many times to retry each move to average out random effects

	boardStack []Exponentile

	rng *rand.Rand
}

func NewRecursiveSearch(depth int) *RecursiveSearch {
	var seed [32]byte
	_, _ = crand.Read(seed[:])
	rng := rand.New(rand.NewChaCha8(seed))
	return &RecursiveSearch{
		Depth: depth,
		rng:   rng,
	}
}

func (rs *RecursiveSearch) NextMove(et *Exponentile) (bool, Move) {
	if len(rs.boardStack) < rs.Depth {
		rs.boardStack = make([]Exponentile, rs.Depth)
		for depth := range rs.boardStack {
			rs.boardStack[depth].alloc(et.Size)
		}
	}
	ok, bestMove, _ := rs.innerNextMove(et, 0)
	return ok, bestMove
}
func (rs *RecursiveSearch) innerNextMove(et *Exponentile, depth int) (ok bool, bestMove Move, score int) {
	moves := et.FindMoves()
	if len(moves) == 0 {
		return false, Move{}, -1
	}
	bestScore := -1
	bestMi := -1
	anyOk := false
	for mi, tm := range moves {
		rs.boardStack[depth].Copy(et)
		rs.boardStack[depth].Rand = rs.rng
		rs.boardStack[depth].ApplyMove(tm)
		var tmScore int
		if depth < rs.Depth-1 {
			ok, _, subScore := rs.innerNextMove(&rs.boardStack[depth], depth+1)
			if ok {
				tmScore = subScore
			} else {
				// there were no further moves possible, this may be the last move, just use whatever score ApplyMove found
				tmScore = rs.boardStack[depth].Score
			}
		} else {
			// just use whatever score ApplyMove found
			tmScore = rs.boardStack[depth].Score
		}
		if !anyOk {
			anyOk = true
			bestMi = mi
			bestScore = tmScore
		} else if tmScore > bestScore {
			bestMi = mi
			bestScore = tmScore
		}
	}
	if anyOk {
		return true, moves[bestMi], bestScore
	}
	return false, Move{}, -1
}
