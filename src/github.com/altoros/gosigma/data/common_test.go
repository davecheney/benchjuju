// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package data

import "testing"

func verifyMeta(t *testing.T, m *Meta, limit, offset, count int) {
	if m.Limit != limit {
		t.Errorf("Meta.Limit = %d, wants %d", m.Limit, limit)
	}
	if m.Offset != offset {
		t.Errorf("Meta.Offset = %d, wants %d", m.Offset, offset)
	}
	if m.TotalCount != count {
		t.Errorf("Meta.TotalCount = %d, wants %d", m.TotalCount, count)
	}
}

func verifyResource(t *testing.T, r *Resource, uri, uuid string) {
	if r.URI != uri {
		t.Errorf("Resource.URI = %s, wants %s", r.URI, uri)
	}
	if r.UUID != uuid {
		t.Errorf("Resource.UUID = %s, wants %s", r.UUID, uuid)
	}
}

func compareMeta(t *testing.T, prefix string, value, wants map[string]string) {
	if len(value) != len(wants) {
		t.Errorf(prefix+": found %#v, wants %#v", value, wants)
	}
	for k, w := range wants {
		if v, ok := value[k]; ok {
			if v != w {
				t.Errorf(prefix+": [%q] found %#v, wants %#v", k, v, w)
			}
		} else {
			t.Errorf(": [%q] not found in values", k)
		}
	}
}
