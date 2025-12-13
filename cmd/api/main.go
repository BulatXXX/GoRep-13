package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"singularity.com/pprof-lab/internal/work"
)

func main() {

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		defer work.TimeIt("FibFast(38) - после")()
		n := 38
		res := work.FibFast(n)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(fmtInt(res)))
	})

	log.Println("Server on :8080; pprof on /debug/pprof/")
	runtime.SetMutexProfileFraction(1)
	log.Fatal(http.ListenAndServe(":8080", nil)) // важно: nil, а не mux
}

func fmtInt(v int) string { return fmt.Sprintf("%d\n", v) }
