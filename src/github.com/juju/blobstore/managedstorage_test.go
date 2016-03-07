// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package blobstore_test

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	jujutxn "github.com/juju/txn"
	txntesting "github.com/juju/txn/testing"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/txn"

	"github.com/juju/blobstore"
)

var _ = gc.Suite(&managedStorageSuite{})

type managedStorageSuite struct {
	testing.IsolationSuite
	testing.MgoSuite
	txnRunner       jujutxn.Runner
	managedStorage  blobstore.ManagedStorage
	db              *mgo.Database
	resourceStorage blobstore.ResourceStorage
}

func (s *managedStorageSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.MgoSuite.SetUpSuite(c)
}

func (s *managedStorageSuite) TearDownSuite(c *gc.C) {
	s.MgoSuite.TearDownSuite(c)
	s.IsolationSuite.TearDownSuite(c)
}

func (s *managedStorageSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.MgoSuite.SetUpTest(c)
	s.db = s.Session.DB("blobstore")
	s.resourceStorage = blobstore.NewGridFS("storage", "test", s.Session)
	s.managedStorage = blobstore.NewManagedStorage(s.db, s.resourceStorage)

	// For testing, we need to ensure there's a single txnRunner for all operations.
	s.txnRunner = jujutxn.NewRunner(jujutxn.RunnerParams{Database: s.db})
	txnRunnerFunc := func(db *mgo.Database) jujutxn.Runner {
		return s.txnRunner
	}
	s.PatchValue(blobstore.TxnRunner, txnRunnerFunc)
}

func (s *managedStorageSuite) TearDownTest(c *gc.C) {
	s.MgoSuite.TearDownTest(c)
	s.IsolationSuite.TearDownTest(c)
}

func (s *managedStorageSuite) TestResourceStoragePath(c *gc.C) {
	for _, test := range []struct {
		envUUID     string
		user        string
		path        string
		storagePath string
		error       string
	}{
		{
			envUUID:     "",
			user:        "",
			path:        "/path/to/blob",
			storagePath: "global/path/to/blob",
		}, {
			envUUID:     "env",
			user:        "",
			path:        "/path/to/blob",
			storagePath: "environs/env/path/to/blob",
		}, {
			envUUID:     "",
			user:        "user",
			path:        "/path/to/blob",
			storagePath: "users/user/path/to/blob",
		}, {
			envUUID:     "env",
			user:        "user",
			path:        "/path/to/blob",
			storagePath: "environs/env/users/user/path/to/blob",
		}, {
			envUUID: "env/123",
			user:    "user",
			path:    "/path/to/blob",
			error:   `.* cannot contain "/"`,
		}, {
			envUUID: "env",
			user:    "user/123",
			path:    "/path/to/blob",
			error:   `.* cannot contain "/"`,
		},
	} {
		result, err := blobstore.ResourceStoragePath(s.managedStorage, test.envUUID, test.user, test.path)
		if test.error == "" {
			c.Check(err, gc.IsNil)
			c.Check(result, gc.Equals, test.storagePath)
		} else {
			c.Check(err, gc.ErrorMatches, test.error)
		}
	}
}

type managedResourceDocStub struct {
	Path       string
	ResourceId string
}

type resourceDocStub struct {
	Path string
}

func (s *managedStorageSuite) TestGetPendingUpload(c *gc.C) {
	// Manually set up a scenario where there's a resource recorded
	// but the upload has not occurred.
	rc := blobstore.GetResourceCatalog(s.managedStorage)
	id, _, err := rc.Put("foo", 100)
	c.Assert(err, gc.IsNil)
	managedResource := blobstore.ManagedResource{
		EnvUUID: "env",
		User:    "user",
		Path:    "environs/env/path/to/blob",
	}
	_, err = blobstore.PutManagedResource(s.managedStorage, managedResource, id)
	c.Assert(err, gc.IsNil)
	_, _, err = s.managedStorage.GetForEnvironment("env", "/path/to/blob")
	c.Assert(err, gc.Equals, blobstore.ErrUploadPending)
}

