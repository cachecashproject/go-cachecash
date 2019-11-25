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
	Read(TypeInstance) (string, error)
	// WriteSize returns code to caculate the size of an instance of the type when serialized
	WriteSize(TypeInstance) (string, error)
	// Write returns code to serialise an instance of the
	Write(TypeInstance) (string, error)
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
// This is the sum of the minimum size of the serialized fields of the type.
func (typ *ConfigType) MinimumSize(instance TypeInstance) uint64 {
	if typ.IsInterface() {
		// The switch marker
		// TODO: switch to the marker + the minimum of the defined cases.
		return typ.cf.GetType(typ.Interface.Input).MinimumSize(
			typ.InterfaceAdapter(instance))
	}
	var minimum uint64
	for _, field := range typ.Fields {
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
func (ct *ConfigType) Read(instance TypeInstance) (string, error) {
	if !ct.IsInterface() {
		return ct.cf.ExecuteString("readstruct.gotmpl", instance)
	}

	readInput, err := ct.cf.GetType(ct.Interface.Input).Read(ct.InterfaceAdapter(instance))
	if err != nil {
		return "", err
	}
	readStruct, err := ct.cf.ExecuteString("readstruct.gotmpl", instance)
	if err != nil {
		return "", err
	}

	return ct.cf.ExecuteString("readinterface.gotmpl", struct {
		Type       Type
		Instance   TypeInstance
		ReadInput  string
		ReadStruct string
	}{ct, instance, readInput, readStruct})
}

// WriteSize returns code to caculate the size of an instance of the type when serialized
func (ct *ConfigType) WriteSize(instance TypeInstance) (string, error) {
	if ct.IsInterface() {
		/* Interface serialialisation */
		// instance describes e.g. TransactionBody
		// interface.input tells us we want TransactionBody.TxType
		sizeCode, err := ct.cf.ExecuteString("interfacesize.gotmpl", instance)
		if err != nil {
			return "", err
		}
		interfaceSize, err := ct.cf.GetType(ct.Interface.Input).WriteSize(ct.InterfaceAdapter(instance))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s + %s", interfaceSize, sizeCode), nil
	}
	return fmt.Sprintf("%s.Size()", instance.WriteSymbolName()), nil
}

// Write returns code to serialise an instance of the
func (ct *ConfigType) Write(instance TypeInstance) (string, error) {
	marshalToStruct, err := ct.cf.ExecuteString("marshaltostruct.gotmpl", instance)
	if err != nil {
		return "", err
	}
	if !ct.IsInterface() {
		return marshalToStruct, nil
	}

	marshalInterfaceKey, err := ct.cf.GetType(ct.Interface.Input).Write(ct.InterfaceAdapter(instance))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s", marshalInterfaceKey, marshalToStruct), nil
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
