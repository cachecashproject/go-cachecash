package ledger

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"go/format"
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// GlobalConfigSchemaYAML ---- start ----

// A globalConfigSchemaYAML defines the available global configuration parameters in the config file. This is parsed into
// a more convenient internal representation, the GlobalConfigSchema.
type globalConfigSchemaYAML struct {
	// What version of the schema - currently must be 1
	Version uint `yaml:"version"`
	// Parameters
	Parameters []schemaParameterYAML `yaml:"parameters"`
}

// GlobalConfigSchemaYAML returns the GlobalConfigSchema for content provided. It will also return an error during
// parsing, if any.
func newGlobalConfigSchema(content []byte) (*GlobalConfigSchema, error) {
	var yamlSchema globalConfigSchemaYAML
	if err := yaml.UnmarshalStrict(content, &yamlSchema); err != nil {
		return nil, errors.Wrap(err, "could not parse schema")
	}

	schema, err := yamlSchema.toSchema()

	if err != nil {
		return nil, errors.Wrap(err, "Failed schema validation")
	}

	return schema, nil
}

// NewGlobalConfigSchemaFromFile parses a GlobalConfigSchemaYAML file into a GlobalConfigSchema. An error is returned on IO errors, or errors
// propogated from NewGlobalConfigSchemaFromBytes
func NewGlobalConfigSchemaFromFile(filename string) (*GlobalConfigSchema, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read filename %v", filename)
	}

	return newGlobalConfigSchema(content)
}

// Validate determines if the GlobalConfigSchema is valid.
func (schema *globalConfigSchemaYAML) toSchema() (*GlobalConfigSchema, error) {
	if schema.Version != 1 {
		return nil, errors.Errorf("Bad schema version %d", schema.Version)
	}
	// check uniqueness and backfill compat
	parameters := make(map[string]schemaParameter)
	result := &GlobalConfigSchema{Parameters: parameters}

	for _, p := range schema.Parameters {
		p := p.Param
		p.Validate()
		_, ok := parameters[p.Name()]
		if ok {
			return nil, errors.Errorf("Duplicate parameter %s", p.Name())
		}
		parameters[p.Name()] = p
	}

	return result, nil
}

// GlobalConfigSchemaYAML ---- end ----

// GlobalConfigSchema ---- start ----

// A GlobalConfigSchema defines the available global configuration parameters in an in-memory convenient fashion.
type GlobalConfigSchema struct {
	// Parameters
	Parameters map[string]schemaParameter
}

func (s *GlobalConfigSchema) Generate() ([]byte, error) {
	t, err := template.New("").Parse(`package ledger

	func parameterCategory(param string) schemaParameterCategory {
		switch param {
        {{- range .Parameters}}
		case "{{.Name}}":
			 return {{.Category.CategoryCode}}
        {{- end}}
		default:
			return SchemaInvalidCategory
		}
	}

	func parameterType(param string) schemaParameterType {
		switch param {
        {{- range .Parameters}}
		case "{{.Name}}":
			 return {{.Type.TypeCode}}
        {{- end}}
		default:
			return SchemaInvalidType
		}
	}

	{{range .Parameters}}
	func (s *GlobalConfigState) Get{{.Name}}() ({{.TypeCode}}, error) {
		stored, ok := s.{{.StateAttribute}}["{{.Name}}"]
		if !ok {
			return {{.DefaultCode}}, nil
		}
		{{.UnmarshalCode}}
	}
    {{end}}
	`)
	if err != nil {
		return nil, err
	}
	byt := bytes.NewBuffer(nil)
	if err := t.Execute(byt, s); err != nil {
		return nil, err
	}
	bytes := byt.Bytes()
	// return bytes, nil
	return format.Source(bytes)
}

// GlobalConfigSchema ---- end ----

// SchemaParameterCategory ---- start ----

type schemaParameterCategory string

const (
	SchemaInvalidCategory schemaParameterCategory = "invalid"
	SchemaScalar          schemaParameterCategory = "scalar"
	SchemaList            schemaParameterCategory = "list"
)

func (category *schemaParameterCategory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var field string
	if err := unmarshal(&field); err != nil {
		return err
	}
	c := schemaParameterCategory(field)
	switch c {
	case SchemaScalar, SchemaList:
		*category = c
	default:
		return errors.New(fmt.Sprintf("Invalid parameter category: %s", field))
	}
	return nil
}

func (category *schemaParameterCategory) CategoryCode() string {
	switch *category {
	case SchemaScalar:
		return "SchemaScalar"
	case SchemaList:
		return "SchemaList"
	default:
		return "SchemaInvalidCategory"
	}
}

// SchemaParameterCategory ---- end ----

// schemaParameterType ---- start ----
type schemaParameterType string

const (
	SchemaInvalidType schemaParameterType = "invalid"
	SchemaUInt64      schemaParameterType = "uint64"
	SchemaInt64       schemaParameterType = "int64"
	SchemaBytes       schemaParameterType = "[]byte"
)

