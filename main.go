package main

import "bufio"
import "os"
import "github.com/robertkrimen/otto"

func main() {
	vm := otto.New()
	reader := bufio.NewReader(os.Stdin)

	for {
		command, _ := reader.ReadString('\n')
		vm.Set("command", command)
		vm.Run(`
		console.log(text)
		`)
	}
}
