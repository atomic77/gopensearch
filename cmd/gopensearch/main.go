package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/alecthomas/repr"
	"github.com/atomic77/gopensearch/pkg/server"
)

func main() {

	dbLoc := flag.String("db", "test.db", "Location of sqlite database")
	flag.Parse()

	s := &server.Server{
		Cfg: server.Config{
			DbLocation: *dbLoc,
		},
	}
	s.Init()
	log.Println(repr.String(s.Cfg))
	log.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", s.Router))
}
