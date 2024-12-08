package data

import (
	"os"
	"path"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestImport(t *testing.T) {
	filePath := path.Join("..", "..", "testdata", "syntheticdataset.dat")

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Error(err)
	}

	data, err := Import(fileContent, filePath)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 measurement, got %d", len(data))
	}

	spew.Dump(data)
}
