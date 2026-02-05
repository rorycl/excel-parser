package main

import (
	"context"
	"fmt"
	"os"
)

func main() {

	ctx := context.Background()
	app := NewApp()
	cmd := BuildCLI(app)
	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

}
