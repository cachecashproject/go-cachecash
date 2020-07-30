package cachecontrol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaxAge(t *testing.T) {
	cc := Parse("max-age=300")
	maxAge := 300 * time.Second
	assert.Equal(t, cc, CacheControl{
		MaxAge: &maxAge,
	})

	cc = Parse("max-age=1")
	maxAge = 1 * time.Second
	assert.Equal(t, cc, CacheControl{
		MaxAge: &maxAge,
	})

	cc = Parse("max-age=a")
	assert.Equal(t, cc, CacheControl{})

	cc = Parse("max-age")
	assert.Equal(t, cc, CacheControl{})
}

func TestImmutable(t *testing.T) {
	cc := Parse("max-age=300, immutable")
	maxAge := 300 * time.Second
	assert.Equal(t, cc, CacheControl{
		MaxAge:    &maxAge,
		Immutable: true,
	})

	cc = Parse("immutable")
	assert.Equal(t, cc, CacheControl{
		Immutable: true,
	})

	cc = Parse("public, immutable")
	assert.Equal(t, cc, CacheControl{
		Immutable: true,
	})

	cc = Parse("max-age=300")
	maxAge = 300 * time.Second
	assert.Equal(t, cc, CacheControl{
		MaxAge: &maxAge,
	})
}

func TestStaleIfError(t *testing.T) {
	cc := Parse("s-maxage=600,public, max-age=300,stale-if-error=1200")
	maxAge := 300 * time.Second
	sMaxAge := 600 * time.Second
	staleIfError := 1200 * time.Second
	assert.Equal(t, cc, CacheControl{
		MaxAge:       &maxAge,
		SMaxAge:      &sMaxAge,
		StaleIfError: &staleIfError,
	})
}
