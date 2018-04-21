package main

import "bufio"
import "os"
import logpkg "log"
import "github.com/robertkrimen/otto"
import "strings"
import "fmt"
import "io/ioutil"

var COMMANDS = make(map[string]func([]string) error) // Map of native commands.
// JS can add to COMMANDS with def(command, function)
var vm = otto.New() // JS environment
var log = logpkg.New(ioutil.Discard, "", 0) // Change ioutil.Discard to something else for logging
var commandBuffer []string // When something doesn't read as a command, add to buffer

// Use from builtin functions made available to JS to signal problem or success
var t, _ = vm.ToValue(true)
var f, _ = vm.ToValue(false)

// addCommand is defined to be invoked from JavaScript as def
// with the first argument being the defined command and the 
// second argument being a JavaScript function to be invoked
// on that command.
func addCommand(call otto.FunctionCall) otto.Value {
	log.Println("Enter addCommand")
	command, err := call.Argument(0).ToString()
	if err != nil {
		return f
	}
	log.Println("Adding command ", command)
	function := call.Argument(1)
	if err != nil {
		return f
	}
	COMMANDS[command] = func(cb []string) error {
		cbs := strings.Join(cb, "\n")
		_, err := function.Call(f,cbs)
		if err != nil {
			return err
		}
		return nil
	}
	log.Printf("%#v", COMMANDS)
	return t
}

func runEditorCommand(call otto.FunctionCall) otto.Value {
	log.Println("Entering runEditorCommand")
	command, err := call.Argument(0).ToString()
	if err != nil {
		return f
	}
	var args = make([]string, len(call.ArgumentList))
	for i, arg := range call.ArgumentList {
		args[i], err = arg.ToString()
		if err != nil {
			return f
		}
	}
	if err = COMMANDS[command](args[1:]); err != nil {
		v, e := vm.ToValue(err)
		if e != nil {
			return f
		} else {
			return v
		}
	}
	return t
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	vm.Set("def", addCommand)
	vm.Set("e", runEditorCommand)

	COMMANDS["quit"] = func(_ []string) error {
		os.Exit(0)
		return nil
	}
	COMMANDS["runjs"] = func(js []string) error {
		jss := strings.Join(js, "\n")
		v, err := vm.Run(jss)
		log.Println(v)
		return err
	}
	COMMANDS["oops!"] = func(_ []string) error {
		commandBuffer = nil
		return nil
	}
	COMMANDS["oops"] = func(cb []string) error {
		commandBuffer = cb[:len(cb)-1]
		return nil
	}
	COMMANDS["cb"] = func(cb []string) error {
		cbs := strings.Join(cb, "\n")
		fmt.Println(cbs)
		return nil
	}
	COMMANDS["log on"] = func(_ []string) error {
		log = logpkg.New(os.Stderr, "Log: ", 0)
		return nil
	}
	COMMANDS["log off"] = func(msg []string) error {
		msgs := strings.Join(msg, "")
		log = logpkg.New(ioutil.Discard, msgs, 0)
		return nil
	}
	COMMANDS["commands"] = func(_ []string) error {
		fmt.Print("Available commands: ")
		var ks []string
		for k := range COMMANDS {
			ks = append(ks, k)
		}
		fmt.Println(strings.Join(ks, ", "))
		return nil
	}


	os.Stderr.WriteString("Welcome to " + os.Args[0] + ", a modern line editor with support for extension via JavaScript.\nType 'commands' to see the available commands on your system at any moment.\n")

	for {
		command, _ := reader.ReadString('\n')
		command = strings.TrimRight(command, "\n\r")
		run, ok := COMMANDS[command]
		if len(command) > 0 && command[0] == '.' {
			command = command[1:]
		}
		if ok {
			log.Println("Enter native command exec")
			err := run(commandBuffer)
			if err != nil {
				os.Stderr.WriteString(err.Error()+"\n")
			} else {
				os.Stderr.WriteString("!\n")
			}
		} else {
			commandBuffer = append(commandBuffer, command)
			log.Printf("%#v", command)
		}
	}
}
