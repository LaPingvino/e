package main

import "bufio"
import "os"
import logpkg "log"
import "github.com/robertkrimen/otto"
import "strings"
import "fmt"
import "io/ioutil"
import "io"
import "bytes"
import "strconv"

var COMMANDS = make(map[string]func([]string) error) // Map of native commands.
// JS can add to COMMANDS with def(command, function)
var vm = otto.New() // JS environment
var log = logpkg.New(ioutil.Discard, "", 0) // Change ioutil.Discard to something else for logging
var commandBuffer []string // When something doesn't read as a command, add to buffer
var Line int

type FileBuffer struct{
	Pos int
	Length int
	File io.ReadWriter
	Bytes ByteChain
	Meta map[string]string
}

type ByteChain struct{
	Previous *ByteChain
	Next *ByteChain
	Bytes []byte
}

var FB FileBuffer

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

func openFullFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	l := len(bytes)
	m := make(map[string]string)
	m["filename"] = filename
	FB = FileBuffer{0, l, file, ByteChain{nil,nil,bytes}, m}
	return nil
}

func printLines(from, many int, page bool) {
	log.Printf("Running printLines with %v, %v and %v\n", from, many, page)
	i := 0
	defer func() {
		Line = i
	}()
	mod := many // Page per mod lines
	var readall bool
	if many == 0 {
		mod = 1
		readall = true
	}
	to := from + many - 1
	b := FB.Bytes
	for {
		bb := bufio.NewReader(bytes.NewReader(b.Bytes))
		for line, err := bb.ReadString('\n'); err == nil; line, err = bb.ReadString('\n') {
			i++
			log.Println("Value of i: ", i)
			if i < from {
				continue
			}
			fmt.Print(line)
			if !readall && !page && i >= to {
				return
			}
			if page && i % mod == 0 {
				if s, _ := bufio.NewReader(os.Stdin).ReadString('\n'); s[:1] == "q" {
					return
				}
			}
		}
		if b.Next != nil {
			b = *b.Next
		} else {
			return
		}
	}
}

func main() {
	if len(os.Args) > 1 {
		openFullFile(os.Args[1])
	}

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
		commandBuffer = nil
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
	COMMANDS["log off"] = func(_ []string) error {
		log = logpkg.New(ioutil.Discard, "", 0)
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
	COMMANDS["open"] = func(path []string) error {
		paths := strings.Join(path, string(os.PathSeparator))
		commandBuffer = nil
		return openFullFile(paths)
	}
	COMMANDS["print"] = func(pos []string) error {
		switch len(pos) {
		case 0:
			printLines(0, 0, false)
		case 1:
			line, _ := strconv.Atoi(pos[0])
			printLines(line,1,false)
		case 2:
			from, _ := strconv.Atoi(pos[0])
			to, _ := strconv.Atoi(pos[1])
			printLines(from,to-from+1,false)
		default:
			os.Stderr.WriteString("?")
			log.Println("commandbuffer too full")
		}
		commandBuffer = nil
		return nil
	}
	COMMANDS["page"] = func(pos []string) error {
		switch len(pos) {
		case 0:
			printLines(0, 1, true)
		case 1:
			pagesize, _ := strconv.Atoi(pos[0])
			printLines(0,pagesize,true)
		case 2:
			from, _ := strconv.Atoi(pos[0])
			pagesize, _ := strconv.Atoi(pos[1])
			printLines(from,pagesize,true)
		default:
			os.Stderr.WriteString("?")
			log.Println("commandbuffer too full")
		}
		commandBuffer = nil
		return nil
	}

	os.Stderr.WriteString("Welcome to " + os.Args[0] + ", a modern line editor with support for extension via JavaScript.\nType 'commands' to see the available commands on your system at any moment.\n")

	for {
		vm.Set("linenumber", Line)
		reader := bufio.NewReader(os.Stdin)
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
