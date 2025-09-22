package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(".")))
	log.Printf("POATCscan running at http://localhost%v", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
