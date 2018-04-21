package main

import "bufio"
import "os"
import logpkg "log"
import "github.com/robertkrimen/otto"
import "strings"

var COMMANDS = make(map[string]func([]string) error) // Map of native commands.
// JS can add to COMMANDS with def(command, function)
var vm = otto.New() // JS environment
var log = logpkg.New(os.Stderr, "Logging is on: ", 0) // During development this is on
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
		jss := strings.Join(js[:len(js)-1], "\n")
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

	for {
		command, _ := reader.ReadString('\n')
		command = strings.TrimRight(command, "\n\r")
		run, ok := COMMANDS[command]
		if len(command) > 0 && command[0] == '.' {
			command = command[1:]
		}
		commandBuffer = append(commandBuffer, command)
		if ok {
			log.Println("Enter native command exec")
			err := run(commandBuffer)
			if err != nil {
				os.Stderr.WriteString(err.Error())
			}
		} else {
			log.Printf("%#v", command)
		}
	}
}
