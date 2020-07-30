package ledger

import (
	"github.com/cachecashproject/go-cachecash/testutil"

	"github.com/cachecashproject/go-cachecash/ledger/models"
)

// MustDecodeBlocKID provides convenient glue for making block ID instances in code.
func MustDecodeBlockID(s string) BlockID {
	d := testutil.MustDecodeString(s)
	var blockid BlockID
	n := copy(blockid[:], d)
	if n != len(blockid) {
		panic("bad length for BlockID")
	}
	return blockid
}

func MustDecodeTXID(s string) models.TXID {
	d := testutil.MustDecodeString(s)
	var txid models.TXID
	if len(d) != len(txid) {
		panic("bad length for TXID")
	}
	copy(txid[:], d)
	return txid
}
