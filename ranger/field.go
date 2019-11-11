package ranger

import "fmt"

// QualName returns the qualified name of the field. For instance Transaction.Version
func (field *ConfigTypeDefinition) QualName() string {
	return fmt.Sprintf("%s.%s", field.TypeName, field.FieldName)
}
