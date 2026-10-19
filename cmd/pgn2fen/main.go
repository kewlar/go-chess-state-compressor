package main

import (
	"fmt"
	"log"
	"os"

	"github.com/corentings/chess/v2"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-file.pgn>", os.Args[0])
	}
	filePath := os.Args[1]

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	// Initialize the streaming PGN scanner from v2
	scanner := chess.NewScanner(f)

	if !scanner.HasNext() {
		fmt.Println(scanner.LastError())
	}

	for scanner.HasNext() {
		game, err := scanner.ParseNext()
		if err != nil {
			log.Fatalf("Error scanning PGN: %v", err)
			continue
		}

		newGame := chess.NewGame()
		fmt.Println(newGame.Position().String())
		for _, move := range game.Moves() {
			err := newGame.Move(move, nil)
			if err != nil {
				panic(err)
			}
			fmt.Println(newGame.Position().String())
		}
	}
}
