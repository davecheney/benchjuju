// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package data

import (
	"strings"
	"testing"
)

func TestDataErrorReaderFail(t *testing.T) {
	r := failReader{}
	if _, err := ReadError(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}
}

func TestDataErrorEmptyJson(t *testing.T) {
	r := strings.NewReader("")
	ee, err := ReadError(r)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ee) != 0 {
		t.Error("Invalid result")
		return
	}
}

func TestDataErrorEmptyArrayJson(t *testing.T) {
	r := strings.NewReader("[]")
	ee, err := ReadError(r)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ee) != 0 {
		t.Error("Invalid result")
		return
	}
}

func TestDataErrorEmptyObjectJson(t *testing.T) {
	r := strings.NewReader("{}")
	ee, err := ReadError(r)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ee) != 1 {
		t.Error("Invalid result")
		return
	}

	var e Error
	if e != ee[0] {
		t.Errorf("return %#v, expected %#v", ee[0], e)
	}
}

func TestDataErrorInvalidArrayJson(t *testing.T) {
	r := strings.NewReader("[xxx]")
	_, err := ReadError(r)
	if err == nil {
		t.Error("should fail")
	}
	t.Log("OK:", err)
}

func TestDataErrorInvalidObjectJson(t *testing.T) {
	r := strings.NewReader("{xxx}")
	_, err := ReadError(r)
	if err == nil {
		t.Error("should fail")
	}
	t.Log("OK:", err)
}

func TestDataErrorArrayJson(t *testing.T) {
	const jsonErrorArray = `[
		{
		"error_point": null,
	 	"error_type": "permission",
	 	"error_message": "Cannot start guest in state \"starting\". Guest should be in state \"stopped\""
		}
	]`

	r := strings.NewReader(jsonErrorArray)
	ee, err := ReadError(r)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ee) != 1 {
		t.Error("Invalid result")
		return
	}

	e := ee[0]
	if e.Point != "" {
		t.Error("e.Point must be empty")
	}
	if e.Type != "permission" {
		t.Error("invalid e.Type")
	}
	if e.Message != "Cannot start guest in state \"starting\". Guest should be in state \"stopped\"" {
		t.Error("invalid e.Message")
	}
}

func TestDataErrorObjectJson(t *testing.T) {
	const jsonErrorObject = `{
		"error_point": null,
	 	"error_type": "permission",
	 	"error_message": "Cannot start guest in state \"starting\". Guest should be in state \"stopped\""
		}`

	r := strings.NewReader(jsonErrorObject)
	ee, err := ReadError(r)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ee) != 1 {
		t.Error("Invalid result")
		return
	}

	e := ee[0]
	if e.Point != "" {
		t.Errorf("e.Point must be empty, value: %#v", e)
	}
	if e.Type != "permission" {
		t.Errorf("invalid e.Type, value: %#v", e)
	}
	if e.Message != "Cannot start guest in state \"starting\". Guest should be in state \"stopped\"" {
		t.Errorf("invalid e.Message, value: %#v", e)
	}
}
