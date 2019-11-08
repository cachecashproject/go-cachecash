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
		if !cf.isNativeType(value.ValueType) {
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
	} else if cf.isNativeType(value.ValueType) {
		return fmt.Sprintf("%s(rand.Uint64()&%s)", value.ValueType, cf.truncated(value.ValueType))
	}

	if value.Interface != nil {
		s := fmt.Sprintf("[]%s{", value.ValueType)
		for _, c := range value.Interface.Cases {
			for _, v := range c {
				s += fmt.Sprintf("&%s{", v)
				for _, field := range cf.Types[v].Fields {
					s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.randomField(field))
				}

				s += "},"
			}
		}
		s += fmt.Sprintf("}[rand.Int()%%%d]", len(value.Interface.Cases))

		return s
	}

	s := fmt.Sprintf("&%s{", value.ValueType)
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

// itemValue is a hack to get us to work within array ranges over a static item.
// looks ugly as we transition: this is going to be removed.
func (cf *ConfigFormat) itemValue(ctd *ConfigTypeDefinition) *ConfigTypeDefinition {
	result := &ConfigTypeDefinition{}
	*result = *ctd
	result.Item = true
	return result
}

func (cf *ConfigFormat) getIsInterface(ctd ConfigType) bool {
	for _, field := range ctd.Fields {
		if field.Interface != nil {
			return true
		}
	}

	return false
}

// DefaultValueFor creates a default of a type. This differs very slightly from
// a zero value in that a pointer type is initialised to a default value of that
// type rather than to nil.
func (cf *ConfigFormat) DefaultValueFor(value *ConfigTypeDefinition) string {
	if value.Interface != nil {
		s := fmt.Sprintf("[]%s{", value.ValueType)
		for _, c := range value.Interface.Cases {
			for _, v := range c {
				s += fmt.Sprintf("&%s{", v)
				for _, field := range cf.Types[v].Fields {
					s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.DefaultValueFor(field))
				}

				s += "},"
			}
		}
		s += fmt.Sprintf("}[rand.Int()%%%d]", len(value.Interface.Cases))

		return s
	}

	if value.StructureType == "array" {
		ptr := ""
		if !cf.isNativeType(value.ValueType) {
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
		s := fmt.Sprintf("&%s{", value.ValueType)
		if _, ok := cf.Types[value.ValueType]; ok {
			for _, field := range cf.Types[value.ValueType].Fields {
				s += fmt.Sprintf("\n%s: %s,", field.FieldName, cf.DefaultValueFor(field))
			}
		}
		s += "\n}"

		return s
	}
}

