// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"errors"
	"net/http"
	"testing"

	"github.com/altoros/gosigma/https"
	"github.com/altoros/gosigma/https/httpstest"
)

func TestErrorNilResponse(t *testing.T) {
	enil := NewError(nil, nil)
	if enil != nil {
		t.Error("Error must be nil")
	}

	e := NewError(nil, errors.New("test error"))
	if e == nil {
		t.Error("Error must not be nil")
	}

	if e.SystemError == nil {
		t.Error("Invalid gosigma error created")
	} else if e.Error() != "test error" {
		t.Error("Invalid gosigma error created")
	}

	r, err := httpstest.CreateResponse(0)
	if err != nil {
		t.Error(err)
		return
	}
	ee := NewError(&https.Response{Response: r}, nil)
	if ee == nil {
		t.Error("Error must not be nil")
	}
	if ee.Error() != "" {
		t.Errorf("Error message must be empty, returned: %s", ee.Error())
	}
}

func TestErrorStatusCode(t *testing.T) {
	r, err := httpstest.CreateResponse(200)
	if err != nil {
		t.Error(err)
		return
	}

	e := NewError(&https.Response{Response: r}, nil)
	if e.SystemError != nil {
		t.Error("e.SystemError must be nil")
	}
	if e.StatusCode != 200 {
		t.Errorf("e.StatusCode == %d, wants 200", e.StatusCode)
	}
	if e.StatusMessage != "200 OK" {
		t.Errorf("e.StatusCode == %s, wants \"200 OK\"", e.StatusMessage)
	}
	if e.ServiceError != nil {
		t.Error("e.ServiceError must be nil")
	}

	if e.Error() != "200 OK" {
		t.Error("Error must return HTTP status message via error interface")
	}
}

func TestErrorSystemError(t *testing.T) {
	r, err := httpstest.CreateResponse(200)
	if err != nil {
		t.Error(err)
		return
	}

	e := NewError(&https.Response{Response: r}, errors.New("test"))
	if e.SystemError == nil {
		t.Error("e.SystemError must not be nil")
	}
	if e.StatusCode != 200 {
		t.Errorf("e.StatusCode == %d, wants 200", e.StatusCode)
	}
	if e.StatusMessage != "200 OK" {
		t.Errorf("e.StatusCode == %s, wants \"200 OK\"", e.StatusMessage)
	}
	if e.ServiceError != nil {
		t.Error("e.ServiceError must be nil")
	}

	if e.Error() != "test" {
		t.Error("Error must return system error message via error interface")
	}
}

func TestErrorSystemErrorHTTPError(t *testing.T) {
	r, err := httpstest.CreateResponse(404)
	if err != nil {
		t.Error(err)
		return
	}

	e := NewError(&https.Response{Response: r}, errors.New("test"))
	if e.SystemError == nil {
		t.Error("e.SystemError must not be nil")
	}
	if e.StatusCode != 404 {
		t.Errorf("e.StatusCode == %d, wants 404", e.StatusCode)
	}

	msg := "404 " + http.StatusText(404)
	if e.StatusMessage != msg {
		t.Errorf("e.StatusCode == %s, wants \"%s\"", e.StatusMessage, msg)
	}
	if e.ServiceError != nil {
		t.Error("e.ServiceError must be nil")
	}

	if e.Error() != msg {
		t.Error("Error must return HTTP status message via error interface")
	}
}

func TestErrorHTTPServiceError(t *testing.T) {
	var s = `[{"error_point": null, "error_type": "notexist", "error_message": "Object with uuid 472835d5-2bbb-4d87-9d08-7364bc373692 does not exist"}]`
	r, err := httpstest.CreateResponseWithBody(404, "application/json; charset=utf-8", s)
	if err != nil {
		t.Error(err)
		return
	}

	e := NewError(&https.Response{Response: r}, errors.New("test"))
	if e.SystemError == nil {
		t.Error("e.SystemError must not be nil")
	}
	if e.StatusCode != 404 {
		t.Errorf("e.StatusCode == %d, wants 404", e.StatusCode)
	}

	msg := "404 " + http.StatusText(404)
	if e.StatusMessage != msg {
		t.Errorf("e.StatusCode == %s, wants \"%s\"", e.StatusMessage, msg)
	}
	if e.ServiceError == nil {
		t.Error("e.ServiceError must not be nil")
	}

	emsg := msg + ", notexist, Object with uuid 472835d5-2bbb-4d87-9d08-7364bc373692 does not exist"
	if e.Error() != emsg {
		t.Errorf("Error must return service error message via error interface, ret: %s, wants: %s", e.Error(), emsg)
	}
}
