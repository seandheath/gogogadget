package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/thatisuday/commando"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func server(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	dir, err := flags["dir"].GetString()
	check(err)

	port, err := flags["port"].GetString()
	check(err)

	fmt.Printf("Starting server for directory %s on port %s\n\n", dir, port)
	fs := http.FileServer(http.Dir(dir))
	log.Fatal(http.ListenAndServe(":"+port, fs))

}

func main() {

	wd, err := os.Getwd()
	check(err)
	commando.
		SetExecutableName("gogogadget").
		SetVersion("0.0.1").
		SetDescription("This tool provides utilities to facilitate penetration testing on multiple architectures.")

	commando.
		Register("server").
		SetShortDescription("starts a server").
		SetDescription("This command starts a server with the specified options.").
		AddFlag("dir,d", "directory to serve", commando.String, wd).
		AddFlag("port,p", "port to serve on", commando.String, "8080").
		SetAction(server)

	commando.Parse(nil)

}
