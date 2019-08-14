// Package kv implements a simple key/value store with backing drivers. Each
// storage system can operate on value types in an independent way, allowing
// for a variety of schemas to be represented.
package kv

// Driver describes the driver implementation of a k/v store's operations.
type Driver interface {
	Create(string, string, []byte) ([]byte, error)
	Delete(string, string, []byte) error
	Get(string, string) ([]byte, error)
	Set(string, string, []byte) ([]byte, error)
	CAS(string, string, []byte, []byte, []byte) ([]byte, error)
}

// Client is the typical end-user entry into the k/v system. This allows for
// the standard operations on drivers.
type Client struct {
	member string
	driver Driver
}

// NewClient creates a new client.
func NewClient(member string, driver Driver) *Client {
	return &Client{member: member, driver: driver}
}

// Delete deletes any key, regardless of type; pass a nonce to delete conditionally.
func (c *Client) Delete(key string, nonce []byte) error {
	return c.driver.Delete(c.member, key, nonce)
}

// GetBytes retrieves a key for the k/v store as marshaled data. The `out` argument
// must be a non-nil byte array.
func (c *Client) GetBytes(key string) ([]byte, error) {
	return c.driver.Get(c.member, key)
}

// SetBytes sets a value explicitly from marshaled data, with no checking.
func (c *Client) SetBytes(key string, value []byte) ([]byte, error) {
	return c.driver.Set(c.member, key, value)
}

// CASBytes implements compare-and-swap semantics with marshaled data.
func (c *Client) CASBytes(key string, nonce, origValue, value []byte) ([]byte, error) {
	return c.driver.CAS(c.member, key, nonce, origValue, value)
}
