// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

var _ = gc.Suite(&payloadSuite{})

type payloadSuite struct{}

type payloadTest struct {
	input string
}

func checkPayload(c *gc.C, id string, tag names.PayloadTag) {
	c.Check(tag.Kind(), gc.Equals, names.PayloadTagKind)
	c.Check(tag.Id(), gc.Equals, id)
	c.Check(tag.String(), gc.Equals, names.PayloadTagKind+"-"+id)
}

func (s *payloadSuite) TestPayloadTag(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	tag := names.NewPayloadTag(id)

	c.Check(tag.Kind(), gc.Equals, names.PayloadTagKind)
	c.Check(tag.Id(), gc.Equals, id)
	c.Check(tag.String(), gc.Equals, names.PayloadTagKind+"-"+id)
}

func (s *payloadSuite) TestIsValidPayload(c *gc.C) {
	for i, test := range []struct {
		id     string
		expect bool
	}{
		{"", false},
		{"spam", false},

		{"f47ac10b-58cc-4372-a567-0e02b2c3d479", true},
	} {
		c.Logf("test %d: %s", i, test.id)
		ok := names.IsValidPayload(test.id)

		c.Check(ok, gc.Equals, test.expect)
	}
}

func (s *payloadSuite) TestParsePayloadTag(c *gc.C) {
	for i, test := range []struct {
		tag      string
		expected names.Tag
		err      error
	}{{
		tag: "",
		err: names.InvalidTagError("", ""),
	}, {
		tag: "payload-",
		err: names.InvalidTagError("payload-", names.PayloadTagKind),
	}, {
		tag: "payload-spam",
		err: names.InvalidTagError("payload-spam", names.PayloadTagKind),
	}, {
		tag:      "payload-f47ac10b-58cc-4372-a567-0e02b2c3d479",
		expected: names.NewPayloadTag("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
	}, {
		tag: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
		err: names.InvalidTagError("f47ac10b-58cc-4372-a567-0e02b2c3d479", ""),
	}, {
		tag: "unit-f47ac10b-58cc-4372-a567-0e02b2c3d479",
		err: names.InvalidTagError("unit-f47ac10b-58cc-4372-a567-0e02b2c3d479", names.UnitTagKind),
	}, {
		tag: "action-f47ac10b-58cc-4372-a567-0e02b2c3d479",
		err: names.InvalidTagError("action-f47ac10b-58cc-4372-a567-0e02b2c3d479", names.PayloadTagKind),
	}} {
		c.Logf("test %d: %s", i, test.tag)
		got, err := names.ParsePayloadTag(test.tag)
		if test.err != nil {
			c.Check(err, jc.DeepEquals, test.err)
		} else {
			c.Check(err, jc.ErrorIsNil)
			c.Check(got, jc.DeepEquals, test.expected)
		}
	}
}
