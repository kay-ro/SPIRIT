package function

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// test if sort does infact sort the points
func TestPoint(t *testing.T) {
	p1 := &Point{X: 1, Y: 2, Error: 3}
	p2 := &Point{X: 4, Y: 5, Error: 6}
	p3 := &Point{X: 7, Y: 8}

	points := Points{
		p2,
		p1,
		p3,
	}

	points.Sort()

	if points[0] != p1 || points[1] != p2 || points[2] != p3 {
		t.Errorf("expected %v got %v", p1, points)
	}

	spew.Dump(points)
}
