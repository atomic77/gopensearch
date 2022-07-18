package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/atomic77/gopensearch/pkg/server"
)

func main() {

	s := &server.Server{
		Cfg: server.Config{
			DbLocation: "test.db",
		},
	}
	s.Init()
	// count := flag.Int("asdf", 5, "count")
	flag.Parse()
	log.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", s.Router))
}
