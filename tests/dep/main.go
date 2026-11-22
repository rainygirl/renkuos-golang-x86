// Needs a module from the network: go mod tidy && go build.
package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("renkuos"))
	fmt.Println(id, "version", id.Version())
}
