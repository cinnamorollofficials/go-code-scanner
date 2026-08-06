package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	generators := []string{
		"./cmd/gen-config-doc/main.go",
		"./cmd/gen-rule-catalog/main.go",
		"./cmd/gen-scanners-doc/main.go",
	}

	for _, g := range generators {
		fmt.Printf("Running generator %s...\n", g)
		cmd := exec.Command("go", "run", g)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running generator %s: %v\n", g, err)
			os.Exit(1)
		}
	}

	fmt.Println("All documentation references generated successfully!")
}
