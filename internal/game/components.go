package game

import (
	"encoding/json"
	"fmt"
)

// Component-type strings live as constants so they're greppable and
// renames are mechanical. The DB column is free-form TEXT (DESIGN.md
// §6.4); validation is purely Go-side.
const (
	ComponentHidden = "hidden"
)

// Hidden is a marker component: its presence on an entity means the
// entity isn't broadcast to clients (think stealth, fog-of-war, GM
// invisibility). The struct is empty by design — the *presence of the
// row* is the signal.
type Hidden struct{}

// ComponentType lets Hidden satisfy Component.
func (Hidden) ComponentType() string { return ComponentHidden }

// EncodeComponent serializes c to the JSONB blob that lands in
// components.state. Centralized so swapping encodings later is a
// one-place change.
func EncodeComponent(c Component) ([]byte, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("encode component %s: %w", c.ComponentType(), err)
	}
	return b, nil
}

// DecodeComponent unmarshals raw JSONB state into the supplied component
// pointer. The pointer's type doubles as the schema — callers know what
// they're reading because they pass it in.
func DecodeComponent(raw []byte, into Component) error {
	if err := json.Unmarshal(raw, into); err != nil {
		return fmt.Errorf("decode component %s: %w", into.ComponentType(), err)
	}
	return nil
}
