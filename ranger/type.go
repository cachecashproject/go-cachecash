package ranger

import (
	"bytes"
	"fmt"
)

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
	HasLen() (bool, error)
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
	HasLen(TypeInstance) (bool, error)
	// MinimumSize returns the minimum serialized size of the type.
	MinimumSize(TypeInstance) (uint64, error)
	// The name of the type
	Name() string
	// PointerType returns whether the type instance is a value or a pointer to
	// value
	PointerType(TypeInstance) bool
	// Read returns code to deserialise an instance of the type
	Read(TypeInstance) (string, error)
	// Type returns the go code to describe the type - e.g. [32]byte.
	Type(TypeInstance) (string, error)
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
func (ct *ConfigType) HasLen(instance TypeInstance) (bool, error) {
	return instance.HasLen()
}

// MinimumSize returns the minimum serialized size of the type.
// This is the sum of the minimum size of the serialized fields of the type.
func (typ *ConfigType) MinimumSize(instance TypeInstance) (uint64, error) {
	if typ.IsInterface() {
		// The switch marker
		// TODO: switch to the marker + the minimum of the defined cases.
		int_type, err := typ.cf.GetType(typ.Interface.Input)
		if err != nil {
			return 0, err
		}
		return int_type.MinimumSize(typ.InterfaceAdapter(instance))
	}
	var minimum uint64
	for _, field := range typ.Fields {
		// roughly:
		// if structural, can delegate already
		// if variable, a varint
		// fixed length is a todo - today its still a varint otherwise
		instance := field.FieldInstance()
		field_type, err := field.GetType()
		if err != nil {
			return 0, err
		}
		has_len, err := field_type.HasLen(instance)
		if err != nil {
			return 0, err
		}
		if has_len {
			// uvarint minimum size, to record a 0 length string/array
			minimum += 1
		} else {
			field_size, err := field_type.MinimumSize(instance)
			if err != nil {
				return 0, err
			}
			minimum += field_size
		}
	}
	return minimum, nil
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

	int_type, err := ct.cf.GetType(ct.Interface.Input)
	if err != nil {
		return "", err
	}
	readInput, err := int_type.Read(ct.InterfaceAdapter(instance))
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

func (ct *ConfigType) Type(instance TypeInstance) (string, error) {
	var result bytes.Buffer
	write := func(s string) {
		result.Write([]byte(s))
	}
	if ct.PointerType(instance) && !ct.IsInterface() {
		write("*")
	}
	write(ct.TypeName)
	return result.String(), nil
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
		int_type, err := ct.cf.GetType(ct.Interface.Input)
		if err != nil {
			return "", err
		}
		interfaceSize, err := int_type.WriteSize(ct.InterfaceAdapter(instance))
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
	int_type, err := ct.cf.GetType(ct.Interface.Input)
	if err != nil {
		return "", err
	}
	marshalInterfaceKey, err := int_type.Write(ct.InterfaceAdapter(instance))
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

func (instance *InputInstanceAdapter) HasLen() (bool, error) {
	// Schema provides no way to declare that the input type for an interface is
	// an array
	return false, nil
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
