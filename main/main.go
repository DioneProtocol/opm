// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/afero"

	"github.com/DioneProtocol/opm/cmd"
)

func main() {
	opm, err := cmd.New(afero.NewOsFs())
	if err != nil {
		fmt.Printf("Failed to initialize the opm command: %s.\n", err)
		os.Exit(1)
	}

	if err := opm.Execute(); err != nil {
		fmt.Printf("Unexpected error %s.\n", err)
		os.Exit(1)
	}
}
