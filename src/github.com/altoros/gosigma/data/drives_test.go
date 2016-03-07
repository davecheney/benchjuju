// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestDataDriversReaderFail(t *testing.T) {
	r := failReader{}

	if _, err := ReadDrive(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}

	if _, err := ReadDrives(r); err == nil || err.Error() != "test error" {
		t.Error("Fail")
	}
}

func TestDataDrivesUnmarshal(t *testing.T) {
	var dd Drives
	dd.Meta.Limit = 12345
	dd.Meta.Offset = 12345
	dd.Meta.TotalCount = 12345
	err := json.Unmarshal([]byte(jsonDrivesData), &dd)
	if err != nil {
		t.Error(err)
	}

	verifyMeta(t, &dd.Meta, 0, 0, 9)

	for i := 0; i < len(drivesData); i++ {
		compareDrives(t, i, &dd.Objects[i], &drivesData[i])
	}
}

func TestDataDrivesDetailUnmarshal(t *testing.T) {
	var dd Drives
	dd.Meta.Limit = 12345
	dd.Meta.Offset = 12345
	dd.Meta.TotalCount = 12345
	err := json.Unmarshal([]byte(jsonDrivesDetailData), &dd)
	if err != nil {
		t.Error(err)
	}

	verifyMeta(t, &dd.Meta, 0, 0, 9)

	for i := 0; i < len(drivesDetailData); i++ {
		compareDrives(t, i, &dd.Objects[i], &drivesDetailData[i])
	}
}

func TestDataDrivesReadDrives(t *testing.T) {
	dd, err := ReadDrives(strings.NewReader(jsonDrivesData))
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < len(drivesData); i++ {
		compareDrives(t, i, &dd[i], &drivesData[i])
	}
}

func TestDataDrivesReadDrive(t *testing.T) {
	d, err := ReadDrive(strings.NewReader(jsonDriveData))
	if err != nil {
		t.Error(err)
	}

	compareDrives(t, 0, d, &driveData)
}

func TestDataDrivesLibDrive(t *testing.T) {
	d, err := ReadDrive(strings.NewReader(jsonLibraryDriveData))
	if err != nil {
		t.Error(err)
	}

	compareDrives(t, 0, d, &libraryDriveData)
}

func compareDrives(t *testing.T, i int, value, wants *Drive) {
	if value.Resource != wants.Resource {
		t.Errorf("Drive.Resource error [%d]: found %#v, wants %#v", i, value.Resource, wants.Resource)
	}

	if len(value.Affinities) != len(wants.Affinities) {
		t.Errorf("Drive.Affinities error [%d]: found %#v, wants %#v", i, value.Affinities, wants.Affinities)
	} else {
		for j := 0; j < len(value.Affinities); j++ {
			v := value.Affinities[j]
			w := wants.Affinities[j]
			if v != w {
				t.Errorf("Drive.Affinities error [%d]: at %d found %#v, wants %#v", i, j, v, w)
			}
		}
	}

	if value.AllowMultimount != wants.AllowMultimount {
		t.Errorf("Drive.AllowMultimount error [%d]: found %#v, wants %#v", i, value.AllowMultimount, wants.AllowMultimount)
	}

	if len(value.Jobs) != len(wants.Jobs) {
		t.Errorf("Drive.Jobs error [%d]: found %#v, wants %#v", i, value.Jobs, wants.Jobs)
	} else {
		for j := 0; j < len(value.Jobs); j++ {
			v := value.Jobs[j]
			w := wants.Jobs[j]
			if v != w {
				t.Errorf("Drive.Jobs error [%d]: at %d found %#v, wants %#v", i, j, v, w)
			}
		}
	}

	if value.Media != wants.Media {
		t.Errorf("Drive.Media error [%d]: found %#v, wants %#v", i, value.Media, wants.Media)
	}

	compareMeta(t, fmt.Sprintf("Drive.Meta error [%d]", i), value.Meta, wants.Meta)

	if value.Name != wants.Name {
		t.Errorf("Drive.Name error [%d]: found %#v, wants %#v", i, value.Name, wants.Name)
	}

	if value.Owner != nil && wants.Owner != nil {
		if *value.Owner != *wants.Owner {
			t.Errorf("Drive.Owner error [%d]: found %#v, wants %#v", i, value.Owner, wants.Owner)
		}
	} else if value.Owner != nil || wants.Owner != nil {
		t.Errorf("Drive.Owner error [%d]: found %#v, wants %#v", i, value.Owner, wants.Owner)
	}

	if value.Size != wants.Size {
		t.Errorf("Drive.Size error [%d]: found %#v, wants %#v", i, value.Size, wants.Size)
	}

	if value.Status != wants.Status {
		t.Errorf("Drive.Status error [%d]: found %#v, wants %#v", i, value.Status, wants.Status)
	}

	if value.StorageType != wants.StorageType {
		t.Errorf("Drive.StorageType error [%d]: found %#v, wants %#v", i, value.StorageType, wants.StorageType)
	}

	//
	// specific for media library drives

	if value.Arch != wants.Arch {
		t.Errorf("Drive.Arch error [%d]: found %#v, wants %#v", i, value.Arch, wants.Arch)
	}
	if value.ImageType != wants.ImageType {
		t.Errorf("Drive.ImageType error [%d]: found %#v, wants %#v", i, value.ImageType, wants.ImageType)
	}
	if value.OS != wants.OS {
		t.Errorf("Drive.OS error [%d]: found %#v, wants %#v", i, value.OS, wants.OS)
	}
	if value.Paid != wants.Paid {
		t.Errorf("Drive.Paid error [%d]: found %#v, wants %#v", i, value.Paid, wants.Paid)
	}
}
