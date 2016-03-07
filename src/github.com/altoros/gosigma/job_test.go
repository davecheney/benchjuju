// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package gosigma

import (
	"testing"
	"time"

	"github.com/altoros/gosigma/data"
	"github.com/altoros/gosigma/mock"
)

func TestJobString(t *testing.T) {
	j := &job{obj: &data.Job{}}
	if s := j.String(); s != `{UUID: "", Operation: , State: , Progress: 0, Resources: []}` {
		t.Errorf("invalid Job.String(): `%s`", s)
	}

	j.obj.UUID = "uuid"
	if s := j.String(); s != `{UUID: "uuid", Operation: , State: , Progress: 0, Resources: []}` {
		t.Errorf("invalid Job.String(): `%s`", s)
	}

	j.obj.Operation = "operation"
	if s := j.String(); s != `{UUID: "uuid", Operation: operation, State: , Progress: 0, Resources: []}` {
		t.Errorf("invalid Job.String(): `%s`", s)
	}

	j.obj.State = JobStateStarted
	if s := j.String(); s != `{UUID: "uuid", Operation: operation, State: started, Progress: 0, Resources: []}` {
		t.Errorf("invalid Job.String(): `%s`", s)
	}

	j.obj.Data.Progress = 99
	if s := j.String(); s != `{UUID: "uuid", Operation: operation, State: started, Progress: 99, Resources: []}` {
		t.Errorf("invalid Job.String(): `%s`", s)
	}
}

func TestJobChildren(t *testing.T) {
	j := &job{obj: &data.Job{}}
	if c := j.Children(); c == nil || len(c) != 0 {
		t.Errorf("invalid Job.Children(): %v", c)
	}

	j.obj.Children = append(j.obj.Children, "child-0")
	if c := j.Children(); c == nil || len(c) != 1 || c[0] != "child-0" {
		t.Errorf("invalid Job.Children(): %v", c)
	}

	j.obj.Children = append(j.obj.Children, "child-1")
	if c := j.Children(); c == nil || len(c) != 2 || c[0] != "child-0" || c[1] != "child-1" {
		t.Errorf("invalid Job.Children(): %v", c)
	}
}

func TestJobCreated(t *testing.T) {
	j := &job{obj: &data.Job{}}
	if c := j.Created(); c != (time.Time{}) {
		t.Errorf("invalid Job.Time(): %v", c)
	}

	j.obj.Created = time.Unix(100, 200)
	if c := j.Created(); c != time.Unix(100, 200) {
		t.Errorf("invalid Job.Time(): %v", c)
	}
}

func TestJobLastModified(t *testing.T) {
	j := &job{obj: &data.Job{}}
	if c := j.LastModified(); c != (time.Time{}) {
		t.Errorf("invalid Job.LastModified(): %v", c)
	}

	j.obj.LastModified = time.Unix(100, 200)
	if c := j.LastModified(); c != time.Unix(100, 200) {
		t.Errorf("invalid Job.LastModified(): %v", c)
	}
}

func TestJobProgress(t *testing.T) {
	mock.Jobs.Reset()

	const uuid = "305867d6-5652-41d2-be5c-bbae1eed5676"

	jd := &data.Job{
		Resource:     *data.MakeJobResource(uuid),
		Created:      time.Date(2014, time.January, 30, 15, 24, 42, 205092, time.UTC),
		Data:         data.JobData{Progress: 97},
		LastModified: time.Date(2014, time.January, 30, 15, 24, 42, 937432, time.UTC),
		Operation:    "drive_clone",
		Resources: []string{
			"/api/2.0/drives/df05497c-1504-4fea-af24-2825fc5133cf/",
			"/api/2.0/drives/db7a095c-622d-4b98-88fd-25a7e34d402e/",
		},
		State: "success",
	}

	mock.Jobs.Add(jd)

	cli, err := createTestClient(t)
	if err != nil {
		t.Error(err)
		return
	}

	j := &job{
		client: cli,
		obj:    &data.Job{Resource: *data.MakeJobResource(uuid)},
	}

	if err := j.Refresh(); err != nil {
		t.Error(err)
		return
	}

	if p := j.Progress(); p != 97 {
		t.Error("invalid Refresh progress")
	}

	setJobProgress := func() {
		jd.Data.Progress = 100
	}
	go setJobProgress()

	if err := j.Wait(); err != nil {
		t.Error(err)
		return
	}
}

func TestJobProgressError(t *testing.T) {
	cli, err := NewClient("https://0.1.2.3:2000/api/2.0/", mock.TestUser, mock.TestPassword, nil)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	cli.ConnectTimeout(100 * time.Millisecond)
	cli.ReadWriteTimeout(100 * time.Millisecond)

	const uuid = "305867d6-5652-41d2-be5c-bbae1eed5676"
	j := &job{
		client: cli,
		obj:    &data.Job{Resource: *data.MakeJobResource(uuid)},
	}

	if err := j.Refresh(); err == nil {
		t.Error("Job.Refresh returned valid result for unavailable endpoint")
	} else {
		t.Log("OK: Job.Refresh()", err)
	}

	if err := j.Wait(); err == nil {
		t.Error("Job.Wait returned valid result for unavailable endpoint")
	} else {
		t.Log("OK: Job.Wait()", err)
	}
}

func TestJobWaitTimeout(t *testing.T) {
	cli, err := createTestClient(t)
	if err != nil || cli == nil {
		t.Error("NewClient() failed:", err, cli)
		return
	}

	if *trace {
		cli.Logger(t)
	}

	cli.ConnectTimeout(100 * time.Millisecond)
	cli.ReadWriteTimeout(100 * time.Millisecond)
	cli.OperationTimeout(1 * time.Millisecond)

	mock.Jobs.Reset()

	const uuid = "305867d6-5652-41d2-be5c-bbae1eed5676"

	jd := &data.Job{
		Resource:     *data.MakeJobResource(uuid),
		Created:      time.Date(2014, time.January, 30, 15, 24, 42, 205092, time.UTC),
		Data:         data.JobData{Progress: 97},
		LastModified: time.Date(2014, time.January, 30, 15, 24, 42, 937432, time.UTC),
		Operation:    "drive_clone",
		Resources: []string{
			"/api/2.0/drives/df05497c-1504-4fea-af24-2825fc5133cf/",
			"/api/2.0/drives/db7a095c-622d-4b98-88fd-25a7e34d402e/",
		},
		State: "started",
	}

	mock.Jobs.Add(jd)

	j := &job{
		client: cli,
		obj:    &data.Job{Resource: *data.MakeJobResource(uuid)},
	}

	if err := j.Wait(); err != ErrOperationTimeout {
		t.Log("invalid Job.Wait()", err)
	} else {
		t.Log("OK: Job.Wait()", err)
	}
}
