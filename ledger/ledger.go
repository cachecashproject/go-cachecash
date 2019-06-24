package ledger

/*
In order to be used as a gogo/protobuf custom type, a struct must implement the folowing methods.
func (t T) Marshal() ([]byte, error) {}
func (t *T) MarshalTo(data []byte) (n int, err error) {}
func (t *T) Unmarshal(data []byte) error {}
func (t *T) Size() int {}
func (t T) MarshalJSON() ([]byte, error) {}
func (t *T) UnmarshalJSON(data []byte) error {}

*/

type Transaction struct {
}

type Block struct {
}
