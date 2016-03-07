// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package txn_test

import (
	"time"

	"github.com/juju/loggo"
	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/txn"

	jujutxn "github.com/juju/txn"
)

type PruneSuite struct {
	jujutesting.MgoSuite
	db     *mgo.Database
	txns   *mgo.Collection
	runner *txn.Runner
}

var _ = gc.Suite(&PruneSuite{})

func (s *PruneSuite) SetUpTest(c *gc.C) {
	s.MgoSuite.SetUpTest(c)
	txn.SetChaos(txn.Chaos{})

	s.db = s.Session.DB("prune-test")
	s.txns = s.db.C("txns")
	s.runner = txn.NewRunner(s.txns)
}

func (s *PruneSuite) TearDownSuite(c *gc.C) {
	txn.SetChaos(txn.Chaos{})
	s.MgoSuite.TearDownSuite(c)
}

func (s *PruneSuite) maybePrune(c *gc.C, pruneFactor float32) {
	r := jujutxn.NewRunner(jujutxn.RunnerParams{
		Database:                  s.db,
		TransactionCollectionName: s.txns.Name,
		ChangeLogName:             s.txns.Name + ".log",
	})
	err := r.MaybePruneTransactions(pruneFactor)
	c.Assert(err, jc.ErrorIsNil)
}

func (s *PruneSuite) TestSingleCollection(c *gc.C) {
	// Create some simple transactions, keeping track of the last
	// transaction id for each document.
	const numDocs = 5
	const updatesPerDoc = 3

	lastTxnIds := make([]bson.ObjectId, numDocs)
	for id := 0; id < numDocs; id++ {
		s.runTxn(c, txn.Op{
			C:      "coll",
			Id:     id,
			Insert: bson.M{},
		})

		for txnNum := 0; txnNum < updatesPerDoc; txnNum++ {
			lastTxnIds[id] = s.runTxn(c, txn.Op{
				C:      "coll",
				Id:     id,
				Update: bson.M{},
			})
		}
	}

	// Ensure that expected number of transactions were created.
	s.assertCollCount(c, "txns", numDocs+(numDocs*updatesPerDoc))

	s.maybePrune(c, 1)

	// Confirm that only the records for the most recent transactions
	// for each document were kept.
	s.assertTxns(c, lastTxnIds...)

	// Run another transaction on each of the docs to ensure mgo/txn
	// is happy.
	for id := 0; id < numDocs; id++ {
		s.runTxn(c, txn.Op{
			C:      "coll",
			Id:     id,
			Update: bson.M{},
		})
	}
}

func (s *PruneSuite) TestMultipleDocumentsInOneTxn(c *gc.C) {
	// Create two documents each in their own txn.
	s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Insert: bson.M{},
	})
	s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     1,
		Insert: bson.M{},
	})

	// Now update both documents in one transaction.
	txnId := s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Update: bson.M{},
	}, txn.Op{
		C:      "coll",
		Id:     1,
		Update: bson.M{},
	})

	s.maybePrune(c, 1)

	// Only the last transaction should be left.
	s.assertTxns(c, txnId)
}

func (s *PruneSuite) TestMultipleCollections(c *gc.C) {
	var lastTxnIds []bson.ObjectId

	// Create a single document.
	s.runTxn(c, txn.Op{
		C:      "coll0",
		Id:     0,
		Insert: bson.M{},
	})

	// Update that document and create two more in other collections,
	// all in one txn. This will be the last txn that touches coll0/0
	// so it should not be pruned.
	txnId := s.runTxn(c, txn.Op{
		C:      "coll0",
		Id:     0,
		Update: bson.M{},
	}, txn.Op{
		C:      "coll1",
		Id:     0,
		Insert: bson.M{},
	}, txn.Op{
		C:      "coll2",
		Id:     0,
		Insert: bson.M{},
	})
	lastTxnIds = append(lastTxnIds, txnId)

	// Update coll1 and coll2 docs together. This will be the last txn
	// to touch coll1/0 and coll2/0 so it should not be pruned.
	txnId = s.runTxn(c, txn.Op{
		C:      "coll1",
		Id:     0,
		Update: bson.M{},
	}, txn.Op{
		C:      "coll2",
		Id:     0,
		Update: bson.M{},
	})
	lastTxnIds = append(lastTxnIds, txnId)

	// Insert more documents into coll0 and coll1.
	txnId = s.runTxn(c, txn.Op{
		C:      "coll0",
		Id:     1,
		Insert: bson.M{},
	}, txn.Op{
		C:      "coll1",
		Id:     1,
		Insert: bson.M{},
	})
	lastTxnIds = append(lastTxnIds, txnId)

	s.maybePrune(c, 1)
	s.assertTxns(c, lastTxnIds...)
}

func (s *PruneSuite) TestWithStash(c *gc.C) {
	// Ensure that txns referenced in the stash are not pruned from
	// the txns collection.

	// An easy way to get something into the stash is to delete a
	// document.
	txnId0 := s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Insert: bson.D{},
	})
	txnId1 := s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Remove: true,
	})
	s.assertCollCount(c, "txns.stash", 1)

	s.maybePrune(c, 1)
	s.assertTxns(c, txnId0, txnId1)
}

