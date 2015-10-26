// Package multicrypto contains the logic for a math puzzle generator that,
// given some input text, creates a math problem per letter, in the form:
//   (3 x 5) - (2 x 3) = _______
// where the answer gives a number between 1-26, mapping to a letter in the
// 26-character decoder key. In this case, given a decoder key of
// "klcnogdwprftyxqismjvehabzu", `15 - 6 = 9`, so this is a clue for "p",
// the ninth letter of the decoder key.
package multicrypto

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// DecoderAlphabetic returns a standard A=1, B=2 decoder scheme
func DecoderAlphabetic() string {
	return "abcdefghijklmnopqrstuvwxyz"
}

// DecoderRandom returns a random decoder key
func DecoderRandom() string {
	alpha := DecoderAlphabetic()
	result := make([]uint8, 26)
	for i := 0; i < 26; i++ {
		result[i] = alpha[i]
	}

	// shuffle by swapping letters 100 times
	var tmp uint8
	var a, b int
	for i := 0; i < 100; i++ {
		a = rand.Intn(26)
		b = rand.Intn(26)
		tmp = result[a]
		result[a] = result[b]
		result[b] = tmp
	}
	return string(result)
}

// operation is a mathematical operation
type operation struct {
	a int
	b int
}

// multiplicationOperation represents two numbers being multiplied together (a x b)
type multiplicationOperation struct {
	operation
}

// subtractionOperation represents the subtraction of one integer from another (a - b)
type subtractionOperation struct {
	operation
}

// DecoderKeyCharacter represents a math problem for a specific letter of the alphabet
type DecoderKeyCharacter struct {
	Letter string
	Clue   string
}

// PuzzleCharacter represents a character in the puzzle, either a math problem
// in the form '(a x b) - (c x d) = ________' or
type PuzzleCharacter struct {
	a           int    // value for a in '(a x b) - (c x d)'
	b           int    // value for b in '(a x b) - (c x d)'
	c           int    // value for c in '(a x b) - (c x d)'
	d           int    // value for d in '(a x b) - (c x d)'
	literalText string // if not empty, this character is fully represented by this
}

// IsMathProblem returns whether this character represents a math problem
func (pc *PuzzleCharacter) IsMathProblem() bool {
	return pc.literalText == ""
}

// String returns a formatted version of this puzzle character.
func (pc *PuzzleCharacter) String() string {
	if len(pc.literalText) > 0 {
		return pc.literalText
	}
	return fmt.Sprintf("(%d x %d) - (%d x %d)", pc.a, pc.b, pc.c, pc.d)
}

// PuzzleGenerator generates puzzles from phrases, where each character
// is represented as a math problem in the form `(a x b) - (c x d)`, and
// the result is a number between 1-26, corresponding to a letter in the
// alphabet with that index.
type PuzzleGenerator struct {
	minMultiDigit         int                               // the minimum digit that can be in a multiplication problem
	maxMultiDigit         int                               // the maximum digit that can be in a multiplication problem
	availableProducts     map[int][]multiplicationOperation // available products to use on each side of the subtraction
	availableSubtractions [][]subtractionOperation          // available subtractions that generate each letter (index 0-25)
	decoder               string                            // 26-character decoder ring
}

// NewPuzzleGenerator returns a new *PuzzleGenerator with:
// - minMultiDigit: the minimum number to use as a multiplication factor
// - maxMultiDigit: the maximum number to use as a multiplication factor
// - decoder: a 26-character string with each letter representing a number by its 0-based index,
//   so if it's "abcdefg...", then A=1, B=2. If it's "bdagq...", then B=1, D=2, etc. Consider
//   using DecoderAlphabetic() and DecoderRandom(). The DecoderKey() method will return
//   clues to solve for the decoder key.
func NewPuzzleGenerator(minMultiDigit int, maxMultiDigit int, decoder string) (*PuzzleGenerator, error) {
	if len(decoder) != 26 {
		return nil, fmt.Errorf("The decoder must be 26 characters")
	}
	// TODO: check to make sure each letter is represented once

	products := calculateAvailableProducts(minMultiDigit, maxMultiDigit)
	return &PuzzleGenerator{
		minMultiDigit:         minMultiDigit,
		maxMultiDigit:         maxMultiDigit,
		decoder:               decoder,
		availableProducts:     products,
		availableSubtractions: calculateSubtractions(products),
	}, nil
}

