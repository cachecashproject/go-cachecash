package ranger

import "fmt"

// TypeInstance is an instance of a type
// e.g. Type == struct definition.
//      TypeInstance == usage - plain, in an array, or as a pointer.
type TypeInstance interface {
	// Get at the ConfigFormat for e.g. running templates
	ConfigFormat() *ConfigFormat
	// What if any Length is configured for 'instance'.
	// NB: this should move to being part of the type.
	GetLength() uint64
	// What max length is configured for 'instance'.
	// Should fall back to MaxByteRange (though Kevin has asked that that be
	// removed)
	GetMaxLength() uint64
	// Does the type support len()
	HasLen() bool
	// Is this a Reference
	IsPointer() bool
	// Whats the fully qualified name (for human errors)
	QualName() string
	// How is this instance address in code when reading in
	ReadSymbolName() string
	// How is this instance addressed in code when writing out
	WriteSymbolName() string
	// Is this instance statically sized? (Separate to static array lengths at
	// least for now)
	Static() bool
}

// A Type is a type for which we can emit (de)serialisation code.
// Built in types cover the basic native types. Users can compose these into
// custom types.
type Type interface {
	// Does the type support len()
	HasLen(TypeInstance) bool
	// MinimumSize returns the minimum serialized size of the type.
	MinimumSize(TypeInstance) uint64
	// The name of the type
	Name() string
	// PointerType returns whether the type instance is a value or a pointer to
	// value
	PointerType(TypeInstance) bool
	// Read returns code to deserialise an instance of the type
	Read(TypeInstance) string
	// WriteSize returns code to caculate the size of an instance of the type when serialized
	WriteSize(TypeInstance) string
	// Write returns code to serialise an instance of the
	Write(TypeInstance) string
}

// SetConfigFormat provides the type with a reference to the root of the
// structure for rendering templates etc
func (typ *ConfigType) SetConfigFormat(cf *ConfigFormat) {
	typ.cf = cf
}

func (typ *ConfigType) ConfigFormat() *ConfigFormat {
	return typ.cf
}

func (ct *ConfigType) IsInterface() bool {
	return ct.Interface != nil
}

// InterfaceAdapter creates an interface adapter for TypeInstance
func (typ *ConfigType) InterfaceAdapter(instance TypeInstance) TypeInstance {
	return &InputInstanceAdapter{
		wrapped: instance,
		typ:     typ,
	}
}

// Does the type support len()
func (ct *ConfigType) HasLen(instance TypeInstance) bool {
	return instance.HasLen()
}

// MinimumSize returns the minimum serialized size of the type.
func (ct *ConfigType) MinimumSize(instance TypeInstance) uint64 {
	if ct.IsInterface() {
		// The switch marker
		// TODO: switch to the marker + the minimum of the defined cases.
		return ct.cf.GetType(ct.Interface.Input).MinimumSize(
			ct.InterfaceAdapter(instance))
	}
	var minimum uint64
	for _, field := range ct.Fields {
		// roughly:
		// if structural, can delegate already
		// if variable, a varint
		// fixed length is a todo - today its still a varint otherwise
		instance := field.FieldInstance()
		field_type := field.GetType()
		if field_type.HasLen(instance) {
			// uvarint minimum size, to record a 0 length string/array
			minimum += 1
		} else {
			minimum += field_type.MinimumSize(instance)
		}
	}
	return minimum
}

// The name of the type
func (ct *ConfigType) Name() string {
	return ct.TypeName
}

// PointerType returns whether the type instance is a value or a pointer to
// value
func (ct *ConfigType) PointerType(instance TypeInstance) bool {
	return instance.IsPointer()
}

// Read returns code to deserialise an instance of the type
func (ct *ConfigType) Read(instance TypeInstance) string {
	if !ct.IsInterface() {
		return ct.cf.ExecuteString("readstruct.gotmpl", instance)
	}

	readInput := ct.cf.GetType(ct.Interface.Input).Read(ct.InterfaceAdapter(instance))
	readStruct := ct.cf.ExecuteString("readstruct.gotmpl", instance)

	return ct.cf.ExecuteString("readinterface.gotmpl", struct {
		Type       Type
		Instance   TypeInstance
		ReadInput  string
		ReadStruct string
	}{ct, instance, readInput, readStruct})

}

// WriteSize returns code to caculate the size of an instance of the type when serialized
func (ct *ConfigType) WriteSize(instance TypeInstance) string {
	if ct.IsInterface() {
		/* Interface serialialisation:
		   - the 'input' type, then the delegated case - currently only works with
			 Minimum / constant size types.
		*/
		// instance describes e.g. TransactionBody
		// interface.input tells us we want TransactionBody.TxType
		return fmt.Sprintf("%s + %s",
			ct.cf.GetType(ct.Interface.Input).WriteSize(ct.InterfaceAdapter(instance)),
			ct.cf.ExecuteString("interfacesize.gotmpl", instance))
	}
	return fmt.Sprintf("%s.Size()", instance.WriteSymbolName())
}

// Write returns code to serialise an instance of the
func (ct *ConfigType) Write(instance TypeInstance) string {
	if !ct.IsInterface() {
		/* XXX: Serialising a reference to a user defined type. These are expected
		   to self-deliimit
		   - this is one of the points to fix once things are neated up
		*/
		return ct.cf.ExecuteString("marshaltostruct.gotmpl", instance)
	}

	return fmt.Sprintf("%s\n%s",
		ct.cf.GetType(ct.Interface.Input).Write(ct.InterfaceAdapter(instance)),
		ct.cf.ExecuteString("marshaltostruct.gotmpl", instance))

}

type InputInstanceAdapter struct {
	wrapped TypeInstance
	typ     *ConfigType
}

func (instance *InputInstanceAdapter) ConfigFormat() *ConfigFormat {
	return instance.typ.ConfigFormat()
}

func (instance *InputInstanceAdapter) GetLength() uint64 {
	return instance.wrapped.GetLength()
}

func (instance *InputInstanceAdapter) HasLen() bool {
	// Schema provides no way to declare that the input type for an interface is
	// an array
	return false
}

func (instance *InputInstanceAdapter) IsPointer() bool {
	// XX: Is delegating appropriate?
	return instance.wrapped.IsPointer()
}

func (instance *InputInstanceAdapter) GetMaxLength() uint64 {
	return instance.wrapped.GetMaxLength()
}

func (instance *InputInstanceAdapter) QualName() string {
	return instance.wrapped.QualName()
}

func (instance *InputInstanceAdapter) ReadSymbolName() string {
	return "intf"
}

func (instance *InputInstanceAdapter) WriteSymbolName() string {
	return fmt.Sprintf("%s.%s()", instance.wrapped.WriteSymbolName(), instance.typ.Interface.Output)
}

func (instance *InputInstanceAdapter) Static() bool {
	// No provision in the schema for choosing this
	return true
}
