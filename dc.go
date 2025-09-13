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
// the floor edges are 0,1,2,3.  vertical edges are 4,5,6,7.  roof edges are 8,9,10,11.
//

const FloorMask = 1 << 0 | 1 << 1 | 1 << 2  | 1 << 3
const VertMask  = 1 << 4 | 1 << 5 | 1 << 6  | 1 << 7
const RoofMask  = 1 << 8 | 1 << 9 | 1 << 10 | 1 << 11
const WidthMask = 1 << 0 | 1 << 2 | 1 << 8  | 1 << 10
const DepthMask = 1 << 1 | 1 << 3 | 1 << 9  | 1 << 11

//
// a 3D cube has an active segment in all three dimensions.
//
func is3d(i uint16) bool {
	return (i & VertMask) !=  0 && (i & WidthMask) != 0 && (i & DepthMask) != 0
}

// rotate the cube 90 degrees around X axis
func spinx(i uint16) (uint16) {
	return  (i & 0x11) << 4  | (i & 2) << 2     | (i & 4) << 5    | (i & 8) << 8 |
		(i & 0x20) >> 5  | (i & 0x440) >> 4 | (i & 0x80) << 3 | (i & 0x100) >> 3 |
		(i & 0x200) >> 8 | (i & 0x800) >> 2
}

// rotate the cube 90 degrees around vertical(z) axis
func spinz(i uint16, cc bool) (uint16) {
	if cc {
		return (i & 0xeee) >> 1 | (i & 0x111) << 3
	}
	return         (i & 0x777) << 1 | (i & 0x888) >> 3
}

func b(n int) uint16 {
	return 1 << n
}


// neighbors for each edge
var nmask = [...]uint16{0x03a, 0x065, 0x0ca, 0x095, 0x909, 0x303, 0x606, 0xc0c, 0xa30, 0x560, 0xac0,  0x590}

// visit neighbors recursively
func isc(s int, i uint16) uint16 {

	// clear current edge
	i &= ^b(s)

	mask := nmask[s]
	for x := 0; i != 0 && x < 12; x++ {
		if (b(x) & i & mask) != 0 {
			i = isc(x, i)
		}
	}
	return i
}

func isconnected(i uint16) (bool) {
	if i == 0 {return false}

	s := 0
	for (i & b(s)) == 0 {
		s++
	}
	// if no more edges, then we visited all of them, we are connected
	return isc(s, i) == 0
}

// rotate around the Z axis 3 times, mark all forms as seen
func (cubes *Cubes) spinz3(i uint16) {

	cubes.seen[i] = true
	for s := 0; s < 3; s++ {
		i = spinz(i, false)
		cubes.seen[i] = true
	}
	return
}

//
// use spinx()/spinz() to rotate the cube around so each of the 6 faces is facing up.
// for each of these, call spinz3() to mark 4 orientations as seen.
//
func (cubes *Cubes) spinmark(i uint16) {
	// top
	cubes.spinz3(i)

	// left
	j := spinx(i)
	cubes.spinz3(j)

	// bottom
	j = spinx(j)
	cubes.spinz3(j)

	// right
	j = spinx(j)
	cubes.spinz3(j)

	// back
	j = spinz(i, false)
	j = spinx(j)
	cubes.spinz3(j)

	// front
	j = spinz(i, true)
	j = spinx(j)
	cubes.spinz3(j)
}

func (cubes *Cubes) calculate() int {
	// consider each of the 4095 incomplete cubes
	for i := uint16(0); i < b(12)-1; i++ {
		if !cubes.seen[i] {
			// cube must have height, width and depth and must be connected
			if is3d(i) && isconnected(i) {
				// this is a valid cube we've not seen before
				cubes.uniq[i] = true
				//fmt.Printf("%b\n", i)
			}
			// mark all versions of this cube as seen
			cubes.spinmark(i)

		}
	}
	return len(cubes.uniq)
}


func main() {

	n := NewCubes().calculate()

	fmt.Printf("total unique incomplete open cubes: %d\n", n)
}
