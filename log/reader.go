package log

import (
	"encoding/binary"
	"errors"
	"os"
)

// Reader is a reader for the structured log bundles. A "log bundle" is a
// stream of bytes consisting of a length prefix in int32 form, a padding
// int32, and the protobuf marshaled out. This pattern repeats itself until
// EOF. The marshaled proto is expected to unmarshal into the log.Entry type
// (see message.proto).
//
// Reading the bundle is an act of instantiating this struct with NewReader()
// and calling NextProto() on the result until io.EOF is returned.
type Reader struct {
	file *os.File
}

// NewReader creates a reader.
func NewReader(filename string) (*Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Reader{file: f}, nil
}

// Close the file
func (r *Reader) Close() error {
	return r.file.Close()
}

// NextProto finds the next protobuf in the bundle, if nil is returned, error
// will contain state which may be io.EOF -- indicating the file is finished.
func (r *Reader) NextProto() (*Entry, error) {
	var length int64

	if err := binary.Read(r.file, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length < 1 {
		return nil, errors.New("invalid length in protobuf read for logger")
	}

	buf := make([]byte, length)
	if _, err := r.file.Read(buf); err != nil {
		return nil, err
	}

	e := &Entry{}
	return e, e.Unmarshal(buf)
}
