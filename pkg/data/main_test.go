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

	data, err := Parse(fileContent)
	if err != nil {
		t.Error(err)
	}

	spew.Dump(data)
}

func TestImportAlt(t *testing.T) {
	filePath := path.Join("..", "..", "testdata", "dataset.dat")

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Error(err)
	}

	data, err := Parse(fileContent)
	if err != nil {
		t.Error(err)
	}

	spew.Dump(data)
}
