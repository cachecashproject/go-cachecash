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
