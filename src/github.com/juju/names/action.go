// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"

	"github.com/juju/utils"
)

const ActionTagKind = "action"

type ActionTag struct {
	// Tags that are serialized need to have fields exported.
	ID utils.UUID
}

// NewActionTag returns the tag of an action with the given id (UUID).
func NewActionTag(id string) ActionTag {
	uuid, err := utils.UUIDFromString(id)
	if err != nil {
		panic(err)
	}
	return ActionTag{ID: uuid}
}

// ParseActionTag parses an action tag string.
func ParseActionTag(actionTag string) (ActionTag, error) {
	tag, err := ParseTag(actionTag)
	if err != nil {
		return ActionTag{}, err
	}
	at, ok := tag.(ActionTag)
	if !ok {
		return ActionTag{}, invalidTagError(actionTag, ActionTagKind)
	}
	return at, nil
}

func (t ActionTag) String() string { return t.Kind() + "-" + t.Id() }
func (t ActionTag) Kind() string   { return ActionTagKind }
func (t ActionTag) Id() string     { return t.ID.String() }

// IsValidAction returns whether id is a valid action id (UUID).
func IsValidAction(id string) bool {
	return utils.IsValidUUIDString(id)
}

func ActionReceiverTag(name string) (Tag, error) {
	if IsValidUnit(name) {
		return NewUnitTag(name), nil
	}
	if IsValidService(name) {
		// TODO(jcw4) enable when leader elections complete
		//return NewServiceTag(name), nil
	}
	return nil, fmt.Errorf("invalid actionreceiver name %q", name)
}
