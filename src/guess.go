package src

import (
	"errors"
	"strings"
)

func guessWord(session *Session, word string) error {
	// Convert both the guessed word and the target word to uppercase
	word = strings.ToUpper(word)
	targetWord := strings.ToUpper(session.Word)

	if len(word) != len(targetWord) {
		return errors.New("word length mismatch")
	}

	// Calculate match (INCORRECT, MISPLACED, CORRECT)
	match := make([]string, len(word))
	targetRunes := []rune(targetWord)
	guessRunes := []rune(word)

	letterCount := make(map[rune]int)
	for i := 0; i < len(guessRunes); i++ {
		letterCount[targetRunes[i]]++
	}

	// Mark letters as CORRECT, MISPLACED, or INCORRECT
	for i := 0; i < len(guessRunes); i++ {
		if guessRunes[i] == targetRunes[i] {
			match[i] = "CORRECT"
			letterCount[guessRunes[i]]--
		} else if letterCount[guessRunes[i]] > 0 {
			match[i] = "MISPLACED"
			letterCount[guessRunes[i]]--
		} else {
			match[i] = "INCORRECT"
		}
	}

	guess := Guess{Word: word}
	for i := 0; i < len(match); i++ {
		guess.Letters = append(guess.Letters, GuessedLetter{Letter: string(guessRunes[i]), Match: match[i]})
	}

	// Record the guess
	session.Guesses = append(session.Guesses, guess)

	return nil
}
