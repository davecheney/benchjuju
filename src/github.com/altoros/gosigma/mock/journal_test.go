// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMockJournalExternalID(t *testing.T) {

	t.Parallel()

	req, err := http.NewRequest("GET", "https://localhost/test", nil)
	if err != nil {
		t.Error(err)
	}
	rec := httptest.NewRecorder()

	id := genID()
	SetID(req.Header, id)

	recordJournal("test", req, rec)

	jee := GetJournal(id)
	if len(jee) != 1 {
		t.Error("len([]JournalEntry) != 1")
	}
	je := jee[0]
	if je.Name != "test" {
		t.Error("JournalEntry.Name failed")
	}
	if je.Request != req {
		t.Error("JournalEntry.Request failed")
	}
	if je.Response != rec {
		t.Error("JournalEntry.Response failed")
	}
}

func TestMockJournalAutoID(t *testing.T) {

	t.Parallel()

	req, err := http.NewRequest("GET", "https://localhost/test", nil)
	if err != nil {
		t.Error(err)
	}
	rec := httptest.NewRecorder()

	recordJournal("test", req, rec)

	id, err := GetID(rec.Header())
	if err != nil {
		t.Error(err)
	}

	jee := GetJournal(id)
	if len(jee) != 1 {
		t.Error("len([]JournalEntry) != 1")
	}
	je := jee[0]
	if je.Name != "test" {
		t.Error("JournalEntry.Name failed")
	}
	if je.Request != req {
		t.Error("JournalEntry.Request failed")
	}
	if je.Response != rec {
		t.Error("JournalEntry.Response failed")
	}
}
