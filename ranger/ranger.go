package ranger

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ConfigFormat defines the format of the configuration file used to generate
// code -- this is the top level.
type ConfigFormat struct {
	// Package is the name of the package we're generating into.
	Package string `yaml:"package"`
	// Types is the listing of types. Each definition consists of a Type name as
	// map key and properties as values.
	Types map[string]ConfigType `yaml:"types"`

	// MaxByteRange defines how large our byte arrays can be.. if they are larger
	// an error is returned.
	MaxByteRange uint64 `yaml:"max_byte_range"`

	decls declarations
}

// ConfigType is the type definition wrapper; used for specifying member fields
// as well as any struct-level operations to apply such as validations or
// filtering.
type ConfigType struct {
	// Fields is a list of single keyed hash tables that correspond member names
	// to type definitions.
	Fields []map[string]*ConfigTypeDefinition `yaml:"fields"`

	TypeName string `yaml:"-"` // populated by parse
}

// ConfigTypeDefinition is the definition of an individual type member.
type ConfigTypeDefinition struct {
	// StructureType is the type of data structure; this may be "struct" or
	// "array". This determines how the field is marshaled by wrapping length
	// headers over array sections for ease of reading.
	StructureType string `yaml:"structure_type"`
	// ValueType is the type that we'll use for the actual marshaling. The
	// ValueType must conform to ranger.Marshaler or be a built in type (uint,
	// string, etc) that we support marshaling to/from natively.
	ValueType string `yaml:"value_type"`
	// Interface is a polymorphic type definition; see ConfigInterface for more.
	Interface *ConfigInterface `yaml:"interface,omitempty"`
	// Match is a list of matching rules for validations.
	Match ConfigMatch `yaml:"match,omitempty"`
	// Require is a list of requirements for validations.
	Require ConfigRequire `yaml:"require,omitempty"`
	// Marshal if false, will not marshal to the byte array.
	Marshal *bool `yaml:"marshal,omitempty"`
	// InlineStruct defines this type as inline in the struct. It must still
	// marshal independently.
	InlineStruct bool `yaml:"inline_struct"`

	FieldName    string `yaml:"-"` // populated by parse
	MaxByteRange uint64 `yaml:"-"` // populated by parse
	TypeName     string `yaml:"-"` // populated by parse
	Item         bool   `yaml:"-"` // populated by itemValue
}

// ConfigInterface defines a polymorphic type to marshal. It requires an input
// type, an output method for returning the type, and a series of cases for
// what member corresponds to what type.
type ConfigInterface struct {
	// Output method. This is a method which is called on the member to determine
	// the type information for embedding into the marshal.
	Output string `yaml:"output"`
	// Input type. This is a type we can unmarshal out of a byte stream to
	// determine what type in cases to use.
	Input string `yaml:"input"`
	// Cases is a list of cases; this is a list of one-key maps that correspond
	// comparisons in case statement, and result in a type marshal of the value.
	// Types specified in this manner must still conform to ranger.Marshaler.
	Cases []map[string]string `yaml:"cases"`
}

// ConfigMatch specifies matching rules for validations. Matching means in this
// context, that two fields must match each other in some way.
type ConfigMatch struct {
	// LengthOfField indicates the field must match the Member field specified in the value.
	LengthOfField string `yaml:"length_of_field"`
}

// ConfigRequire is hard requirements for validations.
type ConfigRequire struct {
	// MaxLength means the field must have a length no longer than this.
	MaxLength uint64 `yaml:"max_length"`
	// Length means the field must be exactly this length.
	Length uint64 `yaml:"length"`
	// Static ensures that the value is of a fixed size for integer types.
	Static bool `yaml:"static"`
}

// Parse parses the content and returns a ConfigFormat; if an error is returned
// ConfigFormat is invalid.
func Parse(content []byte) (*ConfigFormat, error) {
	var cf ConfigFormat

	if err := yaml.UnmarshalStrict(content, &cf); err != nil {
		return nil, err
	}

	return cf.editParams(), cf.validate()
}

func (cf *ConfigFormat) validate() error {
	if cf.MaxByteRange == 0 {
		return errors.New("max_byte_range cannot be 0")
	}

	for typName, typ := range cf.Types {
		for _, item := range typ.Fields {
			for key, value := range item {
				if value.Require.MaxLength == 0 && value.Require.Length == 0 {
					return errors.Errorf("%s.%s is missing a required length parameter: either specify `length` or `max_length`", typName, key)
				}
			}
		}
	}

	return nil
}

func (cf *ConfigFormat) editParams() *ConfigFormat {
	for typName, typ := range cf.Types {
		typ.TypeName = typName
		for _, item := range typ.Fields {
			for key, value := range item {
				value.FieldName = key
				value.TypeName = typName
				value.MaxByteRange = cf.MaxByteRange
			}
		}
	}

	cf.decls = declarations{}

	return cf
}

// ParseFile parses the content in a file specified by filename and returns a
// ConfigFormat; if an error is returned ConfigFormat is invalid.
func ParseFile(filename string) (*ConfigFormat, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Parse(content)
}

func (cf *ConfigFormat) generate(thisTemplate string) ([]byte, error) {
	tpl := template.New("template").Funcs(cf.funcMap())
	tpl, err := tpl.Parse(thisTemplate)
	if err != nil {
		return nil, err
	}

	byt := bytes.NewBuffer(nil)

	if err := tpl.Execute(byt, cf); err != nil {
		return nil, err
	}

	return format.Source(byt.Bytes())
}

// GenerateCode generates the source code.
func (cf *ConfigFormat) GenerateCode() ([]byte, error) {
	return cf.generate(goTemplate)
}

// GenerateTest generates the test code.
func (cf *ConfigFormat) GenerateTest() ([]byte, error) {
	return cf.generate(testTemplate)
}

// GenerateFuzz generates the fuzzer code.
func (cf *ConfigFormat) GenerateFuzz() ([]byte, error) {
	return cf.generate(fuzzTemplate)
}
