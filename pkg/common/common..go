package common

type Orientation int

const (
	Portrait         Orientation = 0
	Landscape        Orientation = 2
	ReversePortrait  Orientation = 1
	ReverseLandscape Orientation = 3
)

type Revision string

const (
	RevisionA Revision = "A"
	RevisionB Revision = "B"
	RevisionC Revision = "C"
	RevisionD Revision = "D"
)
