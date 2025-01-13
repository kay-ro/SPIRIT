package main

import (
	"fmt"
	"physicsGUI/pkg/gui"
	"physicsGUI/pkg/trigger"
)

func main() {
	fmt.Println("Hello, World!")

	// Initialize trigger for recalculating gui based on changes
	trigger.Init()

	// Start GUI (blocking so it needs to be called at the end)
	gui.Start()
}
