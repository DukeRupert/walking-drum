package game

import (
	"testing"
)

func TestHiddenRoundTrip(t *testing.T) {
	orig := Hidden{}
	raw, err := EncodeComponent(orig)
	if err != nil {
		t.Fatalf("EncodeComponent: %v", err)
	}
	// Marker components serialize to "{}" — anything else means we've
	// accidentally given Hidden a field, which changes the contract.
	if string(raw) != "{}" {
		t.Errorf("encoded Hidden: got %q, want %q", raw, "{}")
	}

	var got Hidden
	if err := DecodeComponent(raw, &got); err != nil {
		t.Fatalf("DecodeComponent: %v", err)
	}
	if got != orig {
		t.Errorf("round-trip: got %+v, want %+v", got, orig)
	}
}

func TestHiddenComponentType(t *testing.T) {
	if got, want := (Hidden{}).ComponentType(), "hidden"; got != want {
		t.Errorf("ComponentType: got %q, want %q", got, want)
	}
}
