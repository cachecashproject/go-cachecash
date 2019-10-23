package templates

// Get is a short wrapper for `packr.Box.Find()` so we don't have to expose the
// box. Also casts to string since we always need that.
func Get(name string) (string, error) {
	byt, err := box.Find(name)
	return string(byt), err
}
