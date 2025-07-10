package main

import (
	"fmt"
	"io/ioutil"
	logpkg "log"
	"os"
	"strconv"
	"strings"
)

func initCommands() {
	COMMANDS["quit"] = func(_ []string) error {
		os.Exit(0)
		return nil
	}
	COMMANDS["q"] = func(_ []string) error {
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
		if len(commandBuffer) > 0 {
			commandBuffer = commandBuffer[:len(commandBuffer)-1]
		}
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
		return nil
	}
	COMMANDS["d"] = func(_ []string) error {
		FB.Contents = perLine(FB.Contents)
		if len(FB.Contents) == 0 {
			return nil
		}
		if Line < 0 {
			Line = 0
		}
		if Line < len(FB.Contents) {
			log.Println(Line, Line-1, Line+1)
			before := FB.Contents[:Line]
			after := FB.Contents[Line+1:]
			FB.Contents = append([]string(nil), before...)
			FB.Contents = append(FB.Contents, after...)
		} else if len(FB.Contents) != 0 {
			FB.Contents = FB.Contents[:len(FB.Contents)-1]
		}
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
		return nil
	}
	COMMANDS["save"] = func(path []string) error {
		filename := strings.Join(path, string(os.PathSeparator))
		return saveFullFile(filename)
	}
	
	// Enhanced ed commands
	COMMANDS["c"] = func(cb []string) error {
		rangeStr := "."
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		newLines := cb[1:]
		return executeChange(rangeStr, newLines)
	}
	COMMANDS["m"] = func(cb []string) error {
		if len(cb) < 2 {
			return fmt.Errorf("move command requires source range and destination")
		}
		return executeMove(cb[0], cb[1])
	}
	COMMANDS["t"] = func(cb []string) error {
		if len(cb) < 2 {
			return fmt.Errorf("copy command requires source range and destination")
		}
		return executeCopy(cb[0], cb[1])
	}
	COMMANDS["j"] = func(cb []string) error {
		rangeStr := ".,+1"
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		return executeJoin(rangeStr)
	}
	COMMANDS["s"] = func(cb []string) error {
		if len(cb) == 0 {
			return fmt.Errorf("substitute command requires pattern")
		}
		return executeSubstitute(strings.Join(cb, " "))
	}
	COMMANDS["g"] = func(cb []string) error {
		if len(cb) == 0 {
			return fmt.Errorf("global command requires pattern")
		}
		return executeGlobal(strings.Join(cb, " "))
	}
	COMMANDS["k"] = func(cb []string) error {
		if len(cb) == 0 {
			return fmt.Errorf("mark command requires mark character")
		}
		return executeMark(Line, cb[0])
	}
	COMMANDS["u"] = func(cb []string) error {
		return executeUndo()
	}
	COMMANDS["y"] = func(cb []string) error {
		rangeStr := "."
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		return executeYank(rangeStr)
	}
	COMMANDS["x"] = func(cb []string) error {
		address := "."
		if len(cb) > 0 && cb[0] != "" {
			address = cb[0]
		}
		return executePut(address)
	}
	COMMANDS["w"] = func(cb []string) error {
		rangeStr := "%"
		filename := ""
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		if len(cb) > 1 && cb[1] != "" {
			filename = cb[1]
		}
		return executeWrite(rangeStr, filename)
	}
	COMMANDS["W"] = func(cb []string) error {
		rangeStr := "%"
		filename := ""
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		if len(cb) > 1 && cb[1] != "" {
			filename = cb[1]
		}
		return executeWriteAppend(rangeStr, filename)
	}
	COMMANDS["e"] = func(cb []string) error {
		filename := ""
		if len(cb) > 0 && cb[0] != "" {
			filename = cb[0]
		}
		command := "e"
		if filename != "" {
			command = "e " + filename
		}
		return executeEdit(command)
	}
	COMMANDS["E"] = func(cb []string) error {
		filename := ""
		if len(cb) > 0 && cb[0] != "" {
			filename = cb[0]
		}
		command := "E"
		if filename != "" {
			command = "E " + filename
		}
		return executeEditUnconditional(command)
	}
	COMMANDS["R"] = func(cb []string) error {
		address := "$"
		filename := ""
		if len(cb) > 0 && cb[0] != "" {
			filename = cb[0]
		}
		return executeRead(address, filename)
	}
	COMMANDS["l"] = func(cb []string) error {
		rangeStr := "."
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		output, err := executeList(rangeStr)
		if err != nil {
			return err
		}
		for _, line := range output {
			fmt.Println(line)
		}
		return nil
	}
	COMMANDS["n"] = func(cb []string) error {
		rangeStr := "."
		if len(cb) > 0 && cb[0] != "" {
			rangeStr = cb[0]
		}
		output, err := executeNumber(rangeStr)
		if err != nil {
			return err
		}
		for _, line := range output {
			fmt.Println(line)
		}
		return nil
	}
	COMMANDS["="] = func(cb []string) error {
		address := "."
		if len(cb) > 0 && cb[0] != "" {
			address = cb[0]
		}
		num, err := executeEquals(address)
		if err != nil {
			return err
		}
		fmt.Println(num)
		return nil
	}
	COMMANDS["!"] = func(cb []string) error {
		if len(cb) == 0 {
			return fmt.Errorf("shell command required")
		}
		command := strings.Join(cb, " ")
		output, err := executeShellCommand(command)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	}
}
