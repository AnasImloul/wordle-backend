package src

import (
	"bufio"
	"github.com/google/uuid"
	"math/rand"
	"os"
)

// loadWords reads the words from "words.txt" and returns them as a slice of strings
func loadWords(filePath string, words *[]string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		*words = append(*words, scanner.Text())
	}

	// Check for any scanning error
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// randomWord selects a random word from the list of words loaded from "words.txt"
func randomWord(words []string) string {
	// Choose a random word from the list
	return words[rand.Intn(len(words))]
}

func userId() (string, error) {
	// Generate a new UUID
	newUUID := uuid.New()

	// Return the UUID as a string
	return newUUID.String(), nil
}
