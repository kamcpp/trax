package main

import (
	"flag"

	"github.com/kamcpp/trax/pkg/daemons"
)

func main() {
	useInMemory := flag.Bool("in-memory-store", false, "use in-memory store instead of PostgreSQL")
	flag.Parse()
	daemons.RunTraxCtrl(*useInMemory)
}