func (s *managedStorageSuite) TestPutPendingUpload(c *gc.C) {
	// Manually set up a scenario where there's a resource recorded
	// but the upload has not occurred.
	rc := blobstore.GetResourceCatalog(s.managedStorage)
	hash := "cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7"

	id, path, err := rc.Put(hash, 3)
	c.Assert(err, gc.IsNil)
	c.Assert(path, gc.Equals, "")
	managedResource := blobstore.ManagedResource{
		EnvUUID: "env",
		User:    "user",
		Path:    "environs/env/path/to/blob",
	}
	c.Assert(err, gc.IsNil)

	_, err = blobstore.PutManagedResource(s.managedStorage, managedResource, id)
	_, _, err = s.managedStorage.GetForEnvironment("env", "/path/to/blob")
	c.Assert(errors.Cause(err), gc.Equals, blobstore.ErrUploadPending)

	// Despite the upload being pending, a second concurrent upload will succeed.
	rdr := bytes.NewReader([]byte("abc"))
	err = s.managedStorage.PutForEnvironment("env", "/path/to/blob", rdr, 3)
	c.Assert(err, gc.IsNil)
	s.assertGet(c, "/path/to/blob", []byte("abc"))
}

func (s *managedStorageSuite) assertPut(c *gc.C, path string, blob []byte) string {
	// Put the data.
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", path, rdr, int64(len(blob)))
	c.Assert(err, gc.IsNil)

	// Load the managed resource record.
	var mrDoc managedResourceDocStub
	err = s.db.C("managedStoredResources").Find(bson.D{{"path", "environs/env" + path}}).One(&mrDoc)
	c.Assert(err, gc.IsNil)

	// Load the corresponding resource catalog record.
	var rd resourceDocStub
	err = s.db.C("storedResources").FindId(mrDoc.ResourceId).One(&rd)
	c.Assert(err, gc.IsNil)

	// Use the resource catalog record to load the underlying data from blobstore.
	r, err := s.resourceStorage.Get(rd.Path)
	c.Assert(err, gc.IsNil)
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	c.Assert(err, gc.IsNil)
	c.Assert(data, gc.DeepEquals, blob)
	return rd.Path
}

func (s *managedStorageSuite) assertResourceCatalogCount(c *gc.C, expected int) {
	num, err := s.db.C("storedResources").Count()
	c.Assert(err, gc.IsNil)
	c.Assert(num, gc.Equals, expected)
}

