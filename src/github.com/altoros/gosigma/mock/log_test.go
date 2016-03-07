// Copyright 2014 ALTOROS
// Licensed under the AGPLv3, see LICENSE file for details.

package mock

import "testing"

func TestMockLogSeverity(t *testing.T) {

	t.Parallel()

	if v := parseLogSeverity(nil); v != logNone {
		t.Errorf("Severity of nil == %d, wants logNone", v)
	}

	check := func(s string, sv severity) {
		if v := parseLogSeverity(&s); v != sv {
			t.Errorf("Severity of %v == %d, wants %d", s, v, sv)
		}
	}

	check("", logNone)
	check("n", logNone)
	check("none", logNone)

	check("u", logURL)
	check("url", logURL)

	check("d", logDetail)
	check("detail", logDetail)
}
