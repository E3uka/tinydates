package store

// PotentialMatch is a transient object of the user obtained from the call to
// Discover. This is the cost required to abstract the store methods over a
// concrete implementation.
type PotentialMatch struct {
	Id     int
	Name   string
	Gender string
	Age    int
}
