package main

import (
	"fmt"
	"math/bits"
)

// floor is edges 0, 1, 2, 3
//const FloorMask = 1 << 0 | 1 << 1 | 1 << 2 | 1 << 3
// vertical edges 4, 5, 6, 7
const VertMask = 1 << 4 | 1 << 5 | 1 << 6 | 1 << 7
// width
const WidthMask = 1 << 0 | 1 << 2 | 1 << 8 | 1 << 10
// depth
const DepthMask = 1 << 1 | 1 << 3 | 1 << 9 | 1 << 11

var Neighbors = [][]int {
	{1, 3, 4, 5},  // 0
	{0, 2, 5, 6},  // 1
	{1, 3, 6, 7},  // 2
	{0, 2, 4, 7},  // 3

	// vertical
	{0, 3, 8, 11},  // 4
	{0, 1, 8, 9},   // 5
	{1, 2, 9, 10},  // 6
	{2, 3, 10, 11}, // 7

	// roof
	{4, 5, 9, 11},  // 8
	{5, 6, 8, 10},  // 9
	{6, 7, 9, 11},  // 10
	{4, 7, 8, 10},  // 11
}

var xperm = []uint8{4, 3, 7, 11, 8, 0, 2, 10, 5, 1, 6, 9}

// rotate 90 degrees around X axis
func spinx(i uint) (j uint) {

	for s, d := range xperm {
		j |= ((i >> uint(s)) & 1) << d
	}
	return j
}

// rotate 90 degrees around vertical(z) axis
func spinz(i uint) (uint) {
	return (i & 0x777) << 1 | (i & 0x888) >> 3
}

func b(n int) uint {
	return 1 << n
}


// visit neighbors recursively
func isc(s int, i uint) uint {
	// clear current edge
	i &= ^b(s)

	for _, x := range Neighbors[s] {
		if (i & b(x)) != 0 {
			i = isc(x, i)
		}
	}
	return i
}

func isconnected(i uint) (c bool) {
	if i == 0 {return}

	start := 0
	for (i & b(start)) == 0 {
		start++
	}

	i = isc(start, i)
	if i == 0 {c = true}
	return
}

// rotate around the Z axis 3 times, see if we find a duplicate
func spinz3(dup bool, i uint, cubes map[uint]bool) (seen bool) {
	seen = dup
	if seen {return}

	if cubes[i] {seen = true }
	for s := 0; s < 3; s++ {
		i = spinz(i)
		if cubes[i] {
			seen = true
			return
		}
	}
	return
}

func main() {

	cubes := make(map[uint]bool)

	for i := uint(0); i < 1 << 12; i++ {
		// cube must have height, width and depth
		height := bits.OnesCount(i & VertMask) 		// number of vertical edges
		width := bits.OnesCount(i & WidthMask)
		depth := bits.OnesCount(i & DepthMask)

		connected := isconnected(i)
		// total non-missing edges
		ne := bits.OnesCount(i)

		// must be connected and 3d, more than 2, but not 12 edges
		if ne > 2 && ne < 12 && connected && height > 0 && depth > 0 && width > 0 {

			//
			// use spinx()/spinz() to rotate the cube around soeach of the 6 faces is facing up.
			// for each of these, call spinz3() to see if we have made a duplicate.
			//

			dup := spinz3(false, i, cubes)

			j := spinx(i)
			dup = spinz3(dup, j, cubes)

			j = spinx(j)
			dup = spinz3(dup, j, cubes)

			j = spinx(j)
			dup = spinz3(dup, j, cubes)

			j = spinz(i)
			j = spinx(j)
			dup = spinz3(dup, j, cubes)

			j = spinz(i)
			j = spinz(j)
			j = spinz(j)
			j = spinx(j)
			dup = spinz3(dup, j, cubes)

			if !dup {
				// a new cube we've not seen before
				cubes[i] = true
			}
		}
		
	}
	fmt.Printf("total unique incomplete open cubes: %d\n", len(cubes))
}
