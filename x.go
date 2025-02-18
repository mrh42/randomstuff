package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Constants for grid dimensions.
const (
	GridSize      = 6
	GenomeLength  = GridSize * GridSize
	PopulationSize = 100
	MutationRate   = 0.01
	TournamentSize = 5
	Generations    = 1000
)

// Individual represents a candidate solution.
type Individual struct {
	Grid    [GridSize][GridSize]int
	Fitness int
}

// isPrime checks if n is prime.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	limit := int(math.Sqrt(float64(n)))
	for i := 2; i <= limit; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// extractSequences extracts all numbers formed by consecutive digits
// in horizontal, vertical, and diagonal directions (both forward and backward).
func extractSequences(grid [GridSize][GridSize]int) map[int]struct{} {
	sequences := make(map[int]struct{})
	
	// Helper function to convert a slice of digits to an integer.
	toNumber := func(digits []int) int {
		n := 0
		for _, d := range digits {
			n = n*10 + d
		}
		return n
	}

	// Directions: (dx, dy). We cover 8 directions.
	dirs := [][2]int{
		{0, 1},  // right
		{0, -1}, // left
		{1, 0},  // down
		{-1, 0}, // up
		{1, 1},  // down-right
		{1, -1}, // down-left
		{-1, 1}, // up-right
		{-1, -1},// up-left
	}

	// For every starting cell, extend in each direction.
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			for _, d := range dirs {
				// Build sequences of length 1 to GridSize (or until out of bounds).
				var digits []int
				x, y := i, j
				for {
					// Check bounds.
					if x < 0 || x >= GridSize || y < 0 || y >= GridSize {
						break
					}
					digits = append(digits, grid[x][y])
					// Only consider sequences of length >= 1.
					num := toNumber(digits)
					sequences[num] = struct{}{}
					// Move along the direction.
					x += d[0]
					y += d[1]
				}
			}
		}
	}

	return sequences
}

// fitness calculates the number of distinct primes found in the grid.
func fitness(grid [GridSize][GridSize]int) int {
	seqs := extractSequences(grid)
	count := 0
	// Check each sequence for primality.
	for num := range seqs {
		if isPrime(num) {
			count++
		}
	}
	return count
}

// randomIndividual creates a new individual with a randomly filled grid.
func randomIndividual() Individual {
	var ind Individual
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			ind.Grid[i][j] = rand.Intn(10) // Random digit 0..9
		}
	}
	ind.Fitness = fitness(ind.Grid)
	return ind
}

// tournamentSelection selects one individual via tournament selection.
func tournamentSelection(population []Individual) Individual {
	best := population[rand.Intn(len(population))]
	for i := 1; i < TournamentSize; i++ {
		competitor := population[rand.Intn(len(population))]
		if competitor.Fitness > best.Fitness {
			best = competitor
		}
	}
	return best
}

// crossover produces a child from two parents using uniform crossover.
func crossover(parent1, parent2 Individual) Individual {
	var child Individual
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			if rand.Float64() < 0.5 {
				child.Grid[i][j] = parent1.Grid[i][j]
			} else {
				child.Grid[i][j] = parent2.Grid[i][j]
			}
		}
	}
	child.Fitness = fitness(child.Grid)
	return child
}

// mutate applies random mutation to an individual.
func mutate(ind *Individual) {
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			if rand.Float64() < MutationRate {
				ind.Grid[i][j] = rand.Intn(10)
			}
		}
	}
	ind.Fitness = fitness(ind.Grid)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initialize population.
	population := make([]Individual, PopulationSize)
	for i := range population {
		population[i] = randomIndividual()
	}

	bestOverall := population[0]

	// Main GA loop.
	for gen := 0; gen < Generations; gen++ {
		newPopulation := make([]Individual, 0, PopulationSize)
		
		// Elitism: carry over the best individual.
		currentBest := population[0]
		for _, ind := range population {
			if ind.Fitness > currentBest.Fitness {
				currentBest = ind
			}
		}
		if currentBest.Fitness > bestOverall.Fitness {
			bestOverall = currentBest
			fmt.Printf("Generation %d, new best fitness: %d\n", gen, bestOverall.Fitness)
		}
		newPopulation = append(newPopulation, currentBest)

		// Create new individuals.
		for len(newPopulation) < PopulationSize {
			parent1 := tournamentSelection(population)
			parent2 := tournamentSelection(population)
			child := crossover(parent1, parent2)
			mutate(&child)
			newPopulation = append(newPopulation, child)
		}

		population = newPopulation
	}

	// Print best grid and its fitness.
	fmt.Printf("\nBest Grid with Fitness %d:\n", bestOverall.Fitness)
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			fmt.Printf("%d ", bestOverall.Grid[i][j])
		}
		fmt.Println()
	}
}
