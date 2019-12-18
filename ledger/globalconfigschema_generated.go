package ledger

func parameterCategory(param string) schemaParameterCategory {
	switch param {
	case "GlobalConfigSigningKeys":
		return SchemaList
	default:
		return SchemaInvalidCategory
	}
}

func parameterType(param string) schemaParameterType {
	switch param {
	case "GlobalConfigSigningKeys":
		return SchemaBytes
	default:
		return SchemaInvalidType
	}
}

func (s *GlobalConfigState) GetGlobalConfigSigningKeys() ([][]byte, error) {
	stored, ok := s.Lists["GlobalConfigSigningKeys"]
	if !ok {
		return [][]byte{[]byte("")}, nil
	}
	return stored, nil
}
