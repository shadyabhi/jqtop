package main

import (
	"fmt"

	randomdata "github.com/shadyabhi/go-randomdata"
)

func main() {
	for {
		fmt.Printf("{\"ttms\": %d, \"code\": %d, \"domain\": \"%s/%s\"}\n", randomdata.Number(100), randomdata.Number(599), randomdata.Domain(), randomdata.Noun())
	}
}
