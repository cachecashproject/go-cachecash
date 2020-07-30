package ranger

import (
	"fmt"

	"github.com/pkg/errors"
)

// QualName returns the qualified name of the field. For instance Transaction.Version
func (field *ConfigTypeDefinition) QualName() string {
	return fmt.Sprintf("%s.%s", field.TypeName, field.FieldName)
}

// SetConfigFormat provides a reference to the ConfigFormat instance for lookups
// of referenced type definitions
func (field *ConfigTypeDefinition) SetConfigFormat(cf *ConfigFormat) {
	field.cf = cf
}

// ConfigFormat returns the config format for the occasional cases where we need
// global context
func (field *ConfigTypeDefinition) ConfigFormat() *ConfigFormat {
	return field.cf
}

func (field *ConfigTypeDefinition) GetInterface() *ConfigInterface {
	return field.cf.Types[field.ValueType].Interface
}

func (field *ConfigTypeDefinition) IsInterface() bool {
	if field.IsNativeType() {
		return false
	}
	return field.GetInterface() != nil
}

// GetType returns the type of this field. For array fields this is the type of
// the items of the array.
func (field *ConfigTypeDefinition) GetType() (Type, error) {
	return field.cf.GetType(field.ValueType)
}

func (field *ConfigTypeDefinition) SymbolName() string {
	return fmt.Sprintf("obj.%s", field.FieldName)
}

// FieldInstance returns a TypeInstance implementation adapted to the field in
// structure form.
func (field *ConfigTypeDefinition) FieldInstance() TypeInstance {
	return &FieldInstance{field: field}
}

// ItemInstance returns a TypeInstance implementation adapted to the field in
// array item form.
func (field *ConfigTypeDefinition) ItemInstance() TypeInstance {
	return &ItemInstance{field: field}
}

// MaybeItemInstance returns ItemInstance if the field is an array, FieldInstance otherwise.
func (field *ConfigTypeDefinition) MaybeItemInstance() TypeInstance {
	if field.StructureType == "array" {
		return field.ItemInstance()
	} else {
		return field.FieldInstance()
	}
}

// NeedsInitializer returns true if the zero value isn't usable for unmarshaling into.
func (field *ConfigTypeDefinition) NeedsInitializer() (bool, error) {
	// structs and variable length bytes
	if !field.IsNativeType() {
		return true, nil
	}
	if field.ValueType != "[]byte" {
		return false, nil
	}
	instance := field.MaybeItemInstance()
	length, err := instance.GetLength()
	if err != nil {
		return false, err
	}
	// For unsized types we need to make a slice
	return length == 0, nil
}

// Initializer returns the value to assign for unmarshaling into
func (field *ConfigTypeDefinition) Initializer() (string, error) {
	if !field.IsNativeType() {
		return fmt.Sprintf("&%s{}", field.ValueType), nil
	}
	instance := field.MaybeItemInstance()
	length, err := instance.GetLength()
	if err != nil {
		return "", err
	}
	if field.ValueType != "[]byte" || length != 0 {
		return "", errors.Errorf("no initializer needed %s", field.QualName())
	}
	return "[]byte{}", nil
}

type FieldInstance struct {
	field *ConfigTypeDefinition
}

func (instance *FieldInstance) ConfigFormat() *ConfigFormat {
	return instance.field.ConfigFormat()
}

func (instance *FieldInstance) GetLength() (uint64, error) {
	return instance.field.Require.Length, nil
}

func (instance *FieldInstance) GetMaxLength() (uint64, error) {
	return instance.field.Require.MaxLength, nil
}

func (instance *FieldInstance) HasLen() (bool, error) {
	return instance.field.StructureType == "array", nil
}

func (instance *FieldInstance) IsPointer() bool {
	return !instance.field.Embedded
}

func (instance *FieldInstance) QualName() string {
	return instance.field.QualName()
}

func (instance *FieldInstance) ReadSymbolName() string {
	return instance.WriteSymbolName()
}

func (instance *FieldInstance) WriteSymbolName() string {
	return fmt.Sprintf("obj.%s", instance.field.FieldName)
}

func (instance *FieldInstance) Static() bool {
	return instance.field.Static
}

// ItemInstance adapts a field for use in arrays
type ItemInstance struct {
	field *ConfigTypeDefinition
}

func (instance *ItemInstance) ConfigFormat() *ConfigFormat {
	return instance.field.ConfigFormat()
}

func (instance *ItemInstance) GetLength() (uint64, error) {
	if instance.field.ItemRequire == nil {
		return 0, errors.Errorf("item_require nil %s", instance.field.QualName())
	}
	return instance.field.ItemRequire.Length, nil
}

func (instance *ItemInstance) GetMaxLength() (uint64, error) {
	if instance.field.ItemRequire == nil {
		return 0, errors.Errorf("item_require nil %s", instance.field.QualName())
	}
	return instance.field.ItemRequire.MaxLength, nil
}

// The schema cannot specify arrays of arrays, so this is always false unless the contained type supports len.
func (instance *ItemInstance) HasLen() (bool, error) {
	typ, err := instance.field.GetType()
	if err != nil {
		return false, err
	}
	return typ.HasLen(instance.field.FieldInstance())
}

// We only support pointers to structs in arrays
func (instance *ItemInstance) IsPointer() bool {
	return true
}

func (instance *ItemInstance) QualName() string {
	return instance.field.QualName()
}

func (instance *ItemInstance) ReadSymbolName() string {
	return fmt.Sprintf("obj.%s[i]", instance.field.FieldName)
}

func (instance *ItemInstance) WriteSymbolName() string {
	return "item"
}

func (instance *ItemInstance) Static() bool {
	return instance.field.Static
}
