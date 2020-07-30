package catalog

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func join(t *testing.T, base string, new string) (string, error) {
	u, err := NewHTTPUpstream(logrus.New(), base, time.Minute)
	if err != nil {
		return "", err
	}
	return u.(*httpUpstream).upstreamURL(new)
}

func assertURLJoin(t *testing.T, base string, new string, expected string) {
	newURL, err := join(t, base, new)
	assert.Nil(t, err, "failed to join urls")
	assert.Equal(t, expected, newURL)
}

func assertURLJoinError(t *testing.T, base string, new string) {
	newURL, err := join(t, base, new)
	assert.NotNil(t, err, "joining urls didn't fail as expected: "+newURL)
	assert.Equal(t, "", newURL)
}

func TestUpstreamResolveSimple(t *testing.T) {
	assertURLJoin(t, "http://upstream:80", "/file0.bin", "http://upstream:80/file0.bin")
	assertURLJoin(t, "http://upstream:80/", "/file0.bin", "http://upstream:80/file0.bin")
}

func TestUpstreamInvalid(t *testing.T) {
	assertURLJoinError(t, "http://upstream:80", "file0.bin")
	assertURLJoinError(t, "http://upstream:80/", "file0.bin")
	assertURLJoinError(t, "http://upstream:80", "../../file0.bin")
}

func TestUpstreamWithFolder(t *testing.T) {
	assertURLJoin(t, "http://upstream:80/sub/directory", "/file0.bin", "http://upstream:80/sub/directory/file0.bin")
	assertURLJoin(t, "http://upstream:80/sub/directory/", "/file0.bin", "http://upstream:80/sub/directory/file0.bin")
}

func TestUpstreamDenyAbsolute(t *testing.T) {
	assertURLJoinError(t, "http://upstream:80", "http://example.com/file0.bin")
	assertURLJoinError(t, "http://upstream:80", "/http://example.com/file0.bin")
	assertURLJoinError(t, "http://upstream:80", "///example.com/file0.bin")
	assertURLJoinError(t, "http://upstream:80", "/http:///file0.bin")
	assertURLJoinError(t, "http://upstream:80", "///:80/file0.bin")
	assertURLJoin(t, "http://upstream:80", "////file0.bin", "http://upstream:80///file0.bin")
}

func TestUpstreamDirectoryTraversal(t *testing.T) {
	assertURLJoin(t, "http://upstream:80/", "/../../file0.bin", "http://upstream:80/file0.bin")
	assertURLJoinError(t, "http://upstream:80/sub/directory/", "/../../file0.bin")
	assertURLJoinError(t, "http://upstream:80", "///upstream:8080/")
}
