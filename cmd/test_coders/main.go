package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/go-chess/internal/coders"
	"example.com/go-chess/internal/reporter"
)

type BenchResult struct {
	Duration time.Duration
}

func main() {
	// Configure available coders here.
	coders := []coders.Coder{
		coders.NewDummyCoder(),
		coders.NewFenValidator(),
		coders.NewBinaryFenCoder(),
		coders.NewBinaryFenCompactCoder(),
		coders.NewHuffmanCoder(),
		coders.NewHuffmanCompactCoder(),
		coders.NewArithmeticCoder(),
		coders.NewArithmeticByCountCoder(),
		coders.NewArithmeticAdaptiveCoder(),
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
			coded := coder.FromFen(fen)
			recoded := coder.ToFen(coded)
			if fen != recoded {
				fmt.Printf("Encoded to %s\n", coded.String())
				log.Fatalf("coding with %s coder failed\n  expected: '%s'\n       got: '%s'", coder.Name(), fen, recoded)
			}
			reporters[idx].Log(fen, coded.Length)
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

	fmt.Printf(" %-"+ws+"s | %6s | %6s | %6s | %6s | %8s | %s\n", "Coder", "Min", "Mean", "Median", "Max", "Duration", "Worst case")
	fmt.Println(strings.Repeat("-", w+6*4+8+10+7*3-1))
	for idx, coder := range coders {
		reporter := reporters[idx]
		benchmark := benchmarks[idx]
		fmt.Printf(" %-"+ws+"s | %6d | %6.2f | %6d | %6d | %8.3f | %s\n", coder.Name(), reporter.Min, reporter.Mean(), reporter.Median(), reporter.Max, benchmark.Duration.Seconds(), reporter.WorstCase)
	}
}
