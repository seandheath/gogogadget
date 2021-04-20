package main

import (
	"fmt"
	"io"
	"log"
	"net"
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
	fmt.Println(http.ListenAndServe(":"+port, fs))

}

func download(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	url := args["url"].Value

	outfile, err := flags["outfile"].GetString()
	check(err)

	// Send the request
	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()

	// Make the file
	out, err := os.Create(outfile)
	check(err)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	check(err)
}

func pivot(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	targetAddress := args["target"].Value
	port := args["port"].Value
	protocol, err := flags["protocol"].GetString()
	check(err)

	// Start Server
	incoming, err := net.Listen(protocol, fmt.Sprintf(":%s", port))
	fmt.Println("done with listen")
	check(err)
	fmt.Printf("server running %s\n", port)

	// Accept Connection
	for {
		client, err := incoming.Accept()
		check(err)
		go handleRequest(client, targetAddress, protocol)

	}
}

func handleRequest(client net.Conn, targetAddress string, protocol string) {
	fmt.Printf("client '%v' connected!\n", client.RemoteAddr())

	// Dial out to the target

	target, err := net.Dial(protocol, targetAddress)
	check(err)
	fmt.Printf("connection established to %v\n", target.RemoteAddr())
	go CopyIO(client, target)
	go CopyIO(target, client)
}

func CopyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(src, dest)
}

func main() {

	wd, err := os.Getwd()
	check(err)

	name, err := os.Executable()
	check(err)

	commando.
		SetExecutableName(name).
		SetVersion("0.0.1").
		SetDescription("This tool provides utilities to facilitate penetration testing on multiple architectures.")

	// File server
	commando.
		Register("server").
		SetShortDescription("starts a server").
		SetDescription("This command starts a server with the specified options.").
		AddFlag("dir,d", "directory to serve", commando.String, wd).
		AddFlag("port,p", "port to serve on", commando.String, "8080").
		SetAction(server)

	// Download file command
	commando.
		Register("download").
		SetShortDescription("downloads a file from a URL").
		SetDescription("This command downloads the file at a provided URL").
		AddArgument("url", "target URL", "").
		AddFlag("path,p", "output file", commando.String, wd+"/outfile").
		SetAction(download)

	//Pivot command
	commando.
		Register("pivot").
		SetShortDescription("sets up a forwarding service").
		SetDescription("Sets up a forwarding service for attacking through a compromised host").
		AddArgument("target", "target (<host>:<port>)", "").
		AddArgument("port", "port", "").
		AddFlag("protocol,p", "protocol, defaults to tcp", commando.String, "tcp").
		SetAction(pivot)

	commando.Parse(nil)

}
