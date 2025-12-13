package work

import (
	"log"
	"time"
)

func TimeIt(name string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s took %s", name, time.Since(start))
	}
}
