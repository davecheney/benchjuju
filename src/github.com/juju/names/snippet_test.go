package names

import (
	"strings"

	gc "gopkg.in/check.v1"
)

var snippets = []struct {
	name    string
	snippet string
}{
	{"ContainerTypeSnippet", ContainerTypeSnippet},
	{"ContainerSnippet", ContainerSnippet},
	{"MachineSnippet", MachineSnippet},
	{"NumberSnippet", NumberSnippet},
	{"ServiceSnippet", ServiceSnippet},
	{"RelationSnippet", RelationSnippet},
}

type snippetSuite struct{}

var _ = gc.Suite(&snippetSuite{})

func (s *equalitySuite) TestSnippetsContainNoCapturingGroups(c *gc.C) {
	for _, test := range snippets {
		for i, ch := range test.snippet {
			if ch == '(' && !strings.HasPrefix(test.snippet[i:], "(?:") {
				c.Errorf("%s (%q) contains capturing group", test.name, test.snippet)
			}
		}
	}
}
