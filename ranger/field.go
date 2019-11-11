package ranger

import "fmt"

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
func (field *ConfigTypeDefinition) GetType() Type {
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

type FieldInstance struct {
	field *ConfigTypeDefinition
}

func (instance *FieldInstance) ConfigFormat() *ConfigFormat {
	return instance.field.ConfigFormat()
}

func (instance *FieldInstance) GetLength() uint64 {
	return instance.field.Require.Length
}

func (instance *FieldInstance) GetMaxLength() uint64 {
	if instance.field.Require.MaxLength > 0 {
		return instance.field.Require.MaxLength
	}
	return instance.field.MaxByteRange
}

func (instance *FieldInstance) HasLen() bool {
	return instance.field.StructureType == "array"
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
	return instance.field.Require.Static
}

// ItemInstance adapts a field for use in arrays
type ItemInstance struct {
	field *ConfigTypeDefinition
}

func (instance *ItemInstance) ConfigFormat() *ConfigFormat {
	return instance.field.ConfigFormat()
}

// The schema cannot specify the length of an item within an array.
func (instance *ItemInstance) GetLength() uint64 {
	return 0
}

// The schema cannot specify the maximum length of an item within an array,
// use the global maximum
func (instance *ItemInstance) GetMaxLength() uint64 {
	return instance.field.MaxByteRange
}

// The schema cannot specify arrays of arrays, so this is always false
func (instance *ItemInstance) HasLen() bool {
	return false
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
	return instance.field.Require.Static
}
