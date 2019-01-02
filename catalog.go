package cachecash

/*
N.B.: DEPRECATED: This will be removed in favor of the `catalog` subpackage.
*/

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type ContentCatalog interface {
	GetObjectByPath(ctx context.Context, path string) (ContentObject, error)

	ObjectsByPath() map[string]ContentObject
}

type contentCatalog struct {
	// objects maps paths to content objects.
	objects map[string]ContentObject
}

var _ ContentCatalog = (*contentCatalog)(nil)

func NewCatalogFromDir(path string) (ContentCatalog, error) {
	prefixLen := len(path)

	// Ensure that once we strip the prefix our file paths start with leading slashes.
	if strings.HasSuffix(path, "/") {
		prefixLen -= 1
	}

	cat := &contentCatalog{
		objects: make(map[string]ContentObject),
	}

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "error walking tree")
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		obj, err := NewContentBufferFromFile(path)
		if err != nil {
			return errors.Wrap(err, "failed to load file")
		}

		relPath := path[prefixLen:]
		cat.objects[relPath] = obj
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk directory")
	}

	return cat, nil
}

func (cat *contentCatalog) GetObjectByPath(ctx context.Context, path string) (ContentObject, error) {
	obj, ok := cat.objects[path]
	if !ok {
		// XXX: This will make it difficult to test for 404s.
		return nil, errors.New("no such object")
	}

	return obj, nil
}

// XXX: We should probably replace this with a better way to iterate over the objects.
func (cat *contentCatalog) ObjectsByPath() map[string]ContentObject {
	return cat.objects
}
