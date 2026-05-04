package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"cairn/internal/remoteindex"
)

func main() {
	root := flag.String("root", envDefault("CAIRN_WORKSPACE_ROOT", "/workspace"), "workspace root to index")
	addr := flag.String("addr", envDefault("CAIRN_INDEXER_ADDR", ":8080"), "listen address")
	flag.Parse()

	service := remoteindex.NewService(*root)
	log.Printf("cairn local indexer listening on %s for %s", *addr, *root)
	if err := http.ListenAndServe(*addr, service.Handler()); err != nil {
		log.Fatal(err)
	}
}

func envDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: cairn-indexer [--root DIR] [--addr ADDR]\n")
		flag.PrintDefaults()
	}
}
