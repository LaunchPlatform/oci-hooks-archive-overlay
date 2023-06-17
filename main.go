package main

import (
	"encoding/json"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"log"
	"os"
)

func main() {
	var state spec.State
	err := json.NewDecoder(os.Stdin).Decode(&state)
	if err != nil {
		log.Fatal(err)
	}

}
