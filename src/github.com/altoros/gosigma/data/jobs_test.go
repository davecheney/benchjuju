// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package data

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDataJobReaderFail(t *testing.T) {
	r := failReader{}

	if _, err := ReadJob(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}

	if _, err := ReadJob(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}
}

func TestDataJobUnmarshal(t *testing.T) {
	var j Job
	err := json.Unmarshal([]byte(jsonJobData), &j)
	if err != nil {
		t.Error(err)
	}

	compareJobs(t, &j, &jobData)
}

func TestDataJobReadJob(t *testing.T) {
	j, err := ReadJob(strings.NewReader(jsonJobData))
	if err != nil {
		t.Error(err)
	}

	compareJobs(t, j, &jobData)
}

func compareJobs(t *testing.T, value, wants *Job) {
	if value.Resource != wants.Resource {
		t.Errorf("Job.Resource error: found %#v, wants %#v", value.Resource, wants.Resource)
	}

	if len(value.Children) != len(wants.Children) {
		t.Errorf("Job.Children error: found %#v, wants %#v", value.Children, wants.Children)
	}
	for i := 0; i < len(value.Children); i++ {
		if value.Children[i] != wants.Children[i] {
			t.Errorf("Job.Children error [%d]: found %#v, wants %#v", i, value.Children[i], wants.Children[i])
		}
	}

	if value.Created.Equal(wants.Created) {
		t.Errorf("Job.Created error: found %#v, wants %#v", value.Created, wants.Created)
	}
	if value.Data != wants.Data {
		t.Errorf("Job.Data error: found %#v, wants %#v", value.Data, wants.Data)
	}
	if value.LastModified.Equal(wants.LastModified) {
		t.Errorf("Job.LastModified error: found %#v, wants %#v", value.LastModified, wants.LastModified)
	}
	if value.Operation != wants.Operation {
		t.Errorf("Job.Operation error: found %#v, wants %#v", value.Operation, wants.Operation)
	}

	if len(value.Resources) != len(wants.Resources) {
		t.Errorf("Job.Resources error: found %#v, wants %#v", value.Resources, wants.Resources)
	}
	for i := 0; i < len(value.Resources); i++ {
		if value.Resources[i] != wants.Resources[i] {
			t.Errorf("Job.Resources error [%d]: found %#v, wants %#v", i, value.Resources[i], wants.Resources[i])
		}
	}

	if value.State != wants.State {
		t.Errorf("Job.State error: found %#v, wants %#v", value.State, wants.State)
	}
}
