package ranger

import "fmt"

const (
	valueTypeUint8  = "uint8"
	valueTypeUint16 = "uint16"
	valueTypeUint32 = "uint32"
	valueTypeUint64 = "uint64"
	valueTypeBytes  = "[]byte"
	valueTypeString = "string"
)

func (cf *ConfigFormat) isMarshalable(value ConfigTypeDefinition) bool {
	if value.Marshal != nil {
		return *value.Marshal
	}

	return true
}

func (cf *ConfigFormat) randomField(value *ConfigTypeDefinition) string {
	if value.Marshal != nil && !*value.Marshal {
		return cf.DefaultValueFor(value)
	}

	if value.StructureType == "array" {
		typ := value.ValueType
		if !value.IsNativeType() {
			typ = "*" + typ
		}
		s := fmt.Sprintf("[]%s{", typ)
		l := value.Require.Length
		if l == 0 {
			if value.Require.MaxLength != 0 && value.Require.MaxLength < 10 {
				l = value.Require.MaxLength
			} else {
				l = 10
			}
		}

		for i := 0; i < int(l); i++ {
			v := *value
			v.StructureType = "scalar"

			s += cf.randomField(&v) + ","
		}
		s += "}"

		return s
	}

	if value.IsBytesType() {
		if value.Require.Length != 0 {
			return fmt.Sprintf("%s(genRandom(%d))", value.ValueType, value.Require.Length)
		} else if value.Require.MaxLength != 0 {
			return fmt.Sprintf("%s(genRandom(rand.Int()%%%d))", value.ValueType, value.Require.MaxLength)
		} else {
			return "<HERE> // should have a length"
		}
	} else if value.IsNativeType() {
		return fmt.Sprintf("%s(rand.Uint64()&%s)", value.ValueType, cf.truncated(value.ValueType))
	}

	if value.IsInterface() {
		s := fmt.Sprintf("[]%s{", value.ValueType)
		for _, c := range value.GetInterface().Cases {
			for _, v := range c {
				s += fmt.Sprintf("&%s{", v)
				for _, field := range cf.Types[v].Fields {
					s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.randomField(field))
				}

				s += "},"
			}
		}
		s += fmt.Sprintf("}[rand.Int()%%%d]", len(value.GetInterface().Cases))

		return s
	}

	var reference string
	if !value.Embedded {
		reference = "&"
	}
	s := fmt.Sprintf("%s%s{", reference, value.ValueType)
	if _, ok := cf.Types[value.ValueType]; ok {
		for _, field := range cf.Types[value.ValueType].Fields {
			s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.randomField(field))
		}
	}
	s += "\n}"

	return s
}

func (cf *ConfigFormat) truncated(typ string) string {
	switch typ {
	case valueTypeUint8:
		return "math.MaxUint8"
	case valueTypeUint16:
		return "math.MaxUint16"
	case valueTypeUint32:
		return "math.MaxUint32"
	case valueTypeUint64:
		return "math.MaxUint64"
	default:
		return "<HERE> // truncation is for uint types only. generator error"
	}
}

func (cf *ConfigFormat) size() string {
	return fmt.Sprintf("%d", len(cf.Types))
}

// DefaultValueFor creates a default of a type. This differs very slightly from
// a zero value in that a pointer type is initialised to a default value of that
// type rather than to nil.
func (cf *ConfigFormat) DefaultValueFor(value *ConfigTypeDefinition) string {
	if value.IsInterface() {
		s := fmt.Sprintf("[]%s{", value.ValueType)
		for _, c := range value.GetInterface().Cases {
			for _, v := range c {
				s += fmt.Sprintf("&%s{", v)
				for _, field := range cf.Types[v].Fields {
					s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.DefaultValueFor(field))
				}

				s += "},"
			}
		}
		s += fmt.Sprintf("}[rand.Int()%%%d]", len(value.GetInterface().Cases))

		return s
	}

	if value.StructureType == "array" {
		ptr := ""
		if !value.IsNativeType() {
			ptr = "*"
		}
		return fmt.Sprintf("make([]%s%s, %d)", ptr, value.ValueType, value.Require.Length)
	}

	switch value.ValueType {
	case valueTypeUint8, valueTypeUint16, valueTypeUint32, valueTypeUint64:
		return "0"
	case valueTypeString:
		return `""`
	case valueTypeBytes:
		return fmt.Sprintf("make([]byte, %d)", value.Require.Length)
	default:
		var reference string
		if !value.Embedded {
			reference = "&"
		}
		s := fmt.Sprintf("%s%s{", reference, value.ValueType)
		if _, ok := cf.Types[value.ValueType]; ok {
			for _, field := range cf.Types[value.ValueType].Fields {
				s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.DefaultValueFor(field))
			}
		}
		s += "\n}"

		return s
	}
}

func (value *ConfigTypeDefinition) IsNativeType() bool {
	switch value.ValueType {
	case valueTypeBytes, valueTypeString, valueTypeUint8, valueTypeUint16, valueTypeUint32, valueTypeUint64:
		return true
	default:
		return false
	}
}