// valueFromYAML is the inverse of yamlFromValue - it creates a serialised []byte array from the in-memory
// representation that resulted from parsing a YAML file - either a schema (for defaults), or a patch. Unknown types
// cannot exist in these files, so are treated as errors (which is different to yamlFromValue where unknown types may
// exist in the chain and need to be handled).
func (paramType *schemaParameterType) valueFromYAML(yaml interface{}) ([]byte, error) {
	switch *paramType {
	case SchemaUInt64:
		// Can't figure out how to tell yaml that its really a uint64.
		v, ok := yaml.(int)
		if !ok {
			return nil, errors.Errorf("Missing / bad default %t", yaml)
		}
		return gcpMarshalUInt64(uint64(v)), nil
	case SchemaInt64:
		// Can't figure out how to tell yaml that its really an int64.
		v, ok := yaml.(int)
		if !ok {
			return nil, errors.Errorf("Missing / bad default %t", yaml)
		}
		return gcpMarshalInt64(int64(v)), nil
	case SchemaBytes:
		encoded, ok := yaml.(string)
		if !ok {
			return nil, errors.New("Missing / bad default")
		}
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, errors.Wrap(err, "badly encoded default")
		}
		return data, nil
	default:
		return nil, errors.New("Invalid parameter type")
	}
}

func (paramType *schemaParameterType) CodeFor(serialised []byte) (string, error) {
	var res string
	switch *paramType {
	case SchemaUInt64:
		t, err := gcpUnmarshalUInt64(serialised)
		if err != nil {
			return "", err
		}
		res = fmt.Sprintf("%d", t)
	case SchemaInt64:
		t, err := gcpUnmarshalInt64(serialised)
		if err != nil {
			return "", err
		}
		res = fmt.Sprintf("%d", t)
	case SchemaBytes:
		res = fmt.Sprintf("[]byte(%s)", strconv.Quote(string(serialised)))
	default:
		return "", errors.New("unreachable in SchemaParameterType.New")
	}
	return res, nil
}

func (paramType *schemaParameterType) TypeCode() string {
	switch *paramType {
	case SchemaUInt64:
		return "SchemaUInt64"
	case SchemaInt64:
		return "SchemaInt64"
	case SchemaBytes:
		return "SchemaBytes"
	default:
		return "SchemaInvalidType"
	}
}

func (paramType *schemaParameterType) UnmarshalCode() (string, error) {
	switch *paramType {
	case SchemaUInt64:
		return "return gcpUnmarshalUInt64(stored)", nil
	case SchemaInt64:
		return "return gcpUnmarshalInt64(stored)", nil
	case SchemaBytes:
		return "return stored, nil", nil
	default:
		return "", errors.New("unknown parameter type")
	}
}

func (paramType *schemaParameterType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var field string
	if err := unmarshal(&field); err != nil {
		return err
	}
	Type := schemaParameterType(field)
	switch Type {
	case SchemaUInt64, SchemaInt64, SchemaBytes:
		*paramType = Type
	default:
		return errors.New(fmt.Sprintf("Invalid parameter type: %s", field))
	}
	return nil
}

func (paramType *schemaParameterType) Validate() error {
	if len(*paramType) == 0 {
		return errors.Errorf("Invalid parameter type '%s'", *paramType)
	}
	return nil
}

// yamlFromValue returns a human appropriate representation of the value stored in the chain with the given type.
// Unknown / invalid types are treated as []byte and base64 encoded to ensure that we are able to dump a
// GlobalConfigState to the user at any point in time - whether or not the parameter is currently in the schema.
func (paramType *schemaParameterType) yamlFromValue(stored []byte) (interface{}, error) {
	switch *paramType {
	case SchemaUInt64:
		t, err := gcpUnmarshalUInt64(stored)
		return t, err
	case SchemaInt64:
		t, err := gcpUnmarshalInt64(stored)
		return t, err
	case SchemaBytes:
		return base64.StdEncoding.EncodeToString(stored), nil
	default:
		return base64.StdEncoding.EncodeToString(stored), nil
	}
}

// SchemaParameterType ---- end ----

// SchemaParameter ---- start ----

// schemaParameter describes a single parameter in the GlobalConfigurationParameter schema.
type schemaParameter interface {
	Name() string
	Category() *schemaParameterCategory
	DefaultCode() (string, error)
	StateAttribute() string
	Type() *schemaParameterType
	TypeCode() (string, error)
	UnmarshalCode() (string, error)
	Validate() error
}

// SchemaParameter ---- end ----

// SchemaParameterYAML ---- start ----

// schemaParameterYAML provides the thunk to unmarshal a single parameter from the schema file
type schemaParameterYAML struct {
	Name     string                  `yaml:"name"`
	Category schemaParameterCategory `yaml:"category"`
	Default  interface{}             `yaml:"default"`
	Type     schemaParameterType     `yaml:"type"`
	Param    schemaParameter         `yaml:"-"`
}