func (s *managedStorageSuite) TestPut(c *gc.C) {
	s.assertPut(c, "/path/to/blob", []byte("some resource"))
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutSamePathDifferentData(c *gc.C) {
	resPath := s.assertPut(c, "/path/to/blob", []byte("some resource"))
	secondResPath := s.assertPut(c, "/path/to/blob", []byte("another resource"))
	c.Assert(resPath, gc.Not(gc.Equals), secondResPath)
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutDifferentPathSameData(c *gc.C) {
	resPath := s.assertPut(c, "/path/to/blob", []byte("some resource"))
	secondResPath := s.assertPut(c, "/anotherpath/to/blob", []byte("some resource"))
	c.Assert(resPath, gc.Equals, secondResPath)
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutSamePathDifferentDataMulti(c *gc.C) {
	resPath := s.assertPut(c, "/path/to/blob", []byte("another resource"))
	secondResPath := s.assertPut(c, "/anotherpath/to/blob", []byte("some resource"))
	c.Assert(resPath, gc.Not(gc.Equals), secondResPath)
	s.assertResourceCatalogCount(c, 2)

	thirdResPath := s.assertPut(c, "/path/to/blob", []byte("some resource"))
	c.Assert(resPath, gc.Not(gc.Equals), secondResPath)
	c.Assert(secondResPath, gc.Equals, thirdResPath)
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutManagedResourceFail(c *gc.C) {
	var resourcePath string
	s.PatchValue(blobstore.PutResourceTxn, func(
		coll *mgo.Collection, managedResource blobstore.ManagedResource, resourceId string) (string, []txn.Op, error) {
		rc := blobstore.GetResourceCatalog(s.managedStorage)
		r, err := rc.Get(resourceId)
		c.Assert(err, gc.IsNil)
		resourcePath = r.Path
		return "", nil, errors.Errorf("some error")
	})
	// Attempt to put the data.
	blob := []byte("data")
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", "/some/path", rdr, int64(len(blob)))
	c.Assert(err, gc.ErrorMatches, "cannot update managed resource catalog: some error")

	// Now ensure there's no blob data left behind in storage, nor a resource catalog record.
	s.assertResourceCatalogCount(c, 0)
	_, err = s.resourceStorage.Get(resourcePath)
	c.Assert(err, gc.ErrorMatches, ".*not found")
}

func (s *managedStorageSuite) TestPutForEnvironmentAndCheckHash(c *gc.C) {
	blob := []byte("data")
	rdr := bytes.NewReader(blob)
	sha384Hash := calculateCheckSum(c, 0, 5, []byte("wrong"))
	err := s.managedStorage.PutForEnvironmentAndCheckHash("env", "/some/path", rdr, int64(len(blob)), sha384Hash)
	c.Assert(err, gc.ErrorMatches, "hash mismatch")

	rdr.Seek(0, 0)
	sha384Hash = calculateCheckSum(c, 0, int64(len(blob)), blob)
	err = s.managedStorage.PutForEnvironmentAndCheckHash("env", "/some/path", rdr, int64(len(blob)), sha384Hash)
	c.Assert(err, gc.IsNil)
}

func (s *managedStorageSuite) TestPutForEnvironmentAndCheckHashEmptyHash(c *gc.C) {
	// Passing "" as the hash to PutForEnvironmentAndCheckHash will elide
	// the hash check.
	rdr := strings.NewReader("data")
	err := s.managedStorage.PutForEnvironmentAndCheckHash("env", "/some/path", rdr, int64(rdr.Len()), "")
	c.Assert(err, jc.ErrorIsNil)
}

func (s *managedStorageSuite) TestPutForEnvironmentUnknownLen(c *gc.C) {
	// Passing -1 for the size of the data directs PutForEnvironment
	// to read in the whole amount.
	blob := []byte("data")
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", "/some/path", rdr, -1)
	c.Assert(err, jc.ErrorIsNil)
	s.assertGet(c, "/some/path", blob)
}

func (s *managedStorageSuite) TestPutForEnvironmentOverLong(c *gc.C) {
	// Passing a size to PutForEnvironment that exceeds the actual
	// size of the data will result in metadata recording the actual
	// size.
	blob := []byte("data")
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", "/some/path", rdr, int64(len(blob)+1))
	c.Assert(err, jc.ErrorIsNil)
	s.assertGet(c, "/some/path", blob)
}

func (s *managedStorageSuite) assertGet(c *gc.C, path string, blob []byte) {
	r, length, err := s.managedStorage.GetForEnvironment("env", path)
	c.Assert(err, gc.IsNil)
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	c.Assert(err, gc.IsNil)
	c.Assert(data, gc.DeepEquals, blob)
	c.Assert(int(length), gc.Equals, len(blob))
}

func (s *managedStorageSuite) TestGet(c *gc.C) {
	blob := []byte("some resource")
	s.assertPut(c, "/path/to/blob", blob)
	s.assertGet(c, "/path/to/blob", blob)
}

func (s *managedStorageSuite) TestGetNonExistent(c *gc.C) {
	_, _, err := s.managedStorage.GetForEnvironment("env", "/path/to/nowhere")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}

func (s *managedStorageSuite) TestRemove(c *gc.C) {
	blob := []byte("some resource")
	resPath := s.assertPut(c, "/path/to/blob", blob)
	err := s.managedStorage.RemoveForEnvironment("env", "/path/to/blob")
	c.Assert(err, gc.IsNil)

	// Check the data and catalog entry really are removed.
	_, _, err = s.managedStorage.GetForEnvironment("env", "path/to/blob")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
	_, err = s.resourceStorage.Get(resPath)
	c.Assert(err, gc.NotNil)

	s.assertResourceCatalogCount(c, 0)
}

func (s *managedStorageSuite) TestRemoveNonExistent(c *gc.C) {
	err := s.managedStorage.RemoveForEnvironment("env", "/path/to/nowhere")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}

func (s *managedStorageSuite) TestRemoveDifferentPathKeepsData(c *gc.C) {
	blob := []byte("some resource")
	s.assertPut(c, "/path/to/blob", blob)
	s.assertPut(c, "/anotherpath/to/blob", blob)
	s.assertResourceCatalogCount(c, 1)
	err := s.managedStorage.RemoveForEnvironment("env", "/path/to/blob")
	c.Assert(err, gc.IsNil)
	s.assertGet(c, "/anotherpath/to/blob", blob)
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutRace(c *gc.C) {
	blob := []byte("some resource")
	beforeFunc := func() {
		s.assertPut(c, "/path/to/blob", blob)
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFunc).Check()
	anotherblob := []byte("another resource")
	s.assertPut(c, "/path/to/blob", anotherblob)
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutDeleteRace(c *gc.C) {
	blob := []byte("some resource")
	s.assertPut(c, "/path/to/blob", blob)
	beforeFunc := func() {
		err := s.managedStorage.RemoveForEnvironment("env", "/path/to/blob")
		c.Assert(err, gc.IsNil)
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFunc).Check()
	anotherblob := []byte("another resource")
	s.assertPut(c, "/path/to/blob", anotherblob)
	s.assertResourceCatalogCount(c, 1)
}

func (s *managedStorageSuite) TestPutRaceWhereCatalogEntryRemoved(c *gc.C) {
	blob := []byte("some resource")
	// Remove the resource catalog entry with the resourceId that we are about
	// to write to a managed resource entry.
	beforeFunc := []func(){
		nil, //  resourceCatalog Put()
		nil, //  managedResource Put()
		func() {
			// Shamelessly exploit our knowledge of how ids are made.
			sha384Hash := calculateCheckSum(c, 0, int64(len(blob)), blob)
			_, _, err := blobstore.GetResourceCatalog(s.managedStorage).Remove(sha384Hash)
			c.Assert(err, gc.IsNil)
		},
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFunc...).Check()
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", "/path/to/blob", rdr, int64(len(blob)))
	c.Assert(err, gc.ErrorMatches, "unexpected deletion .*")
	s.assertResourceCatalogCount(c, 0)
}

func (s *managedStorageSuite) TestRemoveRace(c *gc.C) {
	blob := []byte("some resource")
	s.assertPut(c, "/path/to/blob", blob)
	beforeFunc := func() {
		err := s.managedStorage.RemoveForEnvironment("env", "/path/to/blob")
		c.Assert(err, gc.IsNil)
	}
	defer txntesting.SetBeforeHooks(c, s.txnRunner, beforeFunc).Check()
	err := s.managedStorage.RemoveForEnvironment("env", "/path/to/blob")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
	_, _, err = s.managedStorage.GetForEnvironment("env", "/path/to/blob")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}

func (s *managedStorageSuite) TestPutRequestNotFound(c *gc.C) {
	_, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", "sha384")
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
}

func (s *managedStorageSuite) putTestRandomBlob(c *gc.C, path string) (blob []byte, sha384HashHex string) {
	id := bson.NewObjectId().Hex()
	blob = []byte(id)
	return blob, s.putTestBlob(c, path, blob)
}

func (s *managedStorageSuite) putTestBlob(c *gc.C, path string, blob []byte) (sha384HashHex string) {
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", path, rdr, int64(len(blob)))
	c.Assert(err, gc.IsNil)
	s.assertGet(c, path, blob)
	sha384HashHex = calculateCheckSum(c, 0, int64(len(blob)), blob)
	return sha384HashHex
}

func calculateCheckSum(c *gc.C, start, length int64, blob []byte) (sha384HashHex string) {
	data := blob[start : start+length]
	sha384Hash := sha512.New384()
	_, err := sha384Hash.Write(data)
	c.Assert(err, gc.IsNil)
	sha384HashHex = fmt.Sprintf("%x", sha384Hash.Sum(nil))
	return sha384HashHex
}

func (s *managedStorageSuite) TestPutRequestResponseHashMismatch(c *gc.C) {
	_, sha384Hash := s.putTestRandomBlob(c, "path/to/blob")
	reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
	c.Assert(err, gc.IsNil)
	response := blobstore.NewPutResponse(reqResp.RequestId, "notsha384")
	err = s.managedStorage.ProofOfAccessResponse(response)
	c.Assert(err, gc.Equals, blobstore.ErrResponseMismatch)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

func (s *managedStorageSuite) assertPutRequestSingle(c *gc.C, blob []byte, resourceCount int) {
	if blob == nil {
		id := bson.NewObjectId().Hex()
		blob = []byte(id)
	}
	rdr := bytes.NewReader(blob)
	err := s.managedStorage.PutForEnvironment("env", "path/to/blob", rdr, int64(len(blob)))
	c.Assert(err, gc.IsNil)
	sha384Hash := calculateCheckSum(c, 0, int64(len(blob)), blob)
	reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
	c.Assert(err, gc.IsNil)
	sha384Response := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
	response := blobstore.NewPutResponse(reqResp.RequestId, sha384Response)
	err = s.managedStorage.ProofOfAccessResponse(response)
	c.Assert(err, gc.IsNil)
	s.assertGet(c, "path/to/blob", blob)
	s.assertResourceCatalogCount(c, resourceCount)
}

var trigger struct{} = struct{}{}

// patchedAfterFunc returns a function like time.AfterFunc, but is triggered on a channel
// select rather than the expiry of a timer interval.
func patchedAfterFunc(ch chan struct{}) func(d time.Duration, f func()) *time.Timer {
	return func(d time.Duration, f func()) *time.Timer {
		go func() {
			select {
			case <-ch:
				f()
				ch <- trigger
			}
		}()
		return nil
	}
}

func (s *managedStorageSuite) TestPutRequestSingle(c *gc.C) {
	ch := make(chan struct{})
	s.PatchValue(blobstore.AfterFunc, patchedAfterFunc(ch))
	s.assertPutRequestSingle(c, nil, 1)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	// Trigger the request timeout.
	ch <- trigger
	<-ch
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

func (s *managedStorageSuite) TestPutRequestLarge(c *gc.C) {
	ch := make(chan struct{})
	s.PatchValue(blobstore.AfterFunc, patchedAfterFunc(ch))
	// Use a blob size of 4096 which is greater than max range of put response range length.
	blob := make([]byte, 4096)
	for i := 0; i < 4096; i++ {
		blob[i] = byte(rand.Intn(255))
	}
	s.assertPutRequestSingle(c, blob, 1)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	// Trigger the request timeout.
	ch <- trigger
	<-ch
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

func (s *managedStorageSuite) TestPutRequestMultiSequential(c *gc.C) {
	ch := make(chan struct{})
	s.PatchValue(blobstore.AfterFunc, patchedAfterFunc(ch))
	s.assertPutRequestSingle(c, nil, 1)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	// Trigger the request timeout.
	ch <- trigger
	<-ch
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	s.assertPutRequestSingle(c, nil, 1)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	// Trigger the request timeout.
	ch <- trigger
	<-ch
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

func (s *managedStorageSuite) checkPutResponse(c *gc.C, index int, wg *sync.WaitGroup,
	requestId int64, sha384Hash string, blob []byte) {

	// After a random time, respond to a previously queued put request and check the result.
	go func() {
		delay := rand.Intn(3)
		time.Sleep(time.Duration(delay) * time.Millisecond)
		expectError := index == 2
		if expectError {
			sha384Hash = "bad"
		}
		response := blobstore.NewPutResponse(requestId, sha384Hash)
		err := s.managedStorage.ProofOfAccessResponse(response)
		if expectError {
			c.Check(err, gc.NotNil)
		} else {
			c.Check(err, gc.IsNil)
			if err == nil {
				r, length, err := s.managedStorage.GetForEnvironment("env", fmt.Sprintf("path/to/blob%d", index))
				c.Check(err, gc.IsNil)
				if err == nil {
					data, err := ioutil.ReadAll(r)
					c.Check(err, gc.IsNil)
					c.Check(data, gc.DeepEquals, blob)
					c.Check(int(length), gc.DeepEquals, len(blob))
				}
			}
		}
		wg.Done()
	}()
}

func (s *managedStorageSuite) queuePutRequests(c *gc.C, done chan struct{}) {
	var wg sync.WaitGroup
	// One request is allowed to expire so set up wait group for 1 less than number of requests.
	wg.Add(9)
	go func() {
		for i := 0; i < 10; i++ {
			blobPath := fmt.Sprintf("path/to/blob%d", i)
			blob, sha384Hash := s.putTestRandomBlob(c, blobPath)
			reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
			c.Assert(err, gc.IsNil)
			// Let one request timeout
			if i == 3 {
				continue
			}
			sha384Response := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
			s.checkPutResponse(c, i, &wg, reqResp.RequestId, sha384Response, blob)
		}
		wg.Wait()
		close(done)
	}()
}

const (
	ShortWait = 50 * time.Millisecond
	LongWait  = 10 * time.Second
)

var LongAttempt = &utils.AttemptStrategy{
	Total: LongWait,
	Delay: ShortWait,
}

func (s *managedStorageSuite) TestPutRequestMultiRandom(c *gc.C) {
	ch := make(chan struct{})
	s.PatchValue(blobstore.AfterFunc, patchedAfterFunc(ch))
	done := make(chan struct{})
	s.queuePutRequests(c, done)
	select {
	case <-done:
		c.Logf("all done")
	case <-time.After(LongWait):
		c.Fatalf("timed out waiting for put requests to be processed")
	}
	// One request hasn't been processed since we left it to timeout.
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 1)
	// Trigger the request timeout.
	ch <- trigger
	<-ch
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

func (s *managedStorageSuite) TestPutRequestExpired(c *gc.C) {
	ch := make(chan struct{})
	s.PatchValue(blobstore.AfterFunc, patchedAfterFunc(ch))
	blob, sha384Hash := s.putTestRandomBlob(c, "path/to/blob")
	reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
	c.Assert(err, gc.IsNil)
	sha384Response := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
	// Trigger the request timeout.
	ch <- trigger
	<-ch
	response := blobstore.NewPutResponse(reqResp.RequestId, sha384Response)
	err = s.managedStorage.ProofOfAccessResponse(response)
	c.Assert(err, gc.Equals, blobstore.ErrRequestExpired)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

// Run one simple test with the real time.AfterFunc to ensure it works.
func (s *managedStorageSuite) TestPutRequestExpiredWithRealTimeAfter(c *gc.C) {
	s.PatchValue(blobstore.RequestExpiry, 5*time.Millisecond)
	blob, sha384Hash := s.putTestRandomBlob(c, "path/to/blob")
	reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
	c.Assert(err, gc.IsNil)
	sha384Response := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
	// Wait for request timer to trigger.
	time.Sleep(7 * time.Millisecond)
	response := blobstore.NewPutResponse(reqResp.RequestId, sha384Response)
	err = s.managedStorage.ProofOfAccessResponse(response)
	c.Assert(err, gc.Equals, blobstore.ErrRequestExpired)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
}

func (s *managedStorageSuite) TestPutRequestExpiredMulti(c *gc.C) {
	ch := make(chan struct{})
	s.PatchValue(blobstore.AfterFunc, patchedAfterFunc(ch))
	blob, sha384Hash := s.putTestRandomBlob(c, "path/to/blob")
	reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
	c.Assert(err, gc.IsNil)
	sha384Response := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
	reqResp2, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob2", sha384Hash)
	c.Assert(err, gc.IsNil)
	sha384Response2 := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
	// Trigger the request timeouts.
	ch <- trigger
	<-ch
	ch <- trigger
	<-ch
	c.Assert(blobstore.RequestQueueLength(s.managedStorage), gc.Equals, 0)
	response := blobstore.NewPutResponse(reqResp.RequestId, sha384Response)
	response2 := blobstore.NewPutResponse(reqResp2.RequestId, sha384Response2)
	err = s.managedStorage.ProofOfAccessResponse(response)
	c.Assert(err, gc.Equals, blobstore.ErrRequestExpired)
	err = s.managedStorage.ProofOfAccessResponse(response2)
	c.Assert(err, gc.Equals, blobstore.ErrRequestExpired)
}

func (s *managedStorageSuite) TestPutRequestDeleted(c *gc.C) {
	blob, sha384Hash := s.putTestRandomBlob(c, "path/to/blob")
	reqResp, err := s.managedStorage.PutForEnvironmentRequest("env", "path/to/blob", sha384Hash)
	c.Assert(err, gc.IsNil)
	err = s.managedStorage.RemoveForEnvironment("env", "path/to/blob")
	c.Assert(err, gc.IsNil)

	sha384Response := calculateCheckSum(c, reqResp.RangeStart, reqResp.RangeLength, blob)
	response := blobstore.NewPutResponse(reqResp.RequestId, sha384Response)
	err = s.managedStorage.ProofOfAccessResponse(response)
	c.Assert(err, gc.Equals, blobstore.ErrResourceDeleted)
}

func (s *managedStorageSuite) TestPutMultiSameData(c *gc.C) {
	blob := bytes.Repeat([]byte("blobalob"), 1024*1024*10)
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()
				rdr := bytes.NewReader(blob)
				err := s.managedStorage.PutForEnvironment("env", "path", rdr, int64(len(blob)))
				c.Assert(err, gc.IsNil)
			}()
		}
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Minute):
		c.Fatalf("timed out waiting for puts to be processed")
	}
}
