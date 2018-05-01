package main

import "bufio"
import "os"
import logpkg "log"
import "github.com/robertkrimen/otto"
import "strings"
import "fmt"
import "io/ioutil"
import "io"
import "strconv"

var COMMANDS = make(map[string]func([]string) error) // Map of native commands.
// JS can add to COMMANDS with def(command, function)
var vm = otto.New()                         // JS environment
var log = logpkg.New(ioutil.Discard, "", 0) // Dummy logger (logging turned off)
var commandBuffer []string                  // When something doesn't read as a command, add to buffer
var Line int

type FileBuffer struct {
	Pos      int
	Length   int
	File     io.ReadWriter
	Contents []string
	Meta     map[string]string
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
		_, err := function.Call(f, cbs)
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
	FB = FileBuffer{0, 0, nil, nil, map[string]string{"filename": filename}}
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
	FB = FileBuffer{0, l, file, []string{string(bytes)}, m}
	return nil
}

func saveFullFile(filename string) error {
	log.Printf("Metadata: %#v\n", FB.Meta)
	if filename == "" {
		file, ok := FB.Meta["filename"]
		if !ok {
			return fmt.Errorf("no filename available")
		}
		filename = file
	}
	perm := os.ModePerm
	if file, ok := FB.File.(*os.File); ok {
		fi, err := file.Stat()
		if err != nil {
			perm = fi.Mode()
		}
	}
	return ioutil.WriteFile(filename,
		[]byte(strings.Join(FB.Contents, "\n")+"\n"),
		perm)
}

func printLines(from, many int, page bool) {
	log.Printf("Running printLines with %v, %v and %v\n", from, many, page)
	mod := many // Page per mod lines
	var readall bool
	if many == 0 {
		mod = 1
		readall = true
	}
	if page {
		readall = true
	}
	to := from + many - 1
	FB.Contents = perLine(FB.Contents)
	for i := from; i < len(FB.Contents); i++ {
		log.Println("Value of i: ", i)
		if !readall && i > to {
			return
		}
		fmt.Println(FB.Contents[i])
		Line = i
		if page && (i+1)%mod == 0 {
			if s, _ := bufio.NewReader(os.Stdin).ReadString('\n'); s[:1] == "q" {
				return
			}
		}
	}
}

func simpleSearch(keyword string, many int, page bool) {
	log.Printf("Running printLines with %v, %v and %v\n", keyword, many, page)
	mod := many // Page per mod lines
	from := Line
	var readall bool
	if many == 0 {
		mod = 1
		readall = true
	}
	if page {
		readall = true
	}
	FB.Contents = perLine(FB.Contents)
	var q int
	if many < 0 {
		many = -many
		from = 0
	}
	for i, line := range FB.Contents {
		log.Println("Value of i: ", i)
		if !strings.Contains(line, keyword) || i < from {
			continue
		}
		q++
		if !readall && q > many {
			return
		}
		fmt.Println(i, ": ", line)
		Line = i
		if page && q%mod == 0 {
			if s, _ := bufio.NewReader(os.Stdin).ReadString('\n'); s[:1] == "q" {
				return
			}
		}
	}
}

func perLine(in []string) (out []string) {
	for _, s := range in {
		temp := strings.Split(s, "\n")
		for _, temps := range temp {
			out = append(out, temps)
		}
	}
	return out
}

func NativeCommand(run func([]string) error, cb []string) {
	log.Println("Enter native command exec")
	err := run(cb)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
	} else {
		os.Stderr.WriteString("!\n")
	}
}

func main() {
	if len(os.Args) > 1 {
		openFullFile(os.Args[1])
	}

	vm.Set("def", addCommand)
	vm.Set("e", runEditorCommand)

	initCommands()

	os.Stderr.WriteString("Welcome to " + os.Args[0] + ", a modern line editor with support for extension via JavaScript.\nType 'commands' to see the available commands on your system at any moment.\n")

	for {
		vm.Set("linenumber", Line)
		reader := bufio.NewReader(os.Stdin)
		command, _ := reader.ReadString('\n')
		command = strings.TrimRight(command, "\n\r")
		commands := strings.Split(command, ":")
		cb := make([]string, len(commands)-1)
		run, ok := COMMANDS[command]
		if !ok && len(commands) > 1 {
			run, ok = COMMANDS[commands[0]]
			copy(cb,commands[1:])
			if commands[len(commands)-1] == "" {
				log.Println("Clear commandBuffer")
				commandBuffer = nil
			}
			log.Printf("cb is now: %#v\n", cb)
		} else {
			cb = commandBuffer
		}
		if len(command) > 0 {
			switch command[:1] {
			case ".":
				command = command[1:]
			case ":":
				Line, _ = strconv.Atoi(command[1:])
				continue
			}
		}
		if ok {
			log.Printf("2:cb is now: %#v\n", cb)
			NativeCommand(run, cb)
		} else {
			commandBuffer = append(commandBuffer, command)
			log.Printf("%#v", command)
		}
	}
}
