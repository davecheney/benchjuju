// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package httpstest

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

func TestCreateResponse(t *testing.T) {
	var codes = []int{200, 300, 301, 302, 303, 304,
		400, 401, 402, 403, 404, 405,
		500, 501, 502, 503, 504}
	for _, code := range codes {
		resp, err := CreateResponse(code)
		if err != nil {
			t.Error(err)
			continue
		}
		if resp.StatusCode != code {
			t.Errorf("Response status code == %d, waits %d", resp.StatusCode, code)
		}
		s := strconv.Itoa(code) + " " + http.StatusText(code)
		if resp.Status != s {
			t.Errorf("Response status code == %s, waits %s", resp.Status, s)
		}
	}
}

func TestCreateResponseWithType(t *testing.T) {
	var data = []string{
		"",
		"application/json",
		"application/json; charset=utf-8",
	}
	for _, contentType := range data {
		resp, err := CreateResponseWithType(200, contentType)
		if err != nil {
			t.Error(err)
			return
		}

		if resp.StatusCode != 200 {
			t.Errorf("Response status code == %d, waits %d", resp.StatusCode, 200)
		}

		msg := "200 " + http.StatusText(200)
		if resp.Status != msg {
			t.Errorf("Response status code == %s, waits %s", resp.Status, msg)
		}

		h := resp.Header
		vv, ok := h["Content-Type"]
		if !ok {
			t.Error("Content-Type header not found")
		}
		if len(vv) < 1 {
			t.Error("Content-Type header length < 1")
		}

		v := vv[0]
		if v != contentType {
			t.Errorf("Invalid Content-Type = %s, wants %s", v, contentType)
		}
	}
}

func TestCreateResponseWithData(t *testing.T) {
	var s = `[{"error_point": null, "error_type": "notexist", "error_message": "Object with uuid 472835d5-2bbb-4d87-9d08-7364bc373692 does not exist"}]`
	resp, err := CreateResponseWithBody(200, "application/json; charset=utf-8", s)
	if err != nil {
		t.Error(err)
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("Response status code == %d, waits %d", resp.StatusCode, 200)
	}

	msg := "200 " + http.StatusText(200)
	if resp.Status != msg {
		t.Errorf("Response status code == %s, waits %s", resp.Status, msg)
	}

	h := resp.Header
	vv, ok := h["Content-Type"]
	if !ok {
		t.Error("Content-Type header not found")
	}
	if len(vv) < 1 {
		t.Error("Content-Type header length < 1")
	}

	v := vv[0]
	if v != "application/json; charset=utf-8" {
		t.Errorf("Invalid Content-Type = %s, wants %s", v, "application/json; charset=utf-8")
	}

	if resp.Body == nil {
		t.Error("Response body is nil")
		return
	}

	bb, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		t.Error("Error reading from body:", err)
		return
	}

	body := string(bb)
	if body != s {
		t.Errorf("Body set:%s Body read: %s", s, body)
	}
}
