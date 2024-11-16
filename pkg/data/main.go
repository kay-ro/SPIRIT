package data

import "fmt"

// Import data from a file
func Import(data []byte, filename string) error {
	// TODO: implement

	fmt.Printf("Importierte Datei: %s, Größe: %d bytes\n", filename, len(data))

	return nil
}
