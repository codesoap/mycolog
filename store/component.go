package store

import "time"

// ComponentType signifies the type of a component and determines which
// relationships are legal for a component.
type ComponentType string

const (
	TypeSpores   ComponentType = "SPORES"
	TypeMycelium ComponentType = "MYC"
	TypeSpawn    ComponentType = "SPAWN"
	TypeGrow     ComponentType = "GROW"
)

// A Component is one node in a family tree. Type will signify whether
// it's spores, mycelium, spawn or a grow.
type Component struct {
	ID      int64
	Type    ComponentType
	Species string
	Token   string // A short token helping humans identify the component.

	// Time at which a sporeprint was taken, mycelium or spawn was
	// inoculated or a grow was started. Will only be stored as a date, the
	// exact time of day will be ignored.
	CreatedAt time.Time

	Notes string
	Gone  bool // True, if the component does not exist physically anymore.
}

// GrowInfo is additional information for a component of TypeGrow.
type GrowInfo struct {
	ID           int64
	Yield        *int // Yield in milligrams.
	YieldComment string
}
