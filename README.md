# ExponenTile solver

Someone nerd-sniped me with a link to the game ExponenTile

https://www.bellika.dk/exponentile

https://github.com/MikeBellika/tile-game

Too many hours sunk into this game, the best thing since 2048, required a cure.
Once I've written a solver for the game, it's no longer an interesting game.
(I did this for myself for Soduku many years back.)

## Implement the Game

The game as best as I can reverse-engineer it from playing it.
(Oops, later found the source, TODO: check what the source does.)

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

Pretty good! What happens if I increase the search depth by one?

```
[   57] 3968384 (43342 steps) (59m55.547424465s)
[  532] 3424580 (38400 steps) (50m23.142713213s)
[  219] 3243092 (36383 steps) (50m7.140595522s)
[  340] 3193312 (36481 steps) (45m27.303326833s)

650 tests, score min=20460 mean=929657.5384615385 max=3968384
```

It takes a lot longer and has results quite a bit better!

* The min finishing score almost doubled
* The mean finishing score increased 2.36x
* The max score increased 31.7%

The peak of what's possible didn't advance very far, but most solutions go tmuch better.
The tradeoff is it used a lot more CPU time.
I only ran 650 tests in about 24 hours and then got bore of waiting and wrote this up.

## Possible Future Work

The Recursive Solver is single threaded.
This is fine for exploring the solution space and running many independent tests, but no one solution is produced very fast.
There's parallelism possible, at the very least the first level or two of recursion could run multi threaded (with recursion happeninng within a thread for each first or second level move).

The Recursive Solver doesn't account for random effects.
It could potentially use oversampling, run each move multiple times, to get an average effect of the move plus possible random new tiles showing up.

Micro-Optimizations. I threw this together as a hobby project and it's fast enough.
Actually running a profiler to see where the CPU time goes might show some things to tweak.
