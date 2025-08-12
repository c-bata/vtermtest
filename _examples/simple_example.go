package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	fmt.Printf("Hello, %s!\n", name)
	
	fmt.Print("Enter a number: ")
	num, _ := reader.ReadString('\n')
	num = strings.TrimSpace(num)
	fmt.Printf("You entered: %s\n", num)
}