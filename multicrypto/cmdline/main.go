// Package main builds a command line version of the Multi-Crypto math puzzle.
package main

import (
	"fmt"
	"github.com/wblakecaldwell/mathpuzzles/multicrypto"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Need a word or phrase!")
		os.Exit(1)
	}
	phrase := os.Args[1]

	// working on the standard 12x12 "times table", but without 1x's, since that's too easy!
	puzzleGenerator, err := multicrypto.NewPuzzleGenerator(2, 12, multicrypto.DecoderRandom())
	if err != nil {
		fmt.Println("Oops! Something went wrong building the Puzzle Generator!")
		os.Exit(1)
	}

	key, err := puzzleGenerator.GenerateDecoderKey()
	if err != nil {
		fmt.Println("Oops! Something went wrong generating the decoder key!")
		os.Exit(1)
	}
	puzzle, err := puzzleGenerator.GeneratePuzzle(phrase)
	if err != nil {
		fmt.Println("Oops! Something went wrong generating the puzzle!")
		os.Exit(1)
	}

	fmt.Println("Decoder Key\n-----------\n")
	for _, c := range key {
		fmt.Printf("%s: %s = ______\n", c.Letter, c.Clue)
	}

	fmt.Println("\n\n")
	fmt.Println("Secret Message\n--------------\n")
	for _, c := range puzzle {
		if c.IsMathProblem() {
			fmt.Printf("%s = ______\n", c.String())
		} else {
			fmt.Printf("%s\n", c.String())
		}
	}
}
