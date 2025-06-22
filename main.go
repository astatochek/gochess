package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/notnil/chess"
)

// TODO: this works only on linux, change later
func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	clear()

	game := chess.NewGame()
	scanner := bufio.NewScanner(os.Stdin)

	for game.Outcome() == chess.NoOutcome {

		fmt.Println(game.Position().Board().Draw())

		if game.Position().Turn() == chess.White {
			fmt.Print("White>")
		} else {
			fmt.Print("Black>")
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		clear()

		if err := game.MoveStr(input); err != nil {
			fmt.Println("Invalid move:", err)
			continue
		}
	}

	// fmt.Println(game.Position().Board().Draw())
	fmt.Println("Game over. Result:", game.Outcome(), "Method:", game.Method())
}
