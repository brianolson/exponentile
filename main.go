package main

import (
	crand "crypto/rand"
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"
)

type Strategy interface {
	NextMove(et *Exponentile) (bool, Move)
}

type RandomStrategy struct {
	rng *rand.Rand
}

func NewRandomStrategy() *RandomStrategy {
	var seed [32]byte
	_, _ = crand.Read(seed[:])
	rng := rand.New(rand.NewChaCha8(seed))
	return &RandomStrategy{
		rng: rng,
	}
}

func (rs *RandomStrategy) NextMove(et *Exponentile) (bool, Move) {
	moves := et.FindMoves()
	if len(moves) == 0 {
		return false, Move{}
	}
	nextm := moves[rs.rng.IntN(len(moves))]
	return true, nextm
}

type Result struct {
	Dt    time.Duration
	Score int
	Steps int
	I     int
}

func sourceThread(wg *sync.WaitGroup, source chan<- int, N int) {
	defer close(source)
	if wg != nil {
		defer wg.Done()
	}
	for i := 0; i < N; i++ {
		source <- i
	}
}

func testhread(wg *sync.WaitGroup, source <-chan int, results chan<- Result, threadNo int) {
	defer wg.Done()
	strat := NewRandomStrategy()
	et := NewExponentile(8)
	stucky := 0
	for i := range source {
	retry:
		et.randomFill()
		et.Score = 0
		start := time.Now()
		steps := 0
		for {
			ok, nextm := strat.NextMove(et)
			if !ok {
				if steps == 0 {
					fmt.Printf("t[%d] no step 1\n", threadNo)
					et.Printer.Print(et)
					stucky++
					if stucky >= 10 {
						fmt.Printf("t[%d] stuck!\n", threadNo)
						return
					}
					goto retry
				} else {
					stucky = 0
				}
				break
			}
			et.ApplyMove(nextm)
			steps++
		}
		dt := time.Since(start)
		results <- Result{
			Dt:    dt,
			Score: et.Score,
			Steps: steps,
			I:     i,
		}
	}
}

func main() {
	nthreads := runtime.NumCPU() / 2
	if nthreads < 1 {
		nthreads = 1
	}

	ntests := 1000

	source := make(chan int, nthreads*10)
	results := make(chan Result, nthreads*10)

	allstart := time.Now()

	var wg sync.WaitGroup

	//wg.Add(1)
	go sourceThread(nil, source, ntests)

	wg.Add(nthreads)
	for i := 0; i < nthreads; i++ {
		go testhread(&wg, source, results, i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	allScores := make([]int, 0, ntests)
	scoreSum := 0
	scoreMin := 0
	scoreMax := 0
	first := true
	for result := range results {
		allScores = append(allScores, result.Score)
		scoreSum += result.Score
		if first {
			scoreMin = result.Score
			scoreMax = result.Score
			first = false
		} else {
			if result.Score < scoreMin {
				scoreMin = result.Score
			}
			if result.Score > scoreMax {
				scoreMax = result.Score
			}
		}
		fmt.Printf("[%5d] %d (%d steps) (%s)\n", result.I, result.Score, result.Steps, result.Dt.String())
	}
	scoreMean := float64(scoreSum) / float64(len(allScores))
	dt := time.Since(allstart)
	fmt.Printf("%d tests, score min=%d mean=%f max=%d, (%s)\n", len(allScores), scoreMin, scoreMean, scoreMax, dt.String())
}

func TestOne() {
	strat := NewRandomStrategy()

	et := NewExponentile(8)

	start := time.Now()
	steps := 0
	for {
		ok, nextm := strat.NextMove(et)
		if !ok {
			break
		}
		et.ApplyMove(nextm)
		steps++
	}
	dt := time.Since(start)
	fmt.Println("No moves found")
	et.Printer.Print(et)
	fmt.Printf("[%3d] score %d, (%s)\n\n", steps, et.Score, dt.String())
}
