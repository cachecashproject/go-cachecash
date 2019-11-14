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

// FieldInstance returns a TypeInstance implementation adapted to the field in
// structure form.
func (field *ConfigTypeDefinition) FieldInstance() TypeInstance {
	return &FieldInstance{field: field}
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
