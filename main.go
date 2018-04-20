package main

import "bufio"
import "os"
import logpkg "log"
import "github.com/robertkrimen/otto"

var COMMANDS = make(map[string]func(string) error)
var vm = otto.New()
var log = logpkg.New(os.Stderr, "Logging is on: ", 0)

func addCommand(call otto.FunctionCall) otto.Value {
	log.Println("Enter addCommand")
	t, _ := vm.ToValue(true)
	f, _ := vm.ToValue(false)
	command, err := call.Argument(0).ToString()
	if err != nil {
		return f
	}
	log.Println("Adding command ", command)
	function := call.Argument(1)
	if err != nil {
		return f
	}
	COMMANDS[command] = func(e string) error {
		_, err := function.Call(f,e)
		if err != nil {
			return err
		}
		return nil
	}
	log.Printf("%#v", COMMANDS)
	return t
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	vm.Set("def", addCommand)

	for {
		command, _ := reader.ReadString('\n')
		vm.Set("command", command[:len(command)-1])
		run, ok := COMMANDS[command[0:1]]
		if !ok {
			log.Println("Enter JS exec")
			v,err := vm.Run(command)
			if err != nil {
				log.Println(err)
			} else {
				log.Println(v)
			}
		} else {
			log.Println("Enter native command exec")
			err := run(command)
			if err != nil {
				os.Stderr.WriteString(err.Error())
			}
		}
	}
}
