// Copyright 2019 Branen Salmon.  All rights reserved.
// This software is licensed under the GNU GPL, version 3 or later.
// See LICENSE for details.

package xss

import (
	"fmt"
)

func Example() {
	client, err := NewClient()
	if err != nil {
		panic(err)
	}
	info, err := client.Query()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Enabled: %v\n", info.Enabled)
	fmt.Printf("Active: %v\n", info.Active)
	fmt.Printf("Kind: %v\n", info.Kind)
	fmt.Printf("Countdown: %v\n", info.Countdown)
	fmt.Printf("ActiveTime: %v\n", info.ActiveTime)
	fmt.Printf("IdleTime: %v\n", info.IdleTime)
}
