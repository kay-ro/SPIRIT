package data

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

var typeAssociationMap = map[string]FileParser{
	".dat": DefaultDatParser(),
}

// Import data from a file
func Import(data []byte, filename string) error {
	// TODO: implement

	fmt.Printf("Importierte Datei: %s, Größe: %d bytes\n", filename, len(data))

	dataType := strings.ToLower(path.Ext(filename))
	parser, ok := typeAssociationMap[dataType]
	if !ok {
		return errors.New(fmt.Sprintf("Can't import %s: No loader for %s files found", filename, dataType))
	} else {
		res, err := parser.tryParse(data)
		if err != nil {
			return errors.Join(errors.New(fmt.Sprintf("Import failed (%s)", filename)), err)
		}
		//TODO display loaded content
		fmt.Println(res)
	}
	return nil
}
