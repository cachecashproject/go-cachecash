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
		"isMarshalable": cf.isMarshalable,
		"add":           add,
	}
}

func add(i, j int) int {
	return i + j
}

// See declare for more information.
type declarations map[string]map[string]map[string]struct{}

// declare allows the template user to declare variables in a variety of
// contexts, without trampling go's single declaration or unused variable
// rules.
//
// the generation template specifies a few idioms that are not obvious to the
// casual reader:
//
// * n is the length of the total transaction
// * ni is the length of the isolated transaction (pulling a uvarint; the length it read, etc)
// * iLen is the computed length of an object
// * item is the object when working within a for loop w/ range values.
//
// In our declare case, the parameters are that `item` indicates whether or not
// we are in that item loop (with a different scope), t is the outer type name,
// fun is the name of the function, varName is the name of the variable we want
// to declare and typ is the name of the type that variable has.
// The returned string is either empty, or contains a variable declaration at
// the first point we have seen this declaration.
func (d declarations) declare(item bool, t, fun, varName, typ string) string {
	if _, ok := d[t]; !ok {
		d[t] = map[string]map[string]struct{}{}
	}

	if _, ok := d[t][fun]; !ok {
		d[t][fun] = map[string]struct{}{}
	}

	if _, ok := d[t][fun][varName]; !ok || item {
		if !item {
			d[t][fun][varName] = struct{}{}
		}
		return fmt.Sprintf("var %s %s", varName, typ)
	}

	return ""
}
