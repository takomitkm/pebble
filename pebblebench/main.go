// e26pebble: main.go - explicit-flush cost bench for the pebble Windows dir
// sync.  Lives inside the pebble checkout (copied from the rig by the
// workflow) so `go run ./pebblebench` compiles against that checkout's own
// code - no module plumbing needed.
//
// Usage: go run ./pebblebench -db <path> -flushes N -kpf K
//
// Each Flush() forces a memtable flush: SST creation + version edit appended
// and synced in the MANIFEST + the data-directory sync through OpenDir's
// handle.  The three variants differ only in vfs/default_windows.go:
// base = upstream a995e890, dironly = 1e285158, dirroot = 52534185.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/pebble"
)

func main() {
	dbPath := flag.String("db", "", "database path")
	flushes := flag.Int("flushes", 300, "number of explicit Flush() calls")
	kpf := flag.Int("kpf", 200, "keys written per flush round")
	flag.Parse()
	if *dbPath == "" {
		fmt.Println("usage: pebblebench -db <path>")
		os.Exit(2)
	}
	opts := &pebble.Options{
		DisableAutomaticCompactions: true,
	}
	db, err := pebble.Open(*dbPath, opts)
	if err != nil {
		fmt.Printf("open fail: %v\n", err)
		os.Exit(1)
	}
	val := make([]byte, 100)
	for i := range val {
		val[i] = 'x'
	}
	start := time.Now()
	for f := 0; f < *flushes; f++ {
		for i := 0; i < *kpf; i++ {
			key := fmt.Sprintf("k%08d", f*(*kpf)+i)
			if err := db.Set([]byte(key), val, nil); err != nil {
				fmt.Printf("set fail: %v\n", err)
				os.Exit(1)
			}
		}
		if err := db.Flush(); err != nil {
			fmt.Printf("flush fail: %v\n", err)
			os.Exit(1)
		}
	}
	el := time.Since(start)
	ms := el.Seconds() * 1000
	fmt.Printf("%s: %.1f ms total | %.4f ms/flush | flushes=%d keys=%d\n",
		*dbPath, ms, ms/float64(*flushes), *flushes, *flushes**kpf)
	if err := db.Close(); err != nil {
		fmt.Printf("close fail: %v\n", err)
		os.Exit(1)
	}
}
