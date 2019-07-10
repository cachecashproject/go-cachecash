package ledger

// TODO: Should this go elsewhere?
//
// Functions for actual encoding/decoding these values are in `encoding/binary` in the standard library.

func UvarintSize(i uint) int {
	switch {
	case i < 2<<7:
		return 1
	case i < 2<<14:
		return 2
	case i < 2<<21:
		return 3
	case i < 2<<28:
		return 4
	case i < 2<<35:
		return 5
	case i < 2<<42:
		return 6
	case i < 2<<49:
		return 7
	case i < 2<<56:
		return 8
	default:
		return 9
	}
}
