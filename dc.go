package main

import (
	"fmt"
)

type Cubes struct {
	seen   []bool
	uniq   map[uint16]bool
}

func NewCubes() (cubes *Cubes) {
	cubes = &Cubes{seen: make([]bool, b(12)), uniq:make(map[uint16]bool)}
	return
}

//
// an incomplete cube is represented by 12 bits, one for each possible edge.
// 0x000 is no edges, 0xfff is all edges.
// the floor is edges 0,1,2,3.  vertical edges 4,5,6,7.  roof is 8,9,10,11.
//

const FloorMask = 1 << 0 | 1 << 1 | 1 << 2  | 1 << 3
const VertMask  = 1 << 4 | 1 << 5 | 1 << 6  | 1 << 7
const RoofMask  = 1 << 8 | 1 << 9 | 1 << 10 | 1 << 11
const WidthMask = 1 << 0 | 1 << 2 | 1 << 8  | 1 << 10
const DepthMask = 1 << 1 | 1 << 3 | 1 << 9  | 1 << 11

//
// a 3D cube has a segment in all three dimensions.
//
func is3d(i uint16) bool {
	return (i & VertMask) !=  0 && (i & WidthMask) != 0 && (i & DepthMask) != 0
}

var Neighbors = [][]int {
	// floor
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

// rotate the cube 90 degrees around X axis
func spinx(i uint16) (uint16) {

	return  (i & 0x11) << 4  | (i & 2) << 2     | (i & 4) << 5    | (i & 8) << 8 |
		(i & 0x20) >> 5  | (i & 0x440) >> 4 | (i & 0x80) << 3 | (i & 0x100) >> 3 |
		(i & 0x200) >> 8 | (i & 0x800) >> 2
}

// rotate the cube 90 degrees around vertical(z) axis
func spinz(i uint16) (uint16) {
	return (i & 0x777) << 1 | (i & 0x888) >> 3
}

// rotate the cube -90 degrees around vertical(z) axis
func spinzc(i uint16) (uint16) {
	return (i & 0xeee) >> 1 | (i & 0x111) << 3
}

func b(n int) uint16 {
	return 1 << n
}

// visit neighbors recursively
func isc(s int, i uint16) uint16 {
	// clear current edge
	i &= ^b(s)

	for _, x := range Neighbors[s] {
		if (i & b(x)) != 0 {
			i = isc(x, i)
		}
	}
	return i
}

func isconnected(i uint16) (bool) {
	if i == 0 {return false}

	start := 0
	for (i & b(start)) == 0 {
		start++
	}

	i = isc(start, i)

	// if more edges, then we visited all of them, we are connected
	return i == 0
}

func (cubes *Cubes) add(i uint16) {
	cubes.uniq[i] = true
}

// rotate around the Z axis 3 times, mark all forms as seen
func (cubes *Cubes) spinz3(i uint16) {

	cubes.seen[i] = true
	for s := 0; s < 3; s++ {
		i = spinz(i)
		cubes.seen[i] = true
	}
	return
}

//
// use spinx()/spinz() to rotate the cube around so each of the 6 faces is facing up.
// for each of these, call spinz3() to mark 4 orientations as seen.
//
func (cubes *Cubes) spinmark(i uint16) {
	cubes.spinz3(i)

	j := spinx(i)
	cubes.spinz3(j)

	j = spinx(j)
	cubes.spinz3(j)

	j = spinx(j)
	cubes.spinz3(j)

	j = spinz(i)
	j = spinx(j)
	cubes.spinz3(j)

	j = spinzc(i)
	j = spinx(j)
	cubes.spinz3(j)
}

func main() {

	cubes := NewCubes()

	// consider each of the 4095 incomplete cubes
	for i := uint16(0); i < b(12)-1; i++ {
		if !cubes.seen[i] {
			// cube must have height, width and depth and must be connected
			if is3d(i) && isconnected(i) {
				// this is a valid cube we've not seen before
				cubes.add(i)
				//fmt.Printf("%b\n", i)
			}
			// mark all versions of this cube as seen
			cubes.spinmark(i)

		}
	}
	fmt.Printf("total unique incomplete open cubes: %d\n", len(cubes.uniq))
}
