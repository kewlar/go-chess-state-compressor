package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/go-chess/internal/chess"
	"example.com/go-chess/internal/coders"
	"example.com/go-chess/internal/reporter"
)

type BenchResult struct {
	Duration time.Duration
}

func main() {
	// Configure available coders here.
	coders := []coders.Coder{
		// Baseline - encode state to FEN.
		coders.NewFenCoder(),
		// Raw metadata; only piece encoding differs.
		coders.NewBinaryFenCoder(),
		coders.NewHuffmanCoder(),
		coders.NewHuffmanAdaptiveCoder(),
		coders.NewArithmeticCoder(),
		coders.NewArithmeticByCountCoder(),
		coders.NewArithmeticAdaptiveCoder(),
		// Various tricks being used to compress the data.
		coders.NewBinaryFenCompactCoder(),
		coders.NewHuffmanCompactCoder(),
		coders.NewHuffmanAdaptiveCompactCoder(),
		coders.NewArithmeticAdaptiveCompactCoder(),
		coders.NewArithmeticAdaptiveCompact2Coder(),
		coders.NewHuffmanAdaptiveCompact2Coder(),
	}

	// Prepare to read input data.
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-file.txt>", os.Args[0])
	}
	filePath := os.Args[1]

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	// Set up statistic and benchmark gatherers.
	reporters := []*reporter.Reporter{}
	for range coders {
		reporters = append(reporters, reporter.NewReporter())
	}
	benchmarks := make([]BenchResult, len(coders))

	for idx, coder := range coders {
		fmt.Printf("Testing %s coder...\n", coder.Name())

		offset, err := f.Seek(0, 0)
		if err != nil || offset != 0 {
			log.Fatalf("Error rewinding file: %v", err)
		}

		lines := 0
		scanner := bufio.NewScanner(f)
		start := time.Now()
		for scanner.Scan() {
			if lines&1023 == 0 {
				fmt.Printf("%d\r", lines)
			}
			fen := scanner.Text()
			parts := strings.Split(fen, "#")
			fen = strings.TrimSpace(parts[0])
			if len(fen) == 0 {
				continue
			}
			state, err := chess.ParseFen(fen)
			if err != nil {
				log.Fatalf("An error occurred while parsing FEN \"%s\": %v", fen, err)
			}
			bits := coder.Encode(state)
			recoded := coder.Decode(bits).ToFen()
			if fen != recoded {
				fmt.Printf("Encoded to %s\n", bits.String())
				log.Fatalf("coding with %s coder failed\n  expected: '%s'\n       got: '%s'", coder.Name(), fen, recoded)
			}
			reporters[idx].Log(fen, bits.Length)
			lines++
		}
		benchmarks[idx].Duration = time.Since(start)

		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading file: %v", err)
		}

		fmt.Printf("%d lines read.\n", lines)
	}

	// Report statistics.
	w := 0
	for _, coder := range coders {
		w = max(w, len(coder.Name()))
	}
	ws := strconv.Itoa(w)

	fmt.Printf("| %-"+ws+"s | %9s | %9s | %9s | %9s | %11s | %s |\n", "Coder", "Min, b", "Mean, b", "Median, b", "Max, b", "Duration, s", "Worst case FEN")
	fmt.Println("| :" + strings.Repeat("-", w-1) + strings.Repeat(" | --------:", 4) + " | ----------: | :------------- |")
	for idx, coder := range coders {
		reporter := reporters[idx]
		benchmark := benchmarks[idx]
		fmt.Printf("| %-"+ws+"s | %9d | %9.5f | %9d | %9d | %11.3f | %s |\n", coder.Name(), reporter.Min, reporter.Mean(), reporter.Median(), reporter.Max, benchmark.Duration.Seconds(), reporter.WorstCase)
	}
}
