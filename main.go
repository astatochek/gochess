package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Welcome to GoChess!")
	fmt.Println("Type 'help' for commands")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "quit" || input == "exit" {
			break
		}

		if input == "help" {
			printHelp()
			continue
		}

		fmt.Println(input)
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  <from> <to> - make a move (e.g., 'e2 e4')")
	fmt.Println("  help - show this help")
	fmt.Println("  quit/exit - quit the game")
}
