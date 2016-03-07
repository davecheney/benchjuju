// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package mock

import (
	"bytes"
	"testing"
)

func init() {
	Start()
}

func TestMockAuth(t *testing.T) {

	t.Parallel()

	check := func(u, p string, wants int) {
		resp, err := GetAuth("capabilities", u, p)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != wants {
			t.Errorf("Status = %d, wants %d", resp.StatusCode, wants)
		}
	}

	go check("", "", 401)
	go check(TestUser, "", 401)
	go check(TestUser, "1", 401)
	go check(TestUser+"1", TestPassword, 401)
	go check(TestUser, TestPassword+"1", 401)
	go check(TestUser, TestPassword, 200)
}

func TestMockSections(t *testing.T) {

	t.Parallel()

	ch := make(chan int)

	check := func(s string) {
		resp, err := Get(s)
		if err != nil {
			t.Errorf("Section %s: %s", s, err)
		}
		defer resp.Body.Close()

		LogResponse(t, resp)

		if resp.StatusCode != 200 {
			t.Errorf("Section %s: %s", s, resp.Status)
		}

		id := GetIDFromResponse(resp)
		jj := GetJournal(id)
		if len(jj) == 0 {
			t.Errorf("Section %s: journal length is zero", s)
		}

		j := jj[0]

		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)

		if !bytes.Equal(j.Response.Body.Bytes(), buf.Bytes()) {
			t.Errorf("Section: body error")
		}

		ch <- 1
	}

	const sectionCount = 5
	go check("capabilities")
	go check("drives")
	go check("libdrives")
	go check("drives")
	go check("jobs")

	var s int
	for s < sectionCount {
		s += <-ch
	}
}

/*
func TestHeaders(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	url := "https://localhost/headers.php"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}

	req.SetBasicAuth("test@example.com", "test")
	s := req.Header.Get("Authorization")
	t.Log("Auth:", s)

	resp, _ := client.Do(req)
	body, err := httputil.DumpResponse(resp, true)
	t.Log(string(body))
}
*/
