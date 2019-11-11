package ranger

func (cf *ConfigFormat) funcMap() map[string]interface{} {
	return map[string]interface{}{
		"size":          cf.size,
		"randomField":   cf.randomField,
		"isMarshalable": cf.isMarshalable,
		"add":           add,
	}
}

func add(i, j int) int {
	return i + j
}
