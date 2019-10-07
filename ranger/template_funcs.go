package ranger

import "fmt"

func (cf *ConfigFormat) funcMap() map[string]interface{} {
	return map[string]interface{}{
		"native":        cf.isNativeType,
		"typeLength":    cf.getLength,
		"marshaler":     cf.getMarshaler,
		"marshalLength": cf.getLengthMarshaler,
		"itemValue":     cf.itemValue,
		"unmarshaler":   cf.getUnmarshaler,
		"isBytes":       cf.isBytesType,
		"declare":       cf.decls.declare,
		"zeroValue":     cf.getZeroValue,
		"isInterface":   cf.getIsInterface,
		"size":          cf.size,
		"truncated":     cf.truncated,
		"randomField":   cf.randomField,
		"add":           add,
	}
}

func add(i, j int) int {
	return i + j
}

type declarations map[string]map[string]map[string]struct{}

func (d declarations) declare(item bool, t, fun, typ, typtyp string) string {
	if _, ok := d[t]; !ok {
		d[t] = map[string]map[string]struct{}{}
	}

	if _, ok := d[t][fun]; !ok {
		d[t][fun] = map[string]struct{}{}
	}

	if _, ok := d[t][fun][typ]; !ok || item {
		if !item {
			d[t][fun][typ] = struct{}{}
		}
		return fmt.Sprintf("var %s %s", typ, typtyp)
	}

	return ""
}