// GenerateDecoderKey returns a decoder key for this puzzle generator with random
// equations representing the characters each time it's called. The return values
// are in alphabetical order, so the first equation returned represents the value
// for A, the second for B, etc.
func (pg *PuzzleGenerator) GenerateDecoderKey() ([]DecoderKeyCharacter, error) {
	result := make([]DecoderKeyCharacter, 26)
	alpha := DecoderAlphabetic() // for the ordering of the output
	for pos, c := range alpha {
		index := strings.IndexRune(pg.decoder, c)
		puzzleChar := pg.puzzleCharacterForIndex(index)

		result[pos] = DecoderKeyCharacter{
			Letter: strings.ToUpper(string(c)),
			Clue:   puzzleChar.String()}
	}
	return result, nil
}

// GeneratePuzzle builds a puzzle for the input phrase, converting each letter to a math
// problem that looks like "(a x b) - (c x d) = __________"
func (pg *PuzzleGenerator) GeneratePuzzle(phrase string) ([]PuzzleCharacter, error) {
	lcPhrase := strings.ToLower(phrase)
	var result []PuzzleCharacter
	for _, c := range lcPhrase {
		alphaNum := strings.IndexRune(pg.decoder, c)
		if alphaNum < 0 {
			// something other than a letter - just pass it through
			result = append(result, PuzzleCharacter{literalText: string(c)})
		} else {
			result = append(result, pg.puzzleCharacterForIndex(alphaNum))
		}
	}
	return result, nil
}

// puzzleCharacterForIndex returns a random PuzzleCharacter for the input index
// into the PuzzleGenerator's decoder.
func (pg *PuzzleGenerator) puzzleCharacterForIndex(index int) PuzzleCharacter {
	pc := PuzzleCharacter{}
	var randIndex int

	// find a random subtraction for this letter
	randIndex = rand.Intn(len(pg.availableSubtractions[index]))
	subtraction := pg.availableSubtractions[index][randIndex]

	// find random multiplication operation for left side of the subtraction
	randIndex = rand.Intn(len(pg.availableProducts[subtraction.a]))
	pc.a = pg.availableProducts[subtraction.a][randIndex].a
	pc.b = pg.availableProducts[subtraction.a][randIndex].b

	// find random multiplication operation for right side of the subtraction
	randIndex = rand.Intn(len(pg.availableProducts[subtraction.b]))
	pc.c = pg.availableProducts[subtraction.b][randIndex].a
	pc.d = pg.availableProducts[subtraction.b][randIndex].b

	return pc
}

// calculateAvailableProducts calculates all possible multiplication
// products for two numbers between the input `min` and `max` values,
// along with all possible ways to get them.
func calculateAvailableProducts(min int, max int) map[int][]multiplicationOperation {
	// figure out what products are possible
	possibleProducts := make(map[int][]multiplicationOperation)
	var ixj int
	for i := min; i <= max; i++ {
		for j := min; j <= max; j++ {
			ixj = i * j
			possibleProducts[ixj] = append(possibleProducts[ixj], multiplicationOperation{operation{a: i, b: j}})
		}
	}
	return possibleProducts
}

// calculateSubtractions returns a slice of 26 elements, one representing each lowercase letter,
// with each being a slice of possible subtraction operations that equal the index.
// For example,
func calculateSubtractions(products map[int][]multiplicationOperation) [][]subtractionOperation {
	result := make([][]subtractionOperation, 26)
	for alphaNum := 0; alphaNum < 26; alphaNum++ {
		for i := range products {
			for j := range products {
				if i-j == alphaNum+1 { // need the +1 because the math needs to reflect [1, 26], not [0,25]
					result[alphaNum] = append(result[alphaNum], subtractionOperation{operation{a: i, b: j}})
				}
			}
		}
	}
	return result
}
