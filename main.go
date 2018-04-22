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
			printLines(line, 1, false)
		case 2:
			from, _ := strconv.Atoi(pos[0])
			to, _ := strconv.Atoi(pos[1])
			printLines(from, to-from+1, false)
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
			printLines(0, pagesize, true)
		case 2:
			from, _ := strconv.Atoi(pos[0])
			pagesize, _ := strconv.Atoi(pos[1])
			printLines(from, pagesize, true)
		default:
			os.Stderr.WriteString("?")
			log.Println("commandbuffer too full")
		}
		commandBuffer = nil
		return nil
	}
	COMMANDS["r"] = func(line []string) error {
		lines := strings.Join(line, "\n")
		FB.Contents = perLine(FB.Contents)
		if Line < len(FB.Contents) {
			FB.Contents[Line] = lines
		} else {
			FB.Contents = append(FB.Contents, line...)
		}
		commandBuffer = nil
		return nil
	}
	COMMANDS["i"] = func(line []string) error {
		FB.Contents = perLine(FB.Contents)
		if Line < len(FB.Contents) {
			before := FB.Contents[:Line]
			after := FB.Contents[Line:]
			FB.Contents = append([]string(nil), before...)
			FB.Contents = append(FB.Contents, line...)
			FB.Contents = append(FB.Contents, after...)
		} else {
			FB.Contents = append(FB.Contents, line...)
		}
		commandBuffer = nil
		return nil
	}
	COMMANDS["a"] = func(line []string) error {
		FB.Contents = perLine(FB.Contents)
		if Line+1 < len(FB.Contents) {
			before := FB.Contents[:Line+1]
			after := FB.Contents[Line+1:]
			FB.Contents = append([]string(nil), before...)
			FB.Contents = append(FB.Contents, line...)
			FB.Contents = append(FB.Contents, after...)
		} else {
			FB.Contents = append(FB.Contents, line...)
		}
		commandBuffer = nil
		return nil
	}
	COMMANDS["d"] = func(_ []string) error {
		FB.Contents = perLine(FB.Contents)
		if Line+1 < len(FB.Contents) {
			log.Println(Line, Line-1, Line+1)
			before := FB.Contents[:Line]
			after := FB.Contents[Line+1:]
			FB.Contents = append([]string(nil), before...)
			FB.Contents = append(FB.Contents, after...)
		} else {
			FB.Contents = FB.Contents[:len(FB.Contents)-1]
		}
		commandBuffer = nil
		return nil
	}
	COMMANDS["search"] = func(contents []string) error {
		if len(contents) == 0 {
			return fmt.Errorf("keyword missing")
		}
		var n int
		if len(contents) > 1 {
			n, _ = strconv.Atoi(contents[1])
		}
		simpleSearch(contents[0], n, true)
		commandBuffer = nil
		return nil
	}
	COMMANDS["save"] = func(path []string) error {
		filename := strings.Join(path, string(os.PathSeparator))
		commandBuffer = nil
		return saveFullFile(filename)
	}

	os.Stderr.WriteString("Welcome to " + os.Args[0] + ", a modern line editor with support for extension via JavaScript.\nType 'commands' to see the available commands on your system at any moment.\n")

	for {
		vm.Set("linenumber", Line)
		reader := bufio.NewReader(os.Stdin)
		command, _ := reader.ReadString('\n')
		command = strings.TrimRight(command, "\n\r")
		run, ok := COMMANDS[command]
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
			log.Println("Enter native command exec")
			err := run(commandBuffer)
			if err != nil {
				os.Stderr.WriteString(err.Error() + "\n")
			} else {
				os.Stderr.WriteString("!\n")
			}
		} else {
			commandBuffer = append(commandBuffer, command)
			log.Printf("%#v", command)
		}
	}
}


