package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/shadyabhi/jqtop"
)

func main() {
	// Profiling
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	jqtop.Start(nil)
}
