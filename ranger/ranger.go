package ranger

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/gobuffalo/packr"
)

// ConfigFormat defines the format of the configuration file used to generate
// code -- this is the top level.
type ConfigFormat struct {
	// Package is the name of the package we're generating into.
	Package string `yaml:"package"`
	// Types is the listing of types. Each definition consists of a Type name as
	// map key and properties as values.
	Types       map[string]*ConfigType `yaml:"types"`
	nativeTypes map[string]Type

	// MaxByteRange defines how large our byte arrays can be.. if they are larger
	// an error is returned.
	MaxByteRange uint64 `yaml:"max_byte_range"`

	// Comment adds a comment to the package declaration.
	Comment string `yaml:"comment"`

	templates *template.Template
}

// ConfigType is the type definition wrapper; used for specifying member fields
// as well as any struct-level operations to apply such as validations or
// filtering.
type ConfigType struct {
	// Fields is a list of fields in serialisation order. This is ordered because byte stability in the
	// serialisation format is a requirement.
	Fields []*ConfigTypeDefinition `yaml:"fields"`
	// Interface is a polymorphic type definition; see ConfigInterface for more. A type defined with an interface
	// configuration must have no fields defined.
	Interface *ConfigInterface `yaml:"interface,omitempty"`

	// Comment is a field to add a comment to the type's declaration.
	Comment string `yaml:"comment"`

	TypeName string `yaml:"-"` // populated by parse

	cf *ConfigFormat // populated by editParams
}

// ConfigTypeDefinition is the definition of an individual type member.
type ConfigTypeDefinition struct {
	// FieldName is the name of the field in the containing type. While the yaml file can represent multiple fields with
	// the same name this is a semantic error and will produce bad code that won't compile.
	FieldName string `yaml:"name"`
	// StructureType is the type of data structure; this may be "struct" or
	// "array". This determines how the field is marshaled by wrapping length
	// headers over array sections for ease of reading.
	StructureType string `yaml:"structure_type"`
	// ValueType is the type that we'll use for the actual marshaling. The
	// ValueType must conform to ranger.Marshaler or be a built in type (uint,
	// string, etc) that we support marshaling to/from natively.
	ValueType string `yaml:"value_type"`
	// Match is a list of matching rules for validations.
	Match ConfigMatch `yaml:"match,omitempty"`
	// Require is a list of requirements for validations.
	Require ConfigRequire `yaml:"require,omitempty"`
	// Marshal if false, will not marshal in or out.
	Marshal *bool `yaml:"marshal,omitempty"`
	// Embedded defines this field as embedded in the struct.
	Embedded bool `yaml:"embedded"`

	// Comment is a field to add a comment to the field's declaration.
	Comment string `yaml:"comment"`

	MaxByteRange uint64 `yaml:"-"` // populated by parse
	TypeName     string `yaml:"-"` // populated by parse

	cf *ConfigFormat // populated by editParams
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

	cf.populateNativeTypes()

	return cf.editParams(), cf.validate()
}

func (cf *ConfigFormat) validate() error {
	// TODO: some of these could become type calls (ask the type/field to validate itself)
	if cf.MaxByteRange == 0 {
		return errors.New("max_byte_range cannot be 0")
	}

	for typName, typ := range cf.Types {
		if len(typ.Fields) > 0 && typ.Interface != nil {
			return errors.Errorf("%s is invalid: both fields and an interface defined", typName)
		}
		for _, field := range typ.Fields {
			if field.Marshal != nil && !*field.Marshal {
				continue
			}
			field_type, err := field.GetType()
			if err != nil {
				return err
			}
			has_len, err := field_type.HasLen(field.FieldInstance())
			if err != nil {
				return err
			}
			if !has_len {
				if field.Require.MaxLength != 0 || field.Require.Length != 0 {
					return errors.Errorf("%s.%s is invalid; contains a length but is not a container type", typName, field.FieldName)
				}
			} else if field.Require.MaxLength == 0 && field.Require.Length == 0 {
				return errors.Errorf("%s.%s is missing a required length parameter: either specify `length` or `max_length`", typName, field.FieldName)
			}
			if field.Require.Static && (!field.IsNativeType() || field.IsBytesType()) {
				return errors.Errorf("%s.%s cannot be static: only applicable to integral types", typName, field.FieldName)
			}
			if field.Embedded && field.StructureType == "array" {
				return errors.Errorf("%s.%s cannot both be embedded and an array", typName, field.FieldName)
			}
			if field.ValueType == "string" && field.Require.Length != 0 {
				return errors.Errorf("%s.%s strings cannot have fixed widths", typName, field.FieldName)
			}
		}
	}

	return nil
}

func (cf *ConfigFormat) editParams() *ConfigFormat {
	for typName, typ := range cf.Types {
		typ.TypeName = typName
		typ.SetConfigFormat(cf)
		for _, field := range typ.Fields {
			if field.StructureType == "" {
				field.StructureType = "scalar"
			}

			field.TypeName = typName
			field.MaxByteRange = cf.MaxByteRange
			field.SetConfigFormat(cf)
		}
	}

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

// loadTemplates loads a family of templates which can mutually refer to each other for reuse
func (cf *ConfigFormat) loadTemplates() error {
	if cf.templates != nil {
		return nil
	}
	t := template.New("")
	t.Funcs(cf.funcMap())
	box := packr.NewBox("./templates")
	err := box.Walk(func(name string, f packr.File) error {
		if !strings.HasSuffix(name, ".gotmpl") {
			return nil
		}
		s := f.String()
		if len(s) == 0 {
			return errors.Errorf("zero length template %s", name)
		}
		_, err := t.New(name).Parse(s)
		return err
	})
	if err != nil {
		return err
	}
	cf.templates = t
	return nil
}

func (cf *ConfigFormat) ExecuteString(name string, data interface{}) (string, error) {
	byt, err := cf.Execute(name, data)
	if err != nil {
		return "", errors.Wrapf(err, "failed to render in %s", name)
	}
	return string(byt), nil
}

// Execute a named template with some data and return the string it renders.
func (cf *ConfigFormat) Execute(name string, data interface{}) ([]byte, error) {
	byt := bytes.NewBuffer(nil)
	tpl := cf.templates.Lookup(name)
	if tpl == nil {
		return nil, errors.Errorf("No template %s", name)
	}
	if err := tpl.Execute(byt, data); err != nil {
		return nil, err
	}
	return byt.Bytes(), nil
}

func (cf *ConfigFormat) generate(name string) ([]byte, error) {
	err := cf.loadTemplates()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load templates")
	}
	byt, err := cf.Execute(name, cf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render template")
	}
	return format.Source(byt)
}

// GenerateCode generates the source code.
func (cf *ConfigFormat) GenerateCode() ([]byte, error) {
	return cf.generate("go.gotmpl")
}

// GenerateTest generates the test code.
func (cf *ConfigFormat) GenerateTest() ([]byte, error) {
	return cf.generate("test.gotmpl")
}

// GenerateFuzz generates the fuzzer code.
func (cf *ConfigFormat) GenerateFuzz() ([]byte, error) {
	return cf.generate("fuzz.gotmpl")
}

// GetType looks up a specific type in both the user defined types and the built in native type definitions
func (cf *ConfigFormat) GetType(name string) (Type, error) {
	result, ok := cf.Types[name]
	if ok {
		return result, nil
	}
	result2, ok := cf.nativeTypes[name]
	if ok {
		return result2, nil
	}
	return nil, errors.Errorf("Unknown type %s", name)
}
