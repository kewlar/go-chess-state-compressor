package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/corentings/chess/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Converts PGN file into a series of FENs.")
		fmt.Println("If filename is \"-\", reads from standard input.")
		fmt.Println()
		fmt.Printf("Usage: %s <path-to-file.pgn>", os.Args[0])
		os.Exit(0)
	}

	filePath := os.Args[1]
	var reader io.Reader
	if filePath == "-" { // Standard input
		reader = os.Stdin
	} else { // Regular file
		f, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer f.Close()
		reader = f
	}

	// Initialize the streaming PGN scanner from v2
	scanner := chess.NewScanner(reader, chess.WithScannerBufferSize(1024*1024))

	if !scanner.HasNext() {
		fmt.Println(scanner.LastError())
	}

	for scanner.HasNext() {
		game, err := scanner.ParseNext()
		if err != nil {
			log.Fatalf("Error scanning PGN: %v", err)
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
