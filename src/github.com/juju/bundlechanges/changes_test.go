// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package bundlechanges_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"

	"github.com/juju/bundlechanges"
)

type changesSuite struct{}

var _ = gc.Suite(&changesSuite{})

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

// record holds expected information about the contents of a change value.
type record struct {
	Id       string
	Requires []string
	Method   string
	Params   interface{}
	GUIArgs  []interface{}
}

var fromDataTests = []struct {
	// about describes the test.
	about string
	// content is the YAML encoded bundle content.
	content string
	// expected holds the expected changes required to deploy the bundle.
	expected []record
}{{
	about: "minimal bundle",
	content: `
        services:
            django:
                charm: django
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "django",
		},
		GUIArgs: []interface{}{"django"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "django",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}},
}, {
	about: "simple bundle",
	content: `
        services:
            mediawiki:
                charm: cs:precise/mediawiki-10
                num_units: 1
                expose: true
                options:
                    debug: false
                annotations:
                    gui-x: "609"
                    gui-y: "-15"
            mysql:
                charm: cs:precise/mysql-28
                num_units: 1
        series: trusty
        relations:
            - - mediawiki:db
              - mysql:db
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:precise/mediawiki-10",
		},
		GUIArgs: []interface{}{"cs:precise/mediawiki-10"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "mediawiki",
			Options: map[string]interface{}{"debug": false},
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"mediawiki",
			map[string]interface{}{"debug": false},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "expose-2",
		Method: "expose",
		Params: bundlechanges.ExposeParams{
			Service: "$deploy-1",
		},
		GUIArgs:  []interface{}{"$deploy-1"},
		Requires: []string{"deploy-1"},
	}, {
		Id:     "setAnnotations-3",
		Method: "setAnnotations",
		Params: bundlechanges.SetAnnotationsParams{
			Id:          "$deploy-1",
			EntityType:  bundlechanges.ServiceType,
			Annotations: map[string]string{"gui-x": "609", "gui-y": "-15"},
		},
		GUIArgs: []interface{}{
			"$deploy-1",
			"service",
			map[string]string{"gui-x": "609", "gui-y": "-15"},
		},
		Requires: []string{"deploy-1"},
	}, {
		Id:     "addCharm-4",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:precise/mysql-28",
		},
		GUIArgs: []interface{}{"cs:precise/mysql-28"},
	}, {
		Id:     "deploy-5",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-4",
			Service: "mysql",
		},
		GUIArgs: []interface{}{
			"$addCharm-4",
			"mysql",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-4"},
	}, {
		Id:     "addRelation-6",
		Method: "addRelation",
		Params: bundlechanges.AddRelationParams{
			Endpoint1: "$deploy-1:db",
			Endpoint2: "$deploy-5:db",
		},
		GUIArgs:  []interface{}{"$deploy-1:db", "$deploy-5:db"},
		Requires: []string{"deploy-1", "deploy-5"},
	}, {
		Id:     "addUnit-7",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
		},
		GUIArgs:  []interface{}{"$deploy-1", nil},
		Requires: []string{"deploy-1"},
	}, {
		Id:     "addUnit-8",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-5",
		},
		GUIArgs:  []interface{}{"$deploy-5", nil},
		Requires: []string{"deploy-5"},
	}},
}, {
	about: "same charm reused",
	content: `
        services:
            mediawiki:
                charm: precise/mediawiki-10
                num_units: 1
            otherwiki:
                charm: precise/mediawiki-10
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "precise/mediawiki-10",
		},
		GUIArgs: []interface{}{"precise/mediawiki-10"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "mediawiki",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"mediawiki",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "deploy-2",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "otherwiki",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"otherwiki",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addUnit-3",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
		},
		GUIArgs:  []interface{}{"$deploy-1", nil},
		Requires: []string{"deploy-1"},
	}},
}, {
	about: "machines and units placement",
	content: `
        services:
            django:
                charm: cs:trusty/django-42
                num_units: 2
                to:
                    - 1
                    - lxc:2
                constraints: cpu-cores=4 cpu-power=42
            haproxy:
                charm: cs:trusty/haproxy-47
                num_units: 2
                expose: yes
                to:
                    - lxc:django/0
                    - new
                options:
                    bad: wolf
                    number: 42.47
        machines:
            1:
                series: trusty
            2:
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/django-42",
		},
		GUIArgs: []interface{}{"cs:trusty/django-42"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:       "$addCharm-0",
			Service:     "django",
			Constraints: "cpu-cores=4 cpu-power=42",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"cpu-cores=4 cpu-power=42",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addCharm-2",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/haproxy-47",
		},
		GUIArgs: []interface{}{"cs:trusty/haproxy-47"},
	}, {
		Id:     "deploy-3",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-2",
			Service: "haproxy",
			Options: map[string]interface{}{"bad": "wolf", "number": 42.47},
		},
		GUIArgs: []interface{}{
			"$addCharm-2",
			"haproxy",
			map[string]interface{}{"bad": "wolf", "number": 42.47},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-2"},
	}, {
		Id:     "expose-4",
		Method: "expose",
		Params: bundlechanges.ExposeParams{
			Service: "$deploy-3",
		},
		GUIArgs:  []interface{}{"$deploy-3"},
		Requires: []string{"deploy-3"},
	}, {
		Id:     "addMachines-5",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			Series: "trusty",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{Series: "trusty"},
		},
	}, {
		Id:     "addMachines-6",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addUnit-7",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-5",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-5"},
		Requires: []string{"deploy-1", "addMachines-5"},
	}, {
		Id:     "addMachines-11",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "lxc",
			ParentId:      "$addMachines-6",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "lxc",
				ParentId:      "$addMachines-6",
			},
		},
		Requires: []string{"addMachines-6"},
	}, {
		Id:     "addMachines-12",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "lxc",
			ParentId:      "$addUnit-7",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "lxc",
				ParentId:      "$addUnit-7",
			},
		},
		Requires: []string{"addUnit-7"},
	}, {
		Id:     "addMachines-13",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addUnit-8",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-11",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-11"},
		Requires: []string{"deploy-1", "addMachines-11"},
	}, {
		Id:     "addUnit-9",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
			To:      "$addMachines-12",
		},
		GUIArgs:  []interface{}{"$deploy-3", "$addMachines-12"},
		Requires: []string{"deploy-3", "addMachines-12"},
	}, {
		Id:     "addUnit-10",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
			To:      "$addMachines-13",
		},
		GUIArgs:  []interface{}{"$deploy-3", "$addMachines-13"},
		Requires: []string{"deploy-3", "addMachines-13"},
	}},
}, {
	about: "machines with constraints and annotations",
	content: `
        services:
            django:
                charm: cs:trusty/django-42
                num_units: 2
                to:
                    - 1
                    - new
        machines:
            1:
                constraints: "cpu-cores=4"
                annotations:
                    foo: bar
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/django-42",
		},
		GUIArgs: []interface{}{"cs:trusty/django-42"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "django",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addMachines-2",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			Constraints: "cpu-cores=4",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{Constraints: "cpu-cores=4"},
		},
	}, {
		Id:     "setAnnotations-3",
		Method: "setAnnotations",
		Params: bundlechanges.SetAnnotationsParams{
			Id:          "$addMachines-2",
			EntityType:  bundlechanges.MachineType,
			Annotations: map[string]string{"foo": "bar"},
		},
		GUIArgs: []interface{}{
			"$addMachines-2",
			"machine",
			map[string]string{"foo": "bar"},
		},
		Requires: []string{"addMachines-2"},
	}, {
		Id:     "addUnit-4",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-2",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-2"},
		Requires: []string{"deploy-1", "addMachines-2"},
	}, {
		Id:     "addMachines-6",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addUnit-5",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-6",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-6"},
		Requires: []string{"deploy-1", "addMachines-6"},
	}},
}, {
	about: "endpoint without relation name",
	content: `
        services:
            mediawiki:
                charm: cs:precise/mediawiki-10
            mysql:
                charm: cs:precise/mysql-28
                constraints: mem=42G
        relations:
            - - mediawiki:db
              - mysql
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:precise/mediawiki-10",
		},
		GUIArgs: []interface{}{"cs:precise/mediawiki-10"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "mediawiki",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"mediawiki",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addCharm-2",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:precise/mysql-28",
		},
		GUIArgs: []interface{}{"cs:precise/mysql-28"},
	}, {
		Id:     "deploy-3",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:       "$addCharm-2",
			Service:     "mysql",
			Constraints: "mem=42G",
		},
		GUIArgs: []interface{}{
			"$addCharm-2",
			"mysql",
			map[string]interface{}{},
			"mem=42G",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-2"},
	}, {
		Id:     "addRelation-4",
		Method: "addRelation",
		Params: bundlechanges.AddRelationParams{
			Endpoint1: "$deploy-1:db",
			Endpoint2: "$deploy-3",
		},
		GUIArgs:  []interface{}{"$deploy-1:db", "$deploy-3"},
		Requires: []string{"deploy-1", "deploy-3"},
	}},
}, {
	about: "unit placed in service",
	content: `
        services:
            wordpress:
                charm: wordpress
                num_units: 3
            django:
                charm: cs:trusty/django-42
                num_units: 2
                to: [wordpress]
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/django-42",
		},
		GUIArgs: []interface{}{"cs:trusty/django-42"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "django",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addCharm-2",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "wordpress",
		},
		GUIArgs: []interface{}{"wordpress"},
	}, {
		Id:     "deploy-3",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-2",
			Service: "wordpress",
		},
		GUIArgs: []interface{}{
			"$addCharm-2",
			"wordpress",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-2"},
	}, {
		Id:     "addUnit-6",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
		},
		GUIArgs:  []interface{}{"$deploy-3", nil},
		Requires: []string{"deploy-3"},
	}, {
		Id:     "addUnit-7",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
		},
		GUIArgs:  []interface{}{"$deploy-3", nil},
		Requires: []string{"deploy-3"},
	}, {
		Id:     "addUnit-8",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
		},
		GUIArgs:  []interface{}{"$deploy-3", nil},
		Requires: []string{"deploy-3"},
	}, {
		Id:     "addUnit-4",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addUnit-6",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addUnit-6"},
		Requires: []string{"deploy-1", "addUnit-6"},
	}, {
		Id:     "addUnit-5",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addUnit-7",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addUnit-7"},
		Requires: []string{"deploy-1", "addUnit-7"},
	}},
}, {
	about: "unit co-location with other units",
	content: `
        services:
            memcached:
                charm: cs:trusty/mem-47
                num_units: 3
                to: [1, new]
            django:
                charm: cs:trusty/django-42
                num_units: 5
                to:
                    - memcached/0
                    - lxc:memcached/1
                    - lxc:memcached/2
                    - kvm:ror
            ror:
                charm: vivid/rails
                num_units: 2
                to:
                    - new
                    - 1
        machines:
            1:
                series: trusty
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/django-42",
		},
		GUIArgs: []interface{}{"cs:trusty/django-42"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "django",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addCharm-2",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/mem-47",
		},
		GUIArgs: []interface{}{"cs:trusty/mem-47"},
	}, {
		Id:     "deploy-3",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-2",
			Service: "memcached",
		},
		GUIArgs: []interface{}{
			"$addCharm-2",
			"memcached",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-2"},
	}, {
		Id:     "addCharm-4",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "vivid/rails",
		},
		GUIArgs: []interface{}{"vivid/rails"},
	}, {
		Id:     "deploy-5",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-4",
			Service: "ror",
		},
		GUIArgs: []interface{}{
			"$addCharm-4",
			"ror",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-4"},
	}, {
		Id:     "addMachines-6",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			Series: "trusty",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				Series:      "trusty",
				Constraints: "",
			},
		},
	}, {
		Id:     "addUnit-12",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
			To:      "$addMachines-6",
		},
		GUIArgs:  []interface{}{"$deploy-3", "$addMachines-6"},
		Requires: []string{"deploy-3", "addMachines-6"},
	}, {
		Id:     "addUnit-16",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-5",
			To:      "$addMachines-6",
		},
		GUIArgs:  []interface{}{"$deploy-5", "$addMachines-6"},
		Requires: []string{"deploy-5", "addMachines-6"},
	}, {
		Id:     "addMachines-20",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "kvm",
			ParentId:      "$addUnit-16",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "kvm",
				ParentId:      "$addUnit-16",
			},
		},
		Requires: []string{"addUnit-16"},
	}, {
		Id:     "addMachines-21",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addMachines-22",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addMachines-23",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addUnit-7",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addUnit-12",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addUnit-12"},
		Requires: []string{"deploy-1", "addUnit-12"},
	}, {
		Id:     "addUnit-11",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-20",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-20"},
		Requires: []string{"deploy-1", "addMachines-20"},
	}, {
		Id:     "addUnit-13",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
			To:      "$addMachines-21",
		},
		GUIArgs:  []interface{}{"$deploy-3", "$addMachines-21"},
		Requires: []string{"deploy-3", "addMachines-21"},
	}, {
		Id:     "addUnit-14",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-3",
			To:      "$addMachines-22",
		},
		GUIArgs:  []interface{}{"$deploy-3", "$addMachines-22"},
		Requires: []string{"deploy-3", "addMachines-22"},
	}, {
		Id:     "addUnit-15",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-5",
			To:      "$addMachines-23",
		},
		GUIArgs:  []interface{}{"$deploy-5", "$addMachines-23"},
		Requires: []string{"deploy-5", "addMachines-23"},
	}, {
		Id:     "addMachines-17",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "lxc",
			ParentId:      "$addUnit-13",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "lxc",
				ParentId:      "$addUnit-13",
			},
		},
		Requires: []string{"addUnit-13"},
	}, {
		Id:     "addMachines-18",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "lxc",
			ParentId:      "$addUnit-14",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "lxc",
				ParentId:      "$addUnit-14",
			},
		},
		Requires: []string{"addUnit-14"},
	}, {
		Id:     "addMachines-19",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "kvm",
			ParentId:      "$addUnit-15",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "kvm",
				ParentId:      "$addUnit-15",
			},
		},
		Requires: []string{"addUnit-15"},
	}, {
		Id:     "addUnit-8",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-17",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-17"},
		Requires: []string{"deploy-1", "addMachines-17"},
	}, {
		Id:     "addUnit-9",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-18",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-18"},
		Requires: []string{"deploy-1", "addMachines-18"},
	}, {
		Id:     "addUnit-10",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-19",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-19"},
		Requires: []string{"deploy-1", "addMachines-19"},
	}},
}, {
	about: "unit placed to machines",
	content: `
        services:
            django:
                charm: cs:trusty/django-42
                num_units: 5
                to:
                    - new
                    - 4
                    - kvm:8
                    - lxc:new
        machines:
            4:
                constraints: "cpu-cores=4"
            8:
                constraints: "cpu-cores=8"
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/django-42",
		},
		GUIArgs: []interface{}{"cs:trusty/django-42"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "django",
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addMachines-2",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			Constraints: "cpu-cores=4",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				Constraints: "cpu-cores=4",
			},
		},
	}, {
		Id:     "addMachines-3",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			Constraints: "cpu-cores=8",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				Constraints: "cpu-cores=8",
			},
		},
	}, {
		Id:     "addUnit-5",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-2",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-2"},
		Requires: []string{"deploy-1", "addMachines-2"},
	}, {
		Id:     "addMachines-9",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{},
		},
	}, {
		Id:     "addMachines-10",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "kvm",
			ParentId:      "$addMachines-3",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "kvm",
				ParentId:      "$addMachines-3",
			},
		},
		Requires: []string{"addMachines-3"},
	}, {
		Id:     "addMachines-11",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "lxc",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "lxc",
			},
		},
	}, {
		Id:     "addMachines-12",
		Method: "addMachines",
		Params: bundlechanges.AddMachineParams{
			ContainerType: "lxc",
		},
		GUIArgs: []interface{}{
			bundlechanges.AddMachineOptions{
				ContainerType: "lxc",
			},
		},
	}, {
		Id:     "addUnit-4",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-9",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-9"},
		Requires: []string{"deploy-1", "addMachines-9"},
	}, {
		Id:     "addUnit-6",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-10",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-10"},
		Requires: []string{"deploy-1", "addMachines-10"},
	}, {
		Id:     "addUnit-7",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-11",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-11"},
		Requires: []string{"deploy-1", "addMachines-11"},
	}, {
		Id:     "addUnit-8",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
			To:      "$addMachines-12",
		},
		GUIArgs:  []interface{}{"$deploy-1", "$addMachines-12"},
		Requires: []string{"deploy-1", "addMachines-12"},
	}},
}, {
	about: "service with storage",
	content: `
        services:
            django:
                charm: cs:trusty/django-42
                num_units: 2
                storage:
                    osd-devices: 3,30G
                    tmpfs: tmpfs,1G
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "cs:trusty/django-42",
		},
		GUIArgs: []interface{}{"cs:trusty/django-42"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:   "$addCharm-0",
			Service: "django",
			Storage: map[string]string{
				"osd-devices": "3,30G",
				"tmpfs":       "tmpfs,1G",
			},
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{
				"osd-devices": "3,30G",
				"tmpfs":       "tmpfs,1G",
			},
			map[string]string{},
		},
		Requires: []string{"addCharm-0"},
	}, {
		Id:     "addUnit-2",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
		},
		GUIArgs:  []interface{}{"$deploy-1", nil},
		Requires: []string{"deploy-1"},
	}, {
		Id:     "addUnit-3",
		Method: "addUnit",
		Params: bundlechanges.AddUnitParams{
			Service: "$deploy-1",
		},
		GUIArgs:  []interface{}{"$deploy-1", nil},
		Requires: []string{"deploy-1"},
	}},
}, {
	about: "service with endpoint bindings",
	content: `
        services:
            django:
                charm: django
                bindings:
                    foo: bar
    `,
	expected: []record{{
		Id:     "addCharm-0",
		Method: "addCharm",
		Params: bundlechanges.AddCharmParams{
			Charm: "django",
		},
		GUIArgs: []interface{}{"django"},
	}, {
		Id:     "deploy-1",
		Method: "deploy",
		Params: bundlechanges.AddServiceParams{
			Charm:            "$addCharm-0",
			Service:          "django",
			EndpointBindings: map[string]string{"foo": "bar"},
		},
		GUIArgs: []interface{}{
			"$addCharm-0",
			"django",
			map[string]interface{}{},
			"",
			map[string]string{},
			map[string]string{"foo": "bar"},
		},
		Requires: []string{"addCharm-0"},
	}},
}}

func (s *changesSuite) TestFromData(c *gc.C) {
	for i, test := range fromDataTests {
		c.Logf("test %d: %s", i, test.about)

		// Retrieve and validate the bundle data.
		data, err := charm.ReadBundleData(strings.NewReader(test.content))
		c.Assert(err, jc.ErrorIsNil)
		err = data.Verify(nil, nil)
		c.Assert(err, jc.ErrorIsNil)

		// Retrieve the changes, and convert them to a sequence of records.
		changes := bundlechanges.FromData(data)
		records := make([]record, len(changes))
		for i, change := range changes {
			r := record{
				Id:       change.Id(),
				Requires: change.Requires(),
				Method:   change.Method(),
				GUIArgs:  change.GUIArgs(),
			}
			r.Params = reflect.ValueOf(change).Elem().FieldByName("Params").Interface()
			records[i] = r
		}

		// Output the records for debugging.
		b, err := json.MarshalIndent(records, "", "  ")
		c.Assert(err, jc.ErrorIsNil)
		c.Logf("obtained records: %s", b)

		// Check that the obtained records are what we expect.
		c.Assert(records, jc.DeepEquals, test.expected)
	}
}