func (s *PruneSuite) TestInProgressInsertNotPruned(c *gc.C) {
	// Create an incomplete insert transaction.
	txn.SetChaos(txn.Chaos{
		KillChance: 1,
		Breakpoint: "set-applying",
	})
	txnId := s.runFailingTxn(c, txn.ErrChaos, txn.Op{
		C:      "coll",
		Id:     0,
		Insert: bson.M{},
	})

	// There will be in-progress txns and txns.stash entries for the
	// new document now. Remove the stash entry to simulate the point
	// in time where the txns doc has been inserted but the txns.stash
	// doc hasn't yet.
	err := s.db.C("txns.stash").RemoveId(bson.D{
		{"c", "coll"},
		{"id", 0},
	})
	c.Assert(err, jc.ErrorIsNil)

	s.maybePrune(c, 1)
	s.assertTxns(c, txnId)
}

func (s *PruneSuite) TestInProgressUpdateNotPruned(c *gc.C) {
	// Create an insert transaction and then in-progress update
	// transaction.
	txnIdInsert := s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Insert: bson.M{},
	})

	txn.SetChaos(txn.Chaos{
		KillChance: 1,
		Breakpoint: "set-prepared",
	})
	txnIdUpdate := s.runFailingTxn(c, txn.ErrChaos, txn.Op{
		C:      "coll",
		Id:     0,
		Update: bson.M{},
	})

	// Remove reference to the update transaction from the doc to
	// simulate the point in time where the txns doc has been created
	// but it's not referenced from the doc being updated yet.
	coll := s.db.C("coll")
	err := coll.UpdateId(0, bson.M{
		"$pull": bson.M{"txn-queue": bson.M{"$regex": "^" + txnIdUpdate.Hex() + "_*"}},
	})
	c.Assert(err, jc.ErrorIsNil)

	s.maybePrune(c, 1)
	s.assertTxns(c, txnIdInsert, txnIdUpdate)
}

func (s *PruneSuite) TestAbortedTxnsArePruned(c *gc.C) {
	// Create an insert transaction, then an aborted transaction
	// and then a successful transaction, all for same doc.
	s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Insert: bson.M{},
	})
	s.runFailingTxn(c, txn.ErrAborted, txn.Op{
		C:      "coll",
		Id:     0,
		Assert: txn.DocMissing, // Aborts because doc is already there.
	})
	txnId := s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Update: bson.M{},
	})

	s.maybePrune(c, 1)
	s.assertTxns(c, txnId)
}

func (s *PruneSuite) TestManyTxnRemovals(c *gc.C) {
	// This is mainly to test the chunking of removals.
	s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     0,
		Insert: bson.M{},
	})
	var lastTxnId bson.ObjectId
	for i := 0; i < 3000; i++ {
		lastTxnId = s.runTxn(c, txn.Op{
			C:      "coll",
			Id:     0,
			Update: bson.M{},
		})
	}
	s.assertCollCount(c, "txns", 3001)

	s.maybePrune(c, 1)
	s.assertTxns(c, lastTxnId)
}

func (s *PruneSuite) TestFirstRun(c *gc.C) {
	// When there's no pruning stats recorded pruning should always
	// happen.
	s.makeTxnsForNewDoc(c, 10)
	s.assertCollCount(c, "txns", 10)

	s.maybePrune(c, 2.0)

	s.assertCollCount(c, "txns", 1)
	s.assertLastPruneStats(c, 10, 1)
	s.assertPruneStatCount(c, 1)
}

func (s *PruneSuite) TestPruningRequired(c *gc.C) {
	s.makeTxnsForNewDoc(c, 5)
	s.makeTxnsForNewDoc(c, 5)
	s.assertCollCount(c, "txns", 10)

	// Fake that the last txns size was 3 documents so that pruning
	// should be triggered (10 >= 3 * 2.0).
	s.setLastPruneCount(c, 3)

	s.maybePrune(c, 2.0)

	s.assertCollCount(c, "txns", 2)
	s.assertLastPruneStats(c, 10, 2)
	s.assertPruneStatCount(c, 2)
}

func (s *PruneSuite) TestPruningNotRequired(c *gc.C) {
	s.makeTxnsForNewDoc(c, 10)
	s.assertCollCount(c, "txns", 10)

	// Set the last txns count such that pruning won't be triggered
	// with a factor of 2.0  (6 * 2.0 > 10).
	s.setLastPruneCount(c, 6)

	s.maybePrune(c, 2.0)

	// Pruning shouldn't have happened.
	s.assertCollCount(c, "txns", 10)
	s.assertPruneStatCount(c, 1)
}

func (s *PruneSuite) TestPruningStatsHistory(c *gc.C) {
	s.maybePrune(c, 2.0)
	s.assertLastPruneStats(c, 0, 0)
	s.assertPruneStatCount(c, 1)

	s.makeTxnsForNewDoc(c, 5)

	s.maybePrune(c, 2.0)
	s.assertLastPruneStats(c, 5, 1)
	s.assertPruneStatCount(c, 2)

	s.makeTxnsForNewDoc(c, 11)

	s.maybePrune(c, 2.0)
	s.assertLastPruneStats(c, 12, 2)
	s.assertPruneStatCount(c, 3)

	s.makeTxnsForNewDoc(c, 5)

	s.maybePrune(c, 2.0)
	s.assertLastPruneStats(c, 7, 3)
	s.assertPruneStatCount(c, 4)
}

