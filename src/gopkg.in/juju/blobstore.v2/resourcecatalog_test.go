// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package blobstore_test

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/txn"
	txntesting "github.com/juju/txn/testing"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/blobstore.v2"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var _ = gc.Suite(&resourceCatalogSuite{})

type resourceCatalogSuite struct {
	testing.IsolationSuite
	testing.MgoSuite
	txnRunner  txn.Runner
	rCatalog   blobstore.ResourceCatalog
	collection *mgo.Collection
}

func (s *resourceCatalogSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.MgoSuite.SetUpSuite(c)
}

func (s *resourceCatalogSuite) TearDownSuite(c *gc.C) {
	s.MgoSuite.TearDownSuite(c)
	s.IsolationSuite.TearDownSuite(c)
}

func (s *resourceCatalogSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.MgoSuite.SetUpTest(c)
	db := s.Session.DB("blobstore")
	s.collection = db.C("storedResources")
	s.rCatalog = blobstore.NewResourceCatalog(db)

	// For testing, we need to ensure there's a single txnRunner for all operations.
	s.txnRunner = txn.NewRunner(txn.RunnerParams{Database: db})
	txnRunnerFunc := func(db *mgo.Database) txn.Runner {
		return s.txnRunner
	}
	s.PatchValue(blobstore.TxnRunner, txnRunnerFunc)
}

func (s *resourceCatalogSuite) TearDownTest(c *gc.C) {
	s.MgoSuite.TearDownTest(c)
	s.IsolationSuite.TearDownTest(c)
}

func (s *resourceCatalogSuite) assertPut(c *gc.C, expectedNew bool, sha384Hash string) (
	id, path string,
) {
	id, path, err := s.rCatalog.Put(sha384Hash, 200)
	c.Assert(err, gc.IsNil)
	c.Assert(id, gc.Not(gc.Equals), "")
	c.Assert(path, gc.Equals, "")
	s.assertGetPending(c, id)
	return id, path
}

func (s *resourceCatalogSuite) assertGetPending(c *gc.C, id string) {
	r, err := s.rCatalog.Get(id)
	c.Assert(err, gc.Equals, blobstore.ErrUploadPending)
	c.Assert(r, gc.IsNil)
}

func (s *resourceCatalogSuite) asserGetUploaded(c *gc.C, id string, hash string, length int64) {
	r, err := s.rCatalog.Get(id)
	c.Assert(err, gc.IsNil)
	c.Assert(r.SHA384Hash, gc.DeepEquals, hash)
	c.Assert(r.Length, gc.Equals, length)
	c.Assert(r.Path, gc.Not(gc.Equals), "")
}

type resourceDoc struct {
	Id       bson.ObjectId `bson:"_id"`
	RefCount int64
}

func (s *resourceCatalogSuite) assertRefCount(c *gc.C, id string, expected int64) {
	var doc resourceDoc
	err := s.collection.FindId(id).One(&doc)
	c.Assert(err, gc.IsNil)
	c.Assert(doc.RefCount, gc.Equals, expected)
}

func (s *resourceCatalogSuite) TestPut(c *gc.C) {
	id, _ := s.assertPut(c, true, "sha384foo")
	s.assertRefCount(c, id, 1)
}

func (s *resourceCatalogSuite) TestPutLengthMismatch(c *gc.C) {
	id, _ := s.assertPut(c, true, "sha384foo")
	_, _, err := s.rCatalog.Put("sha384foo", 100)
	c.Assert(err, gc.ErrorMatches, "length mismatch in resource document 200 != 100")
	s.assertRefCount(c, id, 1)
}

func (s *resourceCatalogSuite) TestPutSameHashesIncRefCount(c *gc.C) {
	id, _ := s.assertPut(c, true, "sha384foo")
	s.assertPut(c, false, "sha384foo")
	s.assertRefCount(c, id, 2)
}

func (s *resourceCatalogSuite) TestGetNonExistent(c *gc.C) {
	_, err := s.rCatalog.Get(bson.NewObjectId().Hex())
	c.Assert(err, gc.ErrorMatches, `resource with id ".*" not found`)
}

func (s *resourceCatalogSuite) TestGet(c *gc.C) {
	id, path, err := s.rCatalog.Put("sha384foo", 100)
	c.Assert(err, gc.IsNil)
	c.Assert(path, gc.Equals, "")
	s.assertGetPending(c, id)
}

func (s *resourceCatalogSuite) TestFindNonExistent(c *gc.C) {
	_, err := s.rCatalog.Find("sha384foo")
	c.Assert(err, gc.ErrorMatches, `resource with sha384=.* not found`)
}

func (s *resourceCatalogSuite) TestFind(c *gc.C) {
	id, path, err := s.rCatalog.Put("sha384foo", 100)
	c.Assert(err, gc.IsNil)
	c.Assert(path, gc.Equals, "")
	err = s.rCatalog.UploadComplete(id, "wherever")
	c.Assert(err, gc.IsNil)
	foundId, err := s.rCatalog.Find("sha384foo")
	c.Assert(err, gc.IsNil)
	c.Assert(foundId, gc.Equals, id)
}

