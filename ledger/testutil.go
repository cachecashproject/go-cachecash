package ledger

import (
	"github.com/cachecashproject/go-cachecash/testutil"
)

func MustDecodeTXID(s string) TXID {
	d := testutil.MustDecodeString(s)
	var txid TXID
	if len(d) != len(txid) {
		panic("bad length for TXID")
	}
	copy(txid[:], d)
	return txid
}
