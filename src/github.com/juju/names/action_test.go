// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type actionSuite struct{}

var _ = gc.Suite(&actionSuite{})

var parseActionTagTests = []struct {
	tag      string
	expected names.Tag
	err      error
}{
	{tag: "", err: names.InvalidTagError("", "")},
	{tag: "action-f47ac10b-58cc-4372-a567-0e02b2c3d479", expected: names.NewActionTag("f47ac10b-58cc-4372-a567-0e02b2c3d479")},
	{tag: "action-012345678", err: names.InvalidTagError("action-012345678", "action")},
	{tag: "action-1234567", err: names.InvalidTagError("action-1234567", "action")},
	{tag: "bob", err: names.InvalidTagError("bob", "")},
	{tag: "service-ned", err: names.InvalidTagError("service-ned", names.ActionTagKind)}}

func (s *actionSuite) TestParseActionTag(c *gc.C) {
	for i, t := range parseActionTagTests {
		c.Logf("test %d: %s", i, t.tag)
		got, err := names.ParseActionTag(t.tag)
		if err != nil || t.err != nil {
			c.Check(err, gc.DeepEquals, t.err)
			continue
		}
		c.Check(got, gc.FitsTypeOf, t.expected)
		c.Check(got, gc.Equals, t.expected)
	}
}

func (s *actionSuite) TestActionReceiverTag(c *gc.C) {
	testCases := []struct {
		name     string
		expected names.Tag
		valid    bool
	}{
		{name: "mysql", valid: false},
		{name: "mysql/3", expected: names.NewUnitTag("mysql/3"), valid: true},
	}

	for _, tcase := range testCases {
		tag, err := names.ActionReceiverTag(tcase.name)
		c.Check(err == nil, gc.Equals, tcase.valid)
		if err != nil {
			continue
		}
		c.Check(tag, gc.FitsTypeOf, tcase.expected)
		c.Check(tag, gc.Equals, tcase.expected)
	}

}
