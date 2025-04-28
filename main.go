package main

import (
	crand "crypto/rand"
	"fmt"
	"math/rand/v2"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

func main() {

	var seed [32]byte
	_, _ = crand.Read(seed[:])
	rng := rand.New(rand.NewChaCha8(seed))

	et := NewExponentile(8)

	steps := 0
	for {
		et.Printer.Print(et)
		fmt.Printf("[%3d] score %d\n\n", steps, et.Score)
		moves := et.FindMoves()
		if len(moves) == 0 {
			fmt.Println("No moves found")
			break
		}
		nextm := moves[rng.IntN(len(moves))]
		et.ApplyMove(nextm)
		steps++
		if steps > 10 {
			break
		}
	}
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
