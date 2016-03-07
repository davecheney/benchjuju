// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"
	"regexp"
)

const NetworkTagKind = "network"

const (
	NetworkSnippet = "(?:[a-zA-Z0-9_-]+)"
)

var validNetwork = regexp.MustCompile("^" + NetworkSnippet + "$")

// IsValidNetwork reports whether name is a valid network name.
func IsValidNetwork(name string) bool {
	return validNetwork.MatchString(name)
}

type NetworkTag struct {
	name string
}

func (t NetworkTag) String() string { return t.Kind() + "-" + t.Id() }
func (t NetworkTag) Kind() string   { return NetworkTagKind }
func (t NetworkTag) Id() string     { return t.name }

// NewNetworkTag returns the tag of a network with the given name.
func NewNetworkTag(name string) NetworkTag {
	if !IsValidNetwork(name) {
		panic(fmt.Sprintf("%q is not a valid network name", name))
	}
	return NetworkTag{name: name}
}

// ParseNetworkTag parses a network tag string.
func ParseNetworkTag(networkTag string) (NetworkTag, error) {
	tag, err := ParseTag(networkTag)
	if err != nil {
		return NetworkTag{}, err
	}
	nt, ok := tag.(NetworkTag)
	if !ok {
		return NetworkTag{}, invalidTagError(networkTag, NetworkTagKind)
	}
	return nt, nil
}
