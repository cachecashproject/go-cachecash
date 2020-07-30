// +build gofuzz

package ledger

// Fuzz function to be used by https://github.com/dvyukov/go-fuzz
func Fuzz(data []byte) int {
	block := Block{}
	block.Unmarshal(data)
	return 0
}
