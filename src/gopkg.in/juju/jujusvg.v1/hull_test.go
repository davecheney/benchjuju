package jujusvg

import (
	"image"

	gc "gopkg.in/check.v1"
)

type HullSuite struct{}

var _ = gc.Suite(&HullSuite{})

func (s *HullSuite) TestGetPointOutside(c *gc.C) {
	var tests = []struct {
		about    string
		vertices []image.Point
		expected image.Point
	}{
		{
			about:    "zero vertices",
			vertices: []image.Point{},
			expected: image.Point{0, 0},
		},
		{
			about:    "one vertex",
			vertices: []image.Point{{0, 0}},
			expected: image.Point{10, 10},
		},
		{
			about:    "two vertices",
			vertices: []image.Point{{0, 0}, {10, 10}},
			expected: image.Point{20, 20},
		},
		{
			about:    "three vertices (convexHull fall through)",
			vertices: []image.Point{{0, 0}, {0, 10}, {10, 0}},
			expected: image.Point{10, 20},
		},
		{
			about:    "four vertices",
			vertices: []image.Point{{0, 0}, {0, 10}, {10, 0}, {10, 10}},
			expected: image.Point{20, 20},
		},
	}
	for _, test := range tests {
		c.Log(test.about)
		c.Assert(getPointOutside(test.vertices, image.Point{10, 10}), gc.Equals, test.expected)
	}
}

func (s *HullSuite) TestConvexHull(c *gc.C) {
	// Zero vertices
	vertices := []image.Point{}
	c.Assert(convexHull(vertices), gc.DeepEquals, []image.Point{{0, 0}})

	// Identities
	vertices = []image.Point{{1, 1}}
	c.Assert(convexHull(vertices), gc.DeepEquals, vertices)

	vertices = []image.Point{{1, 1}, {2, 2}}
	c.Assert(convexHull(vertices), gc.DeepEquals, vertices)

	vertices = []image.Point{{1, 1}, {2, 2}, {1, 2}}
	c.Assert(convexHull(vertices), gc.DeepEquals, vertices)

	// > 3 vertices
	vertices = []image.Point{}
	for i := 0; i < 100; i++ {
		vertices = append(vertices, image.Point{i / 10, i % 10})
	}
	c.Assert(convexHull(vertices), gc.DeepEquals, []image.Point{
		{0, 0},
		{9, 0},
		{9, 9},
		{0, 9},
	})
}