func (s *PruneSuite) TestPruningStatsBrokenLastPointer(c *gc.C) {
	// Create an initial pruning stats record.
	s.maybePrune(c, 2.0)
	s.assertLastPruneStats(c, 0, 0)
	s.assertPruneStatCount(c, 1)

	// Point the "last" pointer to a non-existent id.
	err := s.db.C("txns.prune").UpdateId("last", bson.M{
		"$set": bson.M{"id": bson.NewObjectId()},
	})
	c.Assert(err, jc.ErrorIsNil)

	var tw loggo.TestWriter
	c.Assert(loggo.RegisterWriter("test", &tw, loggo.WARNING), gc.IsNil)
	defer loggo.RemoveWriter("test")

	// Pruning should occur when "last" pointer is broken.
	s.maybePrune(c, 2.0)
	s.assertPruneStatCount(c, 2) // Note the new pruning stats record.
	c.Assert(tw.Log(), jc.LogMatches,
		[]jc.SimpleMessage{{loggo.WARNING, `pruning stats pointer was broken .+`}})
}

func (s *PruneSuite) makeTxnsForNewDoc(c *gc.C, count int) {
	id := bson.NewObjectId()
	s.runTxn(c, txn.Op{
		C:      "coll",
		Id:     id,
		Insert: bson.M{},
	})
	for i := 0; i < count-1; i++ {
		s.runTxn(c, txn.Op{
			C:      "coll",
			Id:     id,
			Update: bson.M{},
		})
	}
}

func (s *PruneSuite) runTxn(c *gc.C, ops ...txn.Op) bson.ObjectId {
	txnId := bson.NewObjectId()
	err := s.runner.Run(ops, txnId, nil)
	c.Assert(err, jc.ErrorIsNil)
	return txnId
}

func (s *PruneSuite) runFailingTxn(c *gc.C, expectedErr error, ops ...txn.Op) bson.ObjectId {
	txnId := bson.NewObjectId()
	err := s.runner.Run(ops, txnId, nil)
	c.Assert(err, gc.Equals, expectedErr)
	return txnId
}

func (s *PruneSuite) assertTxns(c *gc.C, expectedIds ...bson.ObjectId) {
	var actualIds []bson.ObjectId
	var txnDoc struct {
		Id bson.ObjectId `bson:"_id"`
	}
	iter := s.txns.Find(nil).Select(bson.M{"_id": 1}).Iter()
	for iter.Next(&txnDoc) {
		actualIds = append(actualIds, txnDoc.Id)
	}
	c.Assert(actualIds, jc.SameContents, expectedIds)
}

func (s *PruneSuite) assertCollCount(c *gc.C, collName string, expectedCount int) {
	count := s.getCollCount(c, collName)
	c.Assert(count, gc.Equals, expectedCount)
}

func (s *PruneSuite) getCollCount(c *gc.C, collName string) int {
	n, err := s.db.C(collName).Count()
	c.Assert(err, jc.ErrorIsNil)
	return n
}

func (s *PruneSuite) setLastPruneCount(c *gc.C, count int) {
	id := bson.NewObjectId()
	err := s.db.C("txns.prune").Insert(bson.M{
		"_id":        id,
		"txns-after": count,
	})
	c.Assert(err, jc.ErrorIsNil)

	err = s.db.C("txns.prune").Insert(bson.M{
		"_id": "last",
		"id":  id,
	})
	c.Assert(err, jc.ErrorIsNil)
}

func (s *PruneSuite) assertLastPruneStats(c *gc.C, txnsBefore, txnsAfter int) {
	txnsPrune := s.db.C("txns.prune")
	var doc bson.M

	err := txnsPrune.FindId("last").One(&doc)
	c.Assert(err, jc.ErrorIsNil)

	err = txnsPrune.FindId(doc["id"]).One(&doc)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(doc["txns-before"].(int), gc.Equals, txnsBefore)
	c.Assert(doc["txns-after"].(int), gc.Equals, txnsAfter)

	started := doc["started"].(time.Time)
	completed := doc["completed"].(time.Time)
	c.Assert(completed.Sub(started) >= time.Duration(0), jc.IsTrue)
	assertTimeIsRecent(c, started)
	assertTimeIsRecent(c, completed)
}

func (s *PruneSuite) assertPruneStatCount(c *gc.C, expected int) {
	txnsPrune := s.db.C("txns.prune")
	actual, err := txnsPrune.Count()
	c.Assert(err, jc.ErrorIsNil)

	actual-- // Ignore "last" pointer document
	c.Assert(actual, gc.Equals, expected)
}

func assertTimeIsRecent(c *gc.C, t time.Time) {
	c.Assert(time.Now().Sub(t), jc.LessThan, time.Hour)
}
