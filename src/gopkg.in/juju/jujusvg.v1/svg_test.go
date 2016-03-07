package jujusvg

import (
	"bytes"
	"fmt"

	"github.com/juju/xml"
	gc "gopkg.in/check.v1"
)

type SVGSuite struct{}

var _ = gc.Suite(&SVGSuite{})

func (s *SVGSuite) TestProcessIcon(c *gc.C) {
	tests := []struct {
		about    string
		icon     string
		expected string
		err      string
	}{
		{
			about: "Nothing stripped",
			icon: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-0">
					<g id="foo"></g>
				</svg>`,
		},
		{
			about: "SVG inside an SVG",
			icon: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<svg>
						<g id="foo"></g>
					</svg>
					<g id="bar"></g>
				</svg>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-1">
					<svg>
						<g id="foo"></g>
					</svg>
					<g id="bar"></g>
				</svg>`,
		},
		{
			about: "ProcInst at start stripped",
			icon: `
				<?xml version="1.0"?>
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-2">
					<g id="foo"></g>
				</svg>`,
		},
		{
			about: "Directive at start stripped",
			icon: `
				<!DOCTYPE svg>
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-3">
					<g id="foo"></g>
				</svg>`,
		},
		{
			about: "ProcInst at end stripped",
			icon: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				<?procinst foo="bar"?>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-4">
					<g id="foo"></g>
				</svg>`,
		},
		{
			about: "Directive at end stripped",
			icon: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				<!DOCTYPE svg>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-5">
					<g id="foo"></g>
				</svg>`,
		},
		{
			about: "ProcInsts/Directives inside svg left in place",
			icon: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<!DOCTYPE svg>
					<?proc foo="bar"?>
					<g id="foo"></g>
				</svg>
				`,
			expected: `
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" id="test-6">
					<!DOCTYPE svg>
					<?proc foo="bar"?>
					<g id="foo"></g>
				</svg>`,
		},
		{
			about: "Not an SVG",
			icon: `
				<html xmlns="foo">
					<body>bad-wolf</body>
				</html>
				`,
			err: "icon does not appear to be a valid SVG",
		},
	}
	for i, test := range tests {
		in := bytes.NewBuffer([]byte(test.icon))
		out := bytes.Buffer{}
		err := processIcon(in, &out, fmt.Sprintf("test-%d", i))
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, gc.IsNil)
			assertXMLEqual(c, out.Bytes(), []byte(test.expected))
		}
	}
}

func (s *SVGSuite) TestSetXMLAttr(c *gc.C) {
	// Attribute is added.
	expected := []xml.Attr{
		{
			Name: xml.Name{
				Local: "id",
			},
			Value: "foo",
		},
	}

	result := setXMLAttr([]xml.Attr{}, xml.Name{
		Local: "id",
	}, "foo")
	c.Assert(result, gc.DeepEquals, expected)

	// Attribute is changed.
	result = setXMLAttr([]xml.Attr{
		{
			Name: xml.Name{
				Local: "id",
			},
			Value: "bar",
		},
	}, xml.Name{
		Local: "id",
	}, "foo")
	c.Assert(result, gc.DeepEquals, expected)

	// Attribute is changed, existing attributes unchanged.
	expected = []xml.Attr{
		{
			Name: xml.Name{
				Local: "class",
			},
			Value: "bar",
		},
		{
			Name: xml.Name{
				Local: "id",
			},
			Value: "foo",
		},
	}
	result = setXMLAttr([]xml.Attr{
		{
			Name: xml.Name{
				Local: "class",
			},
			Value: "bar",
		},
		{
			Name: xml.Name{
				Local: "id",
			},
			Value: "bar",
		},
	}, xml.Name{
		Local: "id",
	}, "foo")
	c.Assert(result, gc.DeepEquals, expected)

	// Attribute is added, existing attributes unchanged.
	result = setXMLAttr([]xml.Attr{
		{
			Name: xml.Name{
				Local: "class",
			},
			Value: "bar",
		},
	}, xml.Name{
		Local: "id",
	}, "foo")
	c.Assert(result, gc.DeepEquals, expected)
}
