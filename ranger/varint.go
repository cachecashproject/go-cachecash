package ranger

// UvarintSize calculates the size for a binary.Uvarint
func UvarintSize(i uint64) int {
	var n int

	for {
		n++
		i >>= 7
		if i == 0 {
			break
		}
	}
	return n
}
