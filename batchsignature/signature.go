package batchsignature

import (
	"crypto/sha512"
)

// TODO: Move this to a protobuf message.
// TODO: This could probably use a better name; the residue doesn't include the leaf digest.
// TODO: Add a struct that contains one of these and a signature over the root digest.
type BatchResidue struct {
	pathDirections []bool // If true, the pathDigest is the right child; otherwise, it is the left child.
	pathDigests    [][]byte

	// leafDigest is the value that can be verified using this batch signature.  (TODO: Does it make sense to remove
	// this from this struct?)
	leafDigest []byte
}

func (t *BatchResidue) RootDigest() []byte {
	d := t.leafDigest

	for i := 0; i < len(t.pathDirections); i++ {
		if t.pathDirections[i] {
			d = nodeDigest(d, t.pathDigests[i])
		} else {
			d = nodeDigest(t.pathDigests[i], d)
		}
	}

	return d
}

func (t *BatchResidue) addPathSegment(direction bool, digest []byte) {
	t.pathDirections = append(t.pathDirections, direction)
	t.pathDigests = append(t.pathDigests, digest)
}

// returns (root-digest, digest-trees, error)
func NewDigestTree(leaves [][]byte) ([]byte, []BatchResidue, error) {
	trees := make([]BatchResidue, len(leaves))
	for i := 0; i < len(leaves); i++ {
		trees[i].leafDigest = leaves[i]
	}

	leafRootDigests := make([][]byte, len(leaves))
	for i := 0; i < len(leaves); i++ {
		leafRootDigests[i] = leaves[i]
	}

	for i := uint(0); (1 << i) < len(leaves); i++ {
		subtreeSize := 1 << (i + 1)
		halfSubtreeSize := 1 << i
		// log.Printf("i=%v (subtrees of size %v)\n", i, subtreeSize)
		for j := 0; len(leaves) > j+halfSubtreeSize; j += subtreeSize {
			// log.Printf("  i=%v j=%v\n", i, j)
			leftDigest := leafRootDigests[j]
			rightDigest := leafRootDigests[j+halfSubtreeSize]
			for k := j; k < j+halfSubtreeSize; k++ {
				trees[k].addPathSegment(true, rightDigest)
			}
			for k := j + halfSubtreeSize; (k < j+subtreeSize) && k < len(trees); k++ {
				trees[k].addPathSegment(false, leftDigest)
			}

			d := nodeDigest(leftDigest, rightDigest)
			leafRootDigests[j] = d
			leafRootDigests[j+halfSubtreeSize] = d
		}
	}

	var rootDigest []byte
	if len(leafRootDigests) > 0 {
		rootDigest = leafRootDigests[0]
	}
	return rootDigest, trees, nil
}

func nodeDigest(a, b []byte) []byte {
	h := sha512.New384()
	_, _ = h.Write(a)
	_, _ = h.Write(b)
	return h.Sum(nil)
}
