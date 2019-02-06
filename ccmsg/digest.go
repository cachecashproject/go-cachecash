package ccmsg

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"hash"
)

// XXX: TODO: All of this needs to be improved so that unset fields do not cause nil pointer dereferences.

// XXX: Should we enforce validation of this message, and have CanonicalDigest return an error if the message is not
// valid (instead of silently return a bad digest))?
func (m *TicketBundle) CanonicalDigest() []byte {
	// There are four sub-hashes covering...
	// - the ticket-requests
	// - the ticket-L1s
	// - the encrypted/masked ticket-L2
	// - everything else
	//
	// The message's overall canonical digest is the digest of these four digests.  This allows for Merkle-tree-esquie
	// verification of the top-level digest without sharing the contents of the entire message.  (it's not clear to me
	// that there are real problems with just doing that, though.)

	return m.GetSubdigests().CanonicalDigest()
}

func (m *TicketBundle) GetSubdigests() *TicketBundleSubdigests {
	r := &TicketBundleSubdigests{
		EncryptedTicketL2Digest: m.canonicalEncryptedTicketL2Digest(),
		RemainderDigest:         m.Remainder.CanonicalDigest(),
	}

	for _, subMsg := range m.TicketRequest {
		r.TicketRequestDigest = append(r.TicketRequestDigest, subMsg.CanonicalDigest())
	}

	for _, subMsg := range m.TicketL1 {
		r.TicketL1Digest = append(r.TicketL1Digest, subMsg.CanonicalDigest())
	}

	return r
}

func (m *TicketBundle) canonicalEncryptedTicketL2Digest() []byte {
	h := sha512.New384()
	_, _ = h.Write(m.EncryptedTicketL2)
	return h.Sum(nil)
}

func (m *TicketBundleSubdigests) CanonicalDigest() []byte {
	h := sha512.New384()
	_, _ = h.Write(m.canonicalTicketRequestDigest())
	_, _ = h.Write(m.canonicalTicketL1Digest())
	_, _ = h.Write(m.EncryptedTicketL2Digest)
	_, _ = h.Write(m.RemainderDigest)
	return h.Sum(nil)
}

func (m *TicketBundleSubdigests) canonicalTicketRequestDigest() []byte {
	h := sha512.New384()
	for _, d := range m.TicketRequestDigest {
		_, _ = h.Write(d)
	}
	return h.Sum(nil)
}

func (m *TicketBundleSubdigests) canonicalTicketL1Digest() []byte {
	h := sha512.New384()
	for _, d := range m.TicketL1Digest {
		_, _ = h.Write(d)
	}
	return h.Sum(nil)
}

func (m *TicketBundleSubdigests) ContainsTicketRequestDigest(d []byte) bool {
	for _, x := range m.TicketRequestDigest {
		if bytes.Equal(x, d) {
			return true
		}
	}
	return false
}

func (m *TicketBundleSubdigests) ContainsTicketL1Digest(d []byte) bool {
	for _, x := range m.TicketL1Digest {
		if bytes.Equal(x, d) {
			return true
		}
	}
	return false
}

// XXX: Update this once the message contents are more stable!
func (m *TicketBundleRemainder) CanonicalDigest() []byte {
	h := sha512.New384()
	// _, _ = h.Write(m.ProviderPublicKey.PublicKey)
	// _, _ = h.Write(m.EscrowPublicKey.PublicKey)
	m.PuzzleInfo.canonicalDigest(h)
	return h.Sum(nil)
}

func (m *ColocationPuzzleInfo) canonicalDigest(h hash.Hash) {
	_, _ = h.Write(m.Goal)
	_ = binary.Write(h, binary.LittleEndian, m.Rounds)
	_ = binary.Write(h, binary.LittleEndian, m.StartOffset)
	_ = binary.Write(h, binary.LittleEndian, m.StartRange)
}

func (m *TicketRequest) CanonicalDigest() []byte {
	h := sha512.New384()
	_ = binary.Write(h, binary.LittleEndian, m.BlockIdx)
	_, _ = h.Write(m.InnerKey.Key)
	_, _ = h.Write(m.CachePublicKey.PublicKey)
	return h.Sum(nil)
}

func (m *TicketL1) CanonicalDigest() []byte {
	h := sha512.New384()
	_ = binary.Write(h, binary.LittleEndian, m.TicketNo)
	_, _ = h.Write(m.CachePublicKey.PublicKey)
	return h.Sum(nil)
}

func (m *TicketL2) CanonicalDigest() []byte {
	h := sha512.New384()
	_, _ = h.Write(m.Nonce) // Per docs, never returns an error.
	return h.Sum(nil)
}

func (m *TicketL2Info) EncryptedTicketL2Digest() []byte {
	h := sha512.New384()
	_, _ = h.Write(m.EncryptedTicketL2)
	return h.Sum(nil)
}