func (s *resourceCatalogSuite) TestUploadComplete(c *gc.C) {
	id, _, err := s.rCatalog.Put("sha384foo", 100)
	c.Assert(err, gc.IsNil)
	s.assertGetPending(c, id)
	err = s.rCatalog.UploadComplete(id, "wherever")
	c.Assert(err, gc.IsNil)
	s.asserGetUploaded(c, id, "sha384foo", 100)
	// A second call yields an AlreadyExists error.
	err = s.rCatalog.UploadComplete(id, "wherever")
	c.Assert(err, jc.Satisfies, errors.IsAlreadyExists)
	s.asserGetUploaded(c, id, "sha384foo", 100)
}

func (s *resourceCatalogSuite) TestRemoveOnlyRecord(c *gc.C) {
	id, path := s.assertPut(c, true, "sha384foo")
	wasDeleted, removedPath, err := s.rCatalog.Remove(id)
	c.Assert(err, gc.IsNil)
	c.Assert(wasDeleted, jc.IsTrue)
	c.Assert(removedPath, gc.Equals, path)
	_, err = s.rCatalog.Get(id)
	c.Assert(err, gc.ErrorMatches, `resource with id ".*" not found`)
}

func (s *resourceCatalogSuite) TestRemoveDecRefCount(c *gc.C) {
	id, _ := s.assertPut(c, true, "sha384foo")
	s.assertPut(c, false, "sha384foo")
	s.assertRefCount(c, id, 2)
	wasDeleted, _, err := s.rCatalog.Remove(id)
	c.Assert(err, gc.IsNil)
	c.Assert(wasDeleted, jc.IsFalse)
	s.assertRefCount(c, id, 1)
	s.assertGetPending(c, id)
}

func (s *resourceCatalogSuite) TestRemoveLastCopy(c *gc.C) {
	id, _ := s.assertPut(c, true, "sha384foo")
	s.assertPut(c, false, "sha384foo")
	s.assertRefCount(c, id, 2)
	_, _, err := s.rCatalog.Remove(id)
	c.Assert(err, gc.IsNil)
	s.assertRefCount(c, id, 1)
	_, _, err = s.rCatalog.Remove(id)
	c.Assert(err, gc.IsNil)
	_, err = s.rCatalog.Get(id)
	c.Assert(err, gc.ErrorMatches, `resource with id ".*" not found`)
}

func (s *resourceCatalogSuite) TestRemoveNonExistent(c *gc.C) {
	_, _, err := s.rCatalog.Remove(bson.NewObjectId().Hex())
	c.Assert(err, gc.ErrorMatches, `resource with id ".*" not found`)
}

func (s *resourceCatalogSuite) TestPutNewResourceRace(c *gc.C) {
	var firstId string
	beforeFuncs := []func(){
		func() { firstId, _ = s.assertPut(c, true, "sha384foo") },
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFuncs...).Check()
	id, _, err := s.rCatalog.Put("sha384foo", 200)
	c.Assert(err, gc.IsNil)
	c.Assert(id, gc.Equals, firstId)
	err = s.rCatalog.UploadComplete(id, "wherever")
	c.Assert(err, gc.IsNil)
	r, err := s.rCatalog.Get(id)
	c.Assert(err, gc.IsNil)
	s.assertRefCount(c, id, 2)
	c.Assert(r.SHA384Hash, gc.Equals, "sha384foo")
	c.Assert(int(r.Length), gc.Equals, 200)
}

func (s *resourceCatalogSuite) TestPutDeletedResourceRace(c *gc.C) {
	firstId, _ := s.assertPut(c, true, "sha384foo")
	err := s.rCatalog.UploadComplete(firstId, "wherever")
	c.Assert(err, gc.IsNil)
	beforeFuncs := []func(){
		func() {
			_, _, err := s.rCatalog.Remove(firstId)
			c.Assert(err, gc.IsNil)
		},
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFuncs...).Check()
	id, _, err := s.rCatalog.Put("sha384foo", 200)
	c.Assert(err, gc.IsNil)
	c.Assert(firstId, gc.Equals, id)
	err = s.rCatalog.UploadComplete(id, "wherever")
	c.Assert(err, gc.IsNil)
	r, err := s.rCatalog.Get(id)
	c.Assert(err, gc.IsNil)
	s.assertRefCount(c, id, 1)
	c.Assert(r.SHA384Hash, gc.Equals, "sha384foo")
	c.Assert(r.Length, gc.Equals, int64(200))
}

func (s *resourceCatalogSuite) TestDeleteResourceRace(c *gc.C) {
	id, _ := s.assertPut(c, true, "sha384foo")
	s.assertPut(c, false, "sha384foo")
	beforeFuncs := []func(){
		func() {
			_, _, err := s.rCatalog.Remove(id)
			c.Assert(err, gc.IsNil)
		},
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFuncs...).Check()
	_, _, err := s.rCatalog.Remove(id)
	c.Assert(err, gc.IsNil)
	_, err = s.rCatalog.Get(id)
	c.Assert(err, gc.ErrorMatches, `resource with id ".*" not found`)
}

func (s *resourceCatalogSuite) TestUploadCompleteDeleted(c *gc.C) {
	id, _, err := s.rCatalog.Put("sha384foo", 100)
	c.Assert(err, gc.IsNil)
	remove := func() {
		_, _, err := s.rCatalog.Remove(id)
		c.Assert(err, gc.IsNil)
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, remove).Check()
	err = s.rCatalog.UploadComplete(id, "wherever")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}
