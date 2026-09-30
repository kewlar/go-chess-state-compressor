package main

import (
	"fmt"
	"log"
	"os"

	"github.com/corentings/chess/v2"
	"github.com/klauspost/compress/zstd"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-file.pgn.zst>", os.Args[0])
	}
	filePath := os.Args[1]

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	// Initialize streaming zstd reader
	zstdReader, err := zstd.NewReader(f)
	if err != nil {
		log.Fatalf("Error creating zstd reader: %v", err)
	}
	defer zstdReader.Close()

	// Initialize the streaming PGN scanner from v2
	scanner := chess.NewScanner(zstdReader)

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
