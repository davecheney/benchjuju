// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"regexp"
)

const ServiceTagKind = "service"

const (
	ServiceSnippet = "(?:[a-z][a-z0-9]*(?:-[a-z0-9]*[a-z][a-z0-9]*)*)"
	NumberSnippet  = "(?:0|[1-9][0-9]*)"
)

var validService = regexp.MustCompile("^" + ServiceSnippet + "$")

// IsValidService returns whether name is a valid service name.
func IsValidService(name string) bool {
	return validService.MatchString(name)
}

type ServiceTag struct {
	Name string
}

func (t ServiceTag) String() string { return t.Kind() + "-" + t.Id() }
func (t ServiceTag) Kind() string   { return ServiceTagKind }
func (t ServiceTag) Id() string     { return t.Name }

// NewServiceTag returns the tag for the service with the given name.
func NewServiceTag(serviceName string) ServiceTag {
	return ServiceTag{Name: serviceName}
}

// ParseServiceTag parses a service tag string.
func ParseServiceTag(serviceTag string) (ServiceTag, error) {
	tag, err := ParseTag(serviceTag)
	if err != nil {
		return ServiceTag{}, err
	}
	st, ok := tag.(ServiceTag)
	if !ok {
		return ServiceTag{}, invalidTagError(serviceTag, ServiceTagKind)
	}
	return st, nil
}
