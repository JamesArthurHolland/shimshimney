package main

import (
	"log"

	"github.com/shimshimney/example/common"
)

func main() {
	if err := common.Serve("backend-2"); err != nil {
		log.Fatal(err)
	}
}
