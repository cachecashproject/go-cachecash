package ranger

// Marshaler is an interface to define how marshaling happens in Ranger. All
// specified types must be either a supported, native type such as uint or
// string, or implement this interface by either doing it by hand or through
// rangergen.
type Marshaler interface {
	Marshal() ([]byte, error)
	MarshalTo(data []byte) (n int, err error)
	Unmarshal(data []byte) error
	UnmarshalFrom(data []byte) (int, error)
	Size() int
}
