// SPDX-License-Identifier: CC0-1.0

package main

import (
	"fmt"

	"github.com/nzions/sharedgolibs/pkg/waitlib"
)

func main() {
	fmt.Println("waitlib example - demonstrating usage")
	fmt.Println("This will start a wait process that updates its title with version and uptime")
	fmt.Println("")

	// Start waitlib with a custom version
	waitlib.Run("example-v1.2.3")
}
