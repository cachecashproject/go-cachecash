package ranger

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

// Does the type support len()
func (ct *ConfigType) HasLen(_ TypeInstance) bool {
	panic("not implemented")
}

// MinimumSize returns the minimum serialized size of the type.
func (ct *ConfigType) MinimumSize(_ TypeInstance) uint64 {
	panic("not implemented")
}

// The name of the type
func (ct *ConfigType) Name() string {
	panic("not implemented")
}

// PointerType returns whether the type instance is a value or a pointer to
// value
func (ct *ConfigType) PointerType(_ TypeInstance) bool {
	panic("not implemented")
}

// Read returns code to deserialise an instance of the type
func (ct *ConfigType) Read(_ TypeInstance) string {
	panic("not implemented")
}

// WriteSize returns code to caculate the size of an instance of the type when serialized
func (ct *ConfigType) WriteSize(_ TypeInstance) string {
	panic("not implemented")
}

// Write returns code to serialise an instance of the
func (ct *ConfigType) Write(_ TypeInstance) string {
	panic("not implemented")
}