func (value *ConfigTypeDefinition) IsBytesType() bool {
	switch value.ValueType {
	case valueTypeString, valueTypeBytes:
		return true
	default:
		return false
	}
}

func (cf *ConfigFormat) populateNativeTypes() {
	cf.nativeTypes = make(map[string]Type)
	cf.nativeTypes["uint8"] = &UInt8{}
	cf.nativeTypes["uint16"] = &Integral{2, "uint16", "Uint16", "math.MaxUint16"}
	cf.nativeTypes["uint32"] = &Integral{4, "uint32", "Uint32", "math.MaxUint32"}
	// For 64 bit types the mask could be omitted and no bit masking operations
	// done, as a future enhancement.
	cf.nativeTypes["uint64"] = &Integral{8, "uint64", "Uint64", "math.MaxUint64"}
	cf.nativeTypes["[]byte"] = &Strings{"[]byte", ""}
	cf.nativeTypes["string"] = &Strings{"string", "string"}
}

// A static sized UInt8
type UInt8 struct{}

func (typ *UInt8) HasLen(instance TypeInstance) bool {
	return instance.HasLen()
}

func (typ *UInt8) MinimumSize(instance TypeInstance) (uint64, error) {
	return 1, nil
}

func (typ *UInt8) Name() string {
	return "uint8"
}

func (typ *UInt8) PointerType(instance TypeInstance) bool {
	return false
}

func (typ *UInt8) Read(instance TypeInstance) (string, error) {
	return fmt.Sprintf("%s = data[n]\nn += 1\n", instance.ReadSymbolName()), nil
}

func (typ *UInt8) WriteSize(instance TypeInstance) (string, error) {
	return "1", nil
}

func (typ *UInt8) Write(instance TypeInstance) (string, error) {
	return fmt.Sprintf("data[n] = %s\n    n += 1", instance.WriteSymbolName()), nil
}

// Integral represents integers that can be static or variable length.
type Integral struct {
	staticLen  uint64
	name       string
	staticName string
	mask       string
}

func (typ *Integral) HasLen(instance TypeInstance) bool {
	return instance.HasLen()
}

func (typ *Integral) MinimumSize(instance TypeInstance) (uint64, error) {
	if instance.Static() {
		return typ.staticLen, nil
	}
	return 1, nil
}

func (typ *Integral) Name() string {
	return typ.name
}

func (typ *Integral) PointerType(instance TypeInstance) bool {
	return false
}

func (typ *Integral) Read(instance TypeInstance) (string, error) {
	if instance.Static() {
		return fmt.Sprintf("%s = binary.LittleEndian.%s(data[n:])\nn += %d\n",
			instance.ReadSymbolName(), typ.staticName, typ.staticLen), nil
	}
	return instance.ConfigFormat().ExecuteString("readuvarint.gotmpl", struct {
		SymbolName string
		Cast       string
		QualName   string
		Mask       string
	}{instance.ReadSymbolName(), typ.name, instance.QualName(), typ.mask})
}

func (typ *Integral) WriteSize(instance TypeInstance) (string, error) {
	if instance.Static() {
		return fmt.Sprint(typ.staticLen), nil
	}
	return fmt.Sprintf("ranger.UvarintSize(uint64(%s))", instance.WriteSymbolName()), nil
}

func (typ *Integral) Write(instance TypeInstance) (string, error) {
	if instance.Static() {
		return fmt.Sprintf("binary.LittleEndian.Put%s(data[n:], %s)\n    n += %d",
			typ.staticName, instance.WriteSymbolName(), typ.staticLen), nil
	}
	return fmt.Sprintf("n += binary.PutUvarint(data[n:], uint64(%s))", instance.WriteSymbolName()), nil
}

type Strings struct {
	name string
	cast string
}

func (typ *Strings) HasLen(instance TypeInstance) bool {
	return true
}

func (typ *Strings) MinimumSize(instance TypeInstance) (uint64, error) {
	return 1, nil // the varchar header
}

func (typ *Strings) Name() string {
	return typ.name
}

func (typ *Strings) PointerType(instance TypeInstance) bool {
	return false
}

func (typ *Strings) Read(instance TypeInstance) (string, error) {
	return instance.ConfigFormat().ExecuteString("readstring.gotmpl", struct {
		TI   TypeInstance
		Cast string
	}{instance, typ.cast})
}

func (typ *Strings) WriteSize(instance TypeInstance) (string, error) {
	symbolName := instance.WriteSymbolName()
	if instance.GetLength() != 0 {
		return fmt.Sprintf("int(%d)", instance.GetLength()), nil
	} else {
		return fmt.Sprintf("ranger.UvarintSize(uint64(len(%s))) + len(%s)", symbolName, symbolName), nil
	}
}

func (typ *Strings) Write(instance TypeInstance) (string, error) {
	symbolName := instance.WriteSymbolName()
	if instance.GetLength() != 0 {
		return fmt.Sprintf("n += copy(data[n:], %s)", symbolName), nil
	} else {
		return fmt.Sprintf("n += binary.PutUvarint(data[n:], uint64(len(%s)))\n    n += copy(data[n:n+len(%s)], %s)",
			symbolName, symbolName, symbolName), nil
	}
}
