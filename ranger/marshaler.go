package ranger

// Marshaler is an interface to define how marshaling happens in Ranger. All
// specified types must be either a supported, native type such as uint or
// string, or implement this interface by either doing it by hand or through
// rangergen.
//
// This is deliberately compatible with the way that GRPC/gogoproto generated
// code wants to call things, so that generated types can be embedded as foreign
// types in protobufs.
type Marshaler interface {
	// Allocate storage
	Marshal() ([]byte, error)
	// Marshal into preallocated storage and returned consumed bytes.
	MarshalTo(data []byte) (int, error)
	// Unmarshal the first document in data
	Unmarshal(data []byte) error
	// Unmarshal the first document in data returning consumed bytes
	UnmarshalFrom(data []byte) (int, error)
	// Return the size that the struct will need to Marshal successfully. If the
	// struct cannot be marshalled successfully for some reason, will return a
	// sufficient number of bytes (perhaps even 0) to allow marshalTo to reason
	// the point where that marshalling error will occur, and will itself not
	// panic. MarshalTo will then proceed to attempt to marshal into this space
	// and the marshaling will fail. The failure should still be due to the
	// underlying reason, not due to a shortness of space returned from Size.
	Size() int
}
