package ledger

import (
	"context"
	"encoding/binary"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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

func gcpMarshalUInt64(val uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	// panics on error
	n := binary.PutUvarint(buf, val)
	return buf[:n]
}

func gcpUnmarshalUInt64(stored []byte) (uint64, error) {
	val, n := binary.Uvarint(stored)
	if n < 1 {
		return 0, errors.New("short read on uint64")
	}
	return val, nil
}

func gcpMarshalInt64(val int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	// panics on error
	n := binary.PutVarint(buf, val)
	return buf[:n]
}

func gcpUnmarshalInt64(stored []byte) (int64, error) {
	val, n := binary.Varint(stored)
	if n < 1 {
		return 0, errors.New("short read on int64")
	}
	return val, nil
}

// GlobalConfigStateYAML ---- start ----

// GlobalConfigStateYAML is derived from GlobalConfigState by converting the typed configuration keys into their known
// types. Unknown keys and byte strings are base64 encoded for output.
type GlobalConfigStateYAML struct {
	Scalars map[string]interface{}   `yaml:"scalars"`
	Lists   map[string][]interface{} `yaml:"lists"`
}

// NewGlobalConfigStateYAML creates new YAML-friendly version of the
// GlobalConfigState. It does not come pre-populated with data from any
// GlobalConfigState.
func NewGlobalConfigStateYAML() *GlobalConfigStateYAML {
	return &GlobalConfigStateYAML{
		Scalars: map[string]interface{}{},
		Lists:   map[string][]interface{}{},
	}
}

// ToYAML produces a GlobalConfigStateYAML, which contains known [u]int64 parameters as ints for display.
func (st *GlobalConfigState) ToYAML() (*GlobalConfigStateYAML, error) {
	result := NewGlobalConfigStateYAML()
	for name, val := range st.Scalars {
		paramType := parameterType(name)
		converted, err := paramType.yamlFromValue(val)
		if err != nil {
			return nil, errors.Wrapf(err, "while marshaling %q", name)
		}
		result.Scalars[name] = converted
	}

	for name, vals := range st.Lists {
		paramType := parameterType(name)
		for i, val := range vals {
			converted, err := paramType.yamlFromValue(val)
			if err != nil {
				return nil, errors.Wrapf(err, "while marshaling %q, index %d", name, i)
			}
			result.Lists[name] = append(result.Lists[name], converted)

		}
	}

	return result, nil
}

// GlobalConfigStateYAML ---- end ----

// globalConfigPatchParameterYAML ---- start ----
// globalConfigPatchParameterYAML provides the thunk to unmarshal a single parameter from a patch file
type globalConfigPatchParameterYAML struct {
	Name       string      `yaml:"name"`
	Value      interface{} `yaml:"value"`
	Insertions interface{} `yaml:"insertions"`
	Deletions  interface{} `yaml:"deletions"`
	scalar     *GlobalConfigScalarUpdate
	list       *GlobalConfigListUpdate
}

func (param *globalConfigPatchParameterYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// This might be a List or a Scalar parameter patch; first we unmarshal to identify the parameter, then we know what
	// to unmarshal as.
	type SP globalConfigPatchParameterYAML
	var field *SP = (*SP)(param)
	if err := unmarshal(&field); err != nil {
		return err
	}
	switch parameterCategory(param.Name) {
	case SchemaList:
		var list globalConfigPatchListParameter
		if err := list.UnmarshalYAML(unmarshal); err != nil {
			return err
		}
		param.list = list.list
	case SchemaScalar:
		var scalar globalConfigPatchScalarParameter
		if err := scalar.UnmarshalYAML(unmarshal); err != nil {
			return err
		}
		param.scalar = scalar.scalar
	default:
		return errors.Errorf("unknown parameter %q", param.Name)
	}
	return nil
}

// GlobalConfigPatchParameterYAML ---- end ----

// globalConfigPatchListParameter ---- start ----
// globalConfigPatchListParameter provides the thunk to unmarshal a list parameter patch from a patch file
type globalConfigPatchListParameter struct {
	Name       string                           `yaml:"name"`
	Deletions  []uint64                         `yaml:"deletions"`
	Insertions []globalConfigPatchListInsertion `yaml:"insertions"`
	list       *GlobalConfigListUpdate
}

type globalConfigPatchListInsertion struct {
	Index uint64
	Value interface{}
}

func (param *globalConfigPatchListParameter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type SP globalConfigPatchListParameter
	var field *SP = (*SP)(param)
	if err := unmarshal(&field); err != nil {
		return err
	}
	param.list = &GlobalConfigListUpdate{Key: param.Name, Deletions: param.Deletions, Insertions: make([]GlobalConfigListInsertion, len(param.Insertions))}
	paramType := parameterType(param.Name)
	switch parameterCategory(param.Name) {
	case SchemaScalar:
		return errors.Errorf("Bad category for parameter %q", param.Name)
	case SchemaList:
		for i, insertion := range param.Insertions {
			param.list.Insertions[i].Index = insertion.Index
			data, err := paramType.valueFromYAML(insertion.Value)
			if err != nil {
				return errors.Wrapf(err, "Missing / bad value for parameter %q", param.Name)
			}
			param.list.Insertions[i].Value = data
		}
	default:
		return errors.Errorf("Invalid parameter %q", param.Name)
	}
	return nil
}

// GlobalConfigPatchListParameter ---- end ----

// globalConfigPatchScalarParameter ---- start ----
// globalConfigPatchScalarParameter provides the thunk to unmarshal a list parameter patch from a patch file
type globalConfigPatchScalarParameter struct {
	Name   string      `yaml:"name"`
	Value  interface{} `yaml:"value"`
	scalar *GlobalConfigScalarUpdate
}

func (param *globalConfigPatchScalarParameter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type SP globalConfigPatchScalarParameter
	var field *SP = (*SP)(param)
	if err := unmarshal(&field); err != nil {
		return err
	}
	paramType := parameterType(param.Name)
	switch parameterCategory(param.Name) {
	case SchemaScalar:
		data, err := paramType.valueFromYAML(param.Value)
		if err != nil {
			return errors.Wrapf(err, "Missing / bad value for parameter %q", param.Name)
		}
		param.scalar = &GlobalConfigScalarUpdate{Key: param.Name, Value: data}
	case SchemaList:
		return errors.Errorf("Bad category for parameter %q", param.Name)
	default:
		return errors.Errorf("Invalid parameter %q", param.Name)
	}
	return nil
}

// GlobalConfigPatchScalarParameter ---- end ----

// GlobalConfigPatch ---- start ----

// GlobalConfigPatch is the human supplied patch converted into GlobalConfigTransactions by cachecash-gcp merge.
type GlobalConfigPatch struct {
	// How many blocks to delay activation of this patch for : convenience for operators to set something in motion
	// relatively quickly. It is an error to have both Delay and ActivationHeight non-zero.
	Delay uint64 `yaml:"delay"`
	// Alternatively, set a specific height to activate at - used when the change is far enough out that an exact height
	// can be calculated and published without racing with the authoring / publishing of the patch.
	ActivationHeight uint64                           `yaml:"activationHeight"`
	Parameters       []globalConfigPatchParameterYAML `yaml:"parameters"`
}

// GlobalConfigSchemaYAML returns the NewGlobalConfigPatch for content provided. It will also return an error during
// parsing, if any.
func NewGlobalConfigPatch(content []byte) (*GlobalConfigPatch, error) {
	var p GlobalConfigPatch
	if err := yaml.UnmarshalStrict(content, &p); err != nil {
		return nil, errors.Wrap(err, "could not parse schema")
	}

	if p.Delay != 0 && p.ActivationHeight != 0 {
		return nil, errors.New("Only one of Delay and ActivationHeight may be set")
	}

	if p.Delay == 0 && p.ActivationHeight == 0 {
		return nil, errors.New("One of Delay and ActivationHeight must be set")
	}

	return &p, nil
}

// NewGlobalConfigSchemaFromFile parses a GlobalConfigPatch file. An error is returned on IO errors, or errors
// propogated from NewGlobalConfigPatch
func NewGlobalConfigPatchFromFile(filename string) (*GlobalConfigPatch, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read filename %v", filename)
	}

	return NewGlobalConfigPatch(content)
}

func (p *GlobalConfigPatch) ToTransaction(ctx context.Context, chain *Database) (*GlobalConfigTransaction, error) {
	height, err := chain.Height(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current height")
	}
	scalars := make([]GlobalConfigScalarUpdate, 0)
	lists := make([]GlobalConfigListUpdate, 0)
	for _, p := range p.Parameters {
		if p.scalar != nil {
			scalars = append(scalars, *p.scalar)
		} else if p.list != nil {
			lists = append(lists, *p.list)
		} else {
			return nil, errors.Errorf("invalid parameter %q has no scalar or list", p.Name)
		}
	}
	var ActivationBlockHeight uint64
	if p.Delay == 0 {
		ActivationBlockHeight = p.ActivationHeight
	} else {
		ActivationBlockHeight = height + p.Delay
	}
	result := &GlobalConfigTransaction{
		ActivationBlockHeight: ActivationBlockHeight,
		ScalarUpdates:         scalars,
		ListUpdates:           lists,
	}
	return result, nil
}

// GlobalConfigPatch ---- end ----
