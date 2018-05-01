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
		if Line+1 < len(FB.Contents) {
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
}