func (cf *ConfigFormat) isNativeType(typ string) bool {
	switch typ {
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

func (cf *ConfigFormat) getUnmarshaler(typ, v string, item, static bool) string {
	if item {
		v += "[i]"
	}

	if static {
		switch typ {
		case valueTypeUint8:
			return fmt.Sprintf("%s = data[n]\nn += 1\n", v)
		case valueTypeUint16:
			return fmt.Sprintf("%s = binary.LittleEndian.Uint16(data[n:])\nn += 2\n", v)
		case valueTypeUint32:
			return fmt.Sprintf("%s = binary.LittleEndian.Uint32(data[n:])\nn += 4\n", v)
		case valueTypeUint64:
			return fmt.Sprintf("%s = binary.LittleEndian.Uint64(data[n:])\nn += 8\n", v)
		default:
			return "<BROKEN> // fix your templates -- this should be trapped elsewhere"
		}
	}

	doCast := func(cast string) string {
		return fmt.Sprintf("iL, ni = binary.Uvarint(data[n:])\nif ni <= 0 {\nreturn 0, ranger.ErrShortRead\n}\n%s = %s(iL)", v, cast)
	}

	switch typ {
	case valueTypeUint8:
		return doCast("uint8")
	case valueTypeUint16:
		return doCast("uint16")
	case valueTypeUint32:
		return doCast("uint32")
	case valueTypeUint64:
		return doCast("")
	default:
		return "<BROKEN> // fix your templates -- this should be trapped elsewhere"
	}
}

func (cf *ConfigFormat) getLengthMarshalerSpecial(typ, v string) string {
	switch typ {
	case valueTypeString, valueTypeBytes:
		return fmt.Sprintf("n += binary.PutUvarint(data[n:], uint32(len(%s)))", v)
	default:
		return fmt.Sprintf("n += binary.PutUvarint(data[n:], uint32(%s.Size()))", v)
	}
}

func (cf *ConfigFormat) getLengthMarshaler(typ, v string, static bool) string {
	if static {
		switch typ {
		case valueTypeUint8, valueTypeUint16, valueTypeUint32, valueTypeUint64:
			return "" // no size calculation is required
		default:
			return cf.getLengthMarshalerSpecial(typ, v)
		}
	}

	switch typ {
	case valueTypeUint8, valueTypeUint16, valueTypeUint32, valueTypeUint64:
		if typ != valueTypeUint64 {
			v = fmt.Sprintf("uint64(%s)", v)
		}
		return fmt.Sprintf("n += binary.PutUvarint(data[n:], ranger.UvarintSize(%s))", v)
	default:
		return cf.getLengthMarshalerSpecial(typ, v)
	}
}

func (cf *ConfigFormat) getMarshalerSpecial(typ, v string) string {
	switch typ {
	case valueTypeString, valueTypeBytes:
		return fmt.Sprintf("copy(data[n:n+len(%s)], %s)", v, v)
	default:
		return fmt.Sprintf("%s.MarshalTo(data[n:n+%s.Size()])", v, v)
	}
}

func (cf *ConfigFormat) getMarshaler(typ, v string, static bool) string {
	if static {
		switch typ {
		case valueTypeUint8:
			return fmt.Sprintf("data[n] = %s", v)
		case valueTypeUint16:
			return fmt.Sprintf("binary.LittleEndian.PutUint16(data[n:], %s)", v)
		case valueTypeUint32:
			return fmt.Sprintf("binary.LittleEndian.PutUint32(data[n:], %s)", v)
		case valueTypeUint64:
			return fmt.Sprintf("binary.LittleEndian.PutUint64(data[n:], %s)", v)
		default:
			return cf.getMarshalerSpecial(typ, v)
		}
	}

	switch typ {
	case valueTypeUint8, valueTypeUint16, valueTypeUint32, valueTypeUint64:
		if typ != valueTypeUint64 {
			v = fmt.Sprintf("uint64(%s)", v)
		}
		return fmt.Sprintf("binary.PutUvarint(data[n:], %s)", v)
	default:
		return cf.getMarshalerSpecial(typ, v)
	}
}

func (cf *ConfigFormat) getLengthSpecial(typ, v string) string {
	switch typ {
	case valueTypeString, valueTypeBytes:
		return fmt.Sprintf("ranger.UvarintSize(uint64(len(%s))) + len(%s)", v, v)
	default:
		return fmt.Sprintf("ranger.UvarintSize(uint64(%s.Size())) + %s.Size()", v, v)
	}
}

func (cf *ConfigFormat) getLength(typ, v string, static bool) string {
	if static {
		switch typ {
		case valueTypeUint8:
			return "1"
		case valueTypeUint16:
			return "2"
		case valueTypeUint32:
			return "4"
		case valueTypeUint64:
			return "8"
		default:
			return cf.getLengthSpecial(typ, v)
		}
	}

	switch typ {
	case valueTypeUint8, valueTypeUint16, valueTypeUint32, valueTypeUint64:
		if typ != valueTypeUint64 {
			v = fmt.Sprintf("uint64(%s)", v)
		}
		return fmt.Sprintf("ranger.UvarintSize(%s)", v)
	default:
		return cf.getLengthSpecial(typ, v)
	}
}
