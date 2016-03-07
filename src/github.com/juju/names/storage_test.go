// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names_test

import (
	"fmt"

	gc "gopkg.in/check.v1"

	"github.com/juju/names"
)

type storageSuite struct{}

var _ = gc.Suite(&storageSuite{})

func (s *storageSuite) TestStorageTag(c *gc.C) {
	c.Assert(names.NewStorageTag("store/1").String(), gc.Equals, "storage-store-1")
}

func (s *storageSuite) TestStorageNameValidity(c *gc.C) {
	assertStorageIdValid(c, "shared-fs/0")
	assertStorageIdValid(c, "db-dir/1000")
	assertStorageIdInvalid(c, "store/-1")
	assertStorageIdInvalid(c, "store-1")
	assertStorageIdInvalid(c, "")
	assertStorageIdInvalid(c, "store")
	assertStorageIdInvalid(c, "store/#")
}

func (s *storageSuite) TestParseStorageTag(c *gc.C) {
	assertParseStorageTag(c, "storage-shared-fs-0", names.NewStorageTag("shared-fs/0"))
	assertParseStorageTag(c, "storage-store-88", names.NewStorageTag("store/88"))
	assertParseStorageTagInvalid(c, "", names.InvalidTagError("", ""))
	assertParseStorageTagInvalid(c, "one", names.InvalidTagError("one", ""))
	assertParseStorageTagInvalid(c, "storage-", names.InvalidTagError("storage-", names.StorageTagKind))
	assertParseStorageTagInvalid(c, "machine-0", names.InvalidTagError("machine-0", names.StorageTagKind))
}

func (s *serviceSuite) TestStorageName(c *gc.C) {
	assertStorageNameValid(c, "shared-fs/0", "shared-fs")
	assertStorageNameInvalid(c, "storage-shared-fs-0")
}

func assertStorageIdValid(c *gc.C, name string) {
	c.Assert(names.IsValidStorage(name), gc.Equals, true)
	names.NewStorageTag(name)
}

func assertStorageIdInvalid(c *gc.C, name string) {
	c.Assert(names.IsValidStorage(name), gc.Equals, false)
	testStorageTag := func() { names.NewStorageTag(name) }
	expect := fmt.Sprintf("%q is not a valid storage instance ID", name)
	c.Assert(testStorageTag, gc.PanicMatches, expect)
}

func assertStorageNameValid(c *gc.C, id, expect string) {
	name, err := names.StorageName(id)
	c.Assert(err, gc.IsNil)
	c.Assert(name, gc.Equals, expect)
}

func assertStorageNameInvalid(c *gc.C, id string) {
	_, err := names.StorageName(id)
	expect := fmt.Sprintf("%q is not a valid storage instance ID", id)
	c.Assert(err, gc.ErrorMatches, expect)
}

func assertParseStorageTag(c *gc.C, tag string, expect names.StorageTag) {
	t, err := names.ParseStorageTag(tag)
	c.Assert(err, gc.IsNil)
	c.Assert(t, gc.Equals, expect)
}

func assertParseStorageTagInvalid(c *gc.C, tag string, expect error) {
	_, err := names.ParseStorageTag(tag)
	c.Assert(err, gc.ErrorMatches, expect.Error())
}
