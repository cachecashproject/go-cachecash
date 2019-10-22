package ledger

import (
	"github.com/pkg/errors"
)

// TODO:
// - Need mechanism for tracking not-yet-activated transactions and applying them at the right height.
// - Need mechanism for verifying signature (including that SigPublicKey is in `GlobalConfigKeys`).

type GlobalConfigState struct {
	Scalars map[string][]byte
	Lists   map[string][][]byte
}

func NewGlobalConfigState() *GlobalConfigState {
	return &GlobalConfigState{
		Scalars: make(map[string][]byte),
		Lists:   make(map[string][][]byte),
	}
}

func (st *GlobalConfigState) copy() *GlobalConfigState {
	st2 := &GlobalConfigState{
		Scalars: make(map[string][]byte),
		Lists:   make(map[string][][]byte),
	}

	for k, v := range st.Scalars {
		st2.Scalars[k] = v
	}
	for k, v := range st.Lists {
		lst := make([][]byte, len(v))
		copy(lst, v)
		st2.Lists[k] = lst
	}

	return st2
}

// Apply returns a duplicate of st that has had the changes described in tx applied.
func (st *GlobalConfigState) Apply(tx *GlobalConfigTransaction) (*GlobalConfigState, error) {
	nextSt := st.copy()

	for _, su := range tx.ScalarUpdates {
		if len(su.Value) == 0 {
			delete(nextSt.Scalars, su.Key)
		} else {
			nextSt.Scalars[su.Key] = su.Value
		}
	}

	for _, lu := range tx.ListUpdates {
		// If no list exists, v will be nil.
		v := nextSt.Lists[lu.Key]

		v, err := st.applyListUpdate(v, lu)
		if err != nil {
			return nil, errors.Wrap(err, "failed to apply list update")
		}

		nextSt.Lists[lu.Key] = v
	}

	return nextSt, nil
}

func (st *GlobalConfigState) applyListUpdate(v [][]byte, lu GlobalConfigListUpdate) ([][]byte, error) {
	if len(v) < len(lu.Deletions) {
		return nil, errors.New("more deletions in update than elements in list")
	}
	if err := lu.validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate list update")
	}

	// No-ops are not valid.
	if len(lu.Deletions) == 0 && len(lu.Insertions) == 0 {
		return nil, errors.New("no operations provided for key")
	}

	// Apply deletions.
	v2 := make([][]byte, len(v)-len(lu.Deletions))
	var j, k int
	for i := 0; i < len(v); i++ {
		if j < len(lu.Deletions) && lu.Deletions[j] == uint64(i) {
			// This value has been deleted; skip it.
			j++
		} else {
			v2[k] = v[i]
			k++
		}
	}

	// Apply insertions.
	for _, li := range lu.Insertions {
		idx := int(li.Index) // XXX: Wraparound?
		if idx > len(v2) {
			return nil, errors.New("list insertion index out of range")
		}

		v2 = append(v2[:idx], append([][]byte{li.Value}, v2[idx:]...)...)
	}

	return v2, nil
}

// TODO: Helpers for retrieving GCP scalars/lists of various types (uint64, int64, []byte, string).

func (st *GlobalConfigState) GetUint64(key string) (uint64, error) {
	// v, ok := st.Scalars[key]
	// if !ok {
	// 	return uint64(0), errors.New("no value")
	// }
	return uint64(0), errors.New("no impl")
}