func (param *schemaParameterYAML) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type SP schemaParameterYAML
	var field *SP = (*SP)(param)
	if err := unmarshal(&field); err != nil {
		return err
	}
	// This isn't quite as nice as it could be: it would be nicer to make the defaults match the type
	// - e.g. delegate to the Type and read an int for ints, and only base64 decode for actual []byte arrays.
	// OTOH this matches the actual storage in the chain, so its very consistent with the rest of the layers.
	switch param.Category {
	case SchemaScalar:
		data, err := param.Type.valueFromYAML(param.Default)
		if err != nil {
			return errors.Wrapf(err, "Missing / bad default for parameter %q", param.Name)
		}
		param.Param = &schemaParameterScalar{name: param.Name, _type: param.Type, _default: data}
	case SchemaList:
		encodeds, ok := param.Default.([]interface{})
		if !ok {
			return errors.Errorf("Missing / bad default for parameter %q", param.Name)
		}
		defaults := make([][]byte, 0, len(encodeds))
		for _, _default := range encodeds {
			data, err := param.Type.valueFromYAML(_default)
			if err != nil {
				return errors.Wrapf(err, "Missing / bad default for parameter %q", param.Name)
			}
			defaults = append(defaults, data)
		}
		param.Param = &schemaParameterList{name: param.Name, _type: param.Type, _default: defaults}
	default:
		return errors.Errorf("Invalid category %q", param.Category)
	}
	if err := param.Param.Validate(); err != nil {
		return err
	}
	return nil
}

// SchemaParameterYAML ---- end ----

// schemaParameterScalar ---- start ----
type schemaParameterScalar struct {
	name     string              `yaml:"name"`
	_default []byte              `yaml:"default"`
	_type    schemaParameterType `yaml:"type"`
}

var _ schemaParameter = (*schemaParameterScalar)(nil)

func (param *schemaParameterScalar) Name() string {
	return param.name
}

func (param *schemaParameterScalar) Category() *schemaParameterCategory {
	res := SchemaScalar
	return &res
}

func (param *schemaParameterScalar) DefaultCode() (string, error) {
	return param._type.CodeFor(param._default)
}

func (param *schemaParameterScalar) StateAttribute() string {
	return "Scalars"
}

func (param *schemaParameterScalar) Type() *schemaParameterType {
	return &param._type
}

func (param *schemaParameterScalar) TypeCode() (string, error) {
	return string(param._type), nil
}

func (param *schemaParameterScalar) UnmarshalCode() (string, error) {
	return param._type.UnmarshalCode()
}

func (param *schemaParameterScalar) Validate() error {
	if err := param._type.Validate(); err != nil {
		return err
	}
	if len(param.name) == 0 {
		return errors.Errorf("Invalid name parameter name '%s'", param.name)
	}
	return nil
}

// SchemaParameterScalar ---- end ----

// schemaParameterList ---- start ----
type schemaParameterList struct {
	name     string              `yaml:"name"`
	_default [][]byte            `yaml:"default"`
	_type    schemaParameterType `yaml:"type"`
}

var _ schemaParameter = (*schemaParameterList)(nil)

func (param *schemaParameterList) Name() string {
	return param.name
}

func (param *schemaParameterList) Category() *schemaParameterCategory {
	res := SchemaList
	return &res
}

func (param *schemaParameterList) DefaultCode() (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[]%s{", param._type))
	for i, serialised := range param._default {
		if i != 0 {
			sb.WriteString(", ")
		}
		s, err := param._type.CodeFor(serialised)
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	sb.WriteString("}")
	return sb.String(), nil
}

func (param *schemaParameterList) StateAttribute() string {
	return "Lists"
}

func (param *schemaParameterList) Type() *schemaParameterType {
	return &param._type
}

func (param *schemaParameterList) TypeCode() (string, error) {
	return fmt.Sprintf("[]%s", param._type), nil
}

func (param *schemaParameterList) UnmarshalCode() (string, error) {
	// Avoid generating a loop for arrays of bytes
	if param._type == SchemaBytes {
		return param._type.UnmarshalCode()
	}
	var unmarshal string
	switch *param.Type() {
	case SchemaUInt64:
		unmarshal = "gcpUnmarshalUInt64"
	case SchemaInt64:
		unmarshal = "gcpUnmarshalInt64"
	default:
		return "", errors.New("unknown parameter type")
	}
	typeCode, err := param.TypeCode()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`res := make(%s, len(stored))
	for i, serialised := range stored {
		var err error
		res[i], err = %s(serialised)
		if err != nil {
			return res, err
		}
	}
	return res, nil`, typeCode, unmarshal), nil
}

func (param *schemaParameterList) Validate() error {
	if err := param._type.Validate(); err != nil {
		return err
	}
	if len(param.name) == 0 {
		return errors.Errorf("Invalid name parameter name '%s'", param.name)
	}
	return nil
}

// SchemaParameterList ---- end ----
