package ledger

import (
	"go/format"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalConfigStateYAML(t *testing.T) {
	y := NewGlobalConfigStateYAML()
	assert.NotNil(t, y.Scalars)
	assert.NotNil(t, y.Lists)
}

// GlobalConfigSchema ---- start ----

func referenceSchema() (*GlobalConfigSchema, error) {
	i := []byte(`
    version: 1
    parameters:
      - name: KeyA
        category: scalar
        default: 12
        type: uint64
      - name: KeyC
        category: list
        default:
        - -1
        - 15
        type: int64
      - name: KeyD
        category: scalar
        default: YWJjZGVmZwo=
        type: "[]byte"
    `)
	return newGlobalConfigSchema(i)
}

func TestParseSchemaGood(t *testing.T) {
	schema, err := referenceSchema()
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	for _, p := range schema.Parameters {
		switch p.Name() {
		case "KeyA":
			assert.Equal(t, &schemaParameterScalar{name: "KeyA", _default: []uint8{0xc}, _type: "uint64"}, p)
		case "KeyC":
			assert.Equal(t, &schemaParameterList{name: "KeyC", _default: [][]uint8{[]uint8{0x1}, []uint8{0x1e}}, _type: "int64"}, p)
		case "KeyD":
			assert.Equal(t, &schemaParameterScalar{name: "KeyD", _default: []uint8{0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0xa}, _type: "[]byte"}, p)
		default:
			assert.Failf(t, "Unexpected parameter", "%+v", p)
		}
	}
}

func TestParseSchemaBad(t *testing.T) {
	_, err := NewGlobalConfigSchemaFromFile("nonexistent")
	assert.EqualError(t, err, "could not read filename nonexistent: open nonexistent: no such file or directory")

	i := []byte(`
    version: 1
    parameters:
      - name: missingtype
        category: scalar
    `)
	schema, err := newGlobalConfigSchema(i)
	assert.Regexp(t, "Invalid parameter type", err)
	assert.Nil(t, schema)

	i = []byte(`
    version: 1
    parameters:
      - name: badtype
        category: scalar
        type: bad
    `)
	schema, err = newGlobalConfigSchema(i)
	assert.Regexp(t, "bad", err)
	assert.Nil(t, schema)

	i = []byte(`
    version: 1
    parameters:
      - name: missingcategory
        type: int64
    `)
	schema, err = newGlobalConfigSchema(i)
	assert.Nil(t, schema)
	assert.Regexp(t, "Invalid category", err)

	i = []byte(`
    version: 1
    parameters:
      - name: badcategory
        category: bad
        type: int64
    `)
	schema, err = newGlobalConfigSchema(i)
	assert.Regexp(t, "bad", err)
	assert.Nil(t, schema)
}

func TestSchemaGenerate(t *testing.T) {
	schema, err := referenceSchema()
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	content, err := schema.Generate()
	assert.NoError(t, err)
	expected, err := format.Source([]byte(`
	package ledger

	func parameterCategory(param string) schemaParameterCategory {
		switch param {
		case "KeyA":
			 return SchemaScalar
		case "KeyC":
			return SchemaList
		case "KeyD":
			return SchemaScalar
		default:
			return SchemaInvalidCategory
		}
	}

	func parameterType(param string) schemaParameterType {
		switch param {
		case "KeyA":
			 return SchemaUInt64
		case "KeyC":
			return SchemaInt64
		case "KeyD":
			return SchemaBytes
		default:
			return SchemaInvalidType
		}
	}

	func (s *GlobalConfigState) GetKeyA() (uint64, error) {
		stored, ok := s.Scalars["KeyA"]
		if !ok {
			return 12, nil
		}
		return gcpUnmarshalUInt64(stored)
	}

	func (s *GlobalConfigState) GetKeyC() ([]int64, error) {
		stored, ok := s.Lists["KeyC"]
		if !ok {
			return []int64{-1, 15}, nil
		}
		res := make([]int64, len(stored))
		for i, serialised := range stored {
			var err error
			res[i], err = gcpUnmarshalInt64(serialised)
			if err != nil {
				return res, err
			}
		}
		return res, nil
	}

	func (s *GlobalConfigState) GetKeyD() ([]byte, error) {
		stored, ok := s.Scalars["KeyD"]
		if !ok {
			return []byte("abcdefg\n"), nil
		}
		return stored, nil
	}
	`))
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(content))
}

// GlobalConfigSchema ---- end ----
