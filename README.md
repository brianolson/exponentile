# ExponenTile solver

Someone nerd-sniped me with a link to the game ExponenTile

https://www.bellika.dk/exponentile

Too many hours sunk into this game, the best thing since 2048, required a cure.
Once I've written a solver for the game, it's no longer an interesting game.
(I did this for myself for Soduku many years back.)

## Implement the Game

The game as best as I can reverse-engineer it from playing it.

* 8x8 grid of power-of-two integers, randomly starting with 2,4,8 in each cell
* Swap pairs of cells to make matches
* Match 3 of a kind to replace with 2x that kind. [4, 4, 4] -> [8]
  * Match 4 for 4x, match 5 for 8x, match 6 for 16x, etc 
* The cell that was moved is the anchor and gets the new value
  * OR a match along both X and Y is anchored to the intersection
  * OR a secondary match due to shifting and infill shifts towards the top left
* gravity pulls cells down towards the bottom when there are empty cells
* empty cells at the top get new 2,4,8 values

This is implemented in [exponentile.go](exponentile.go)

## Random Solver

List all the possible moves, pick one at random.

This does okay! (Good for me: playing the game by hand, I can do better than random!)

```
1000 tests, score min=2524 mean=30387.268000 max=135884, (1.850669066s)
```

I can beat the average random game, but I haven't played 1000 games and my high score is not yet up to 135884.

## Recursive Solver

This is the basic game playing solver that goes back to the original Chess playing programs from decades ago.
For each possible move, test each possible move under that, and each possible move under those, out to some depth of recursion.
To start with I set the recursion limit to 3.
This was a guess about how much would be needed to make a better solver, and how much I could do on a home computer.
It was a good guess!
Much better solutions, and a single game runs in between 10 and 200 seconds on a 2024 Ryzen 7 9700 core.

A few of the better runs I got over the course of a couple hours:

```
[   14] 1447524 (16585 steps) (2m40.556446924s)
[   32] 1458700 (15781 steps) (2m58.127756147s)
[  144] 1047248 (12268 steps) (2m7.247434917s)
[  162] 1665980 (18519 steps) (3m10.785653562s)
[  357] 3013644 (33196 steps) (5m48.180016105s)

1000 tests, score min=11360 mean=392987.204000 max=3013644, (1h51m12.580284624s)
```