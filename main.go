package main

import (
	"fmt"
	"physicsGUI/pkg/gui"
)

func main() {
	fmt.Println("Hello, World!")

	// Start GUI (blocking so it needs to be called at the end)
	gui.Start()
}
