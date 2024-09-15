package src

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var words []string
var wordsMap map[int][]string

// startSession creates a new Wordle game session and returns a JWT session_id
func startSession(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		WordLength int `json:"wordLength"`
	}

	// Decode the request body to get the word length
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a random word for the session
	word := randomWord(wordsMap[req.WordLength])

	user, err := userId()
	// Create a new session with the word and empty guesses
	session := &Session{
		Word:       word,
		InProgress: true,
		Guesses:    []Guess{},
		UserId:     user,
	}

	// Encode the session into a JWT session_id
	sessionID, err := encodeSession(session)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Return the session_id (JWT token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID,
	})
}

// getSession retrieves the current session state from the JWT session_id (from request body)
func getSession(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		SessionID string `json:"session_id"`
	}

	// Parse the request body to extract session_id
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode the session from the session_id (JWT)
	session, err := decodeSession(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid session_id", http.StatusBadRequest)
		return
	}

	// Create a PublicSession object (omit the target word)
	publicSession := &PublicSession{
		SessionID: req.SessionID,
		Guesses:   session.Guesses,
	}

	// Return the public session (without the target word)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicSession)
}

// guessWordHandler handles the word guess and updates the session
func guessWordHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Word      string `json:"word"`
		SessionID string `json:"session_id"`
	}

	// Decode the request body to extract the guessed word and session_id
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Decode the session
	session, err := decodeSession(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid session_id", http.StatusBadRequest)
		return
	}

	if !session.InProgress {
		http.Error(w, "Session is not in progress", http.StatusBadRequest)
		return
	}

	// Process the guess
	if err := guessWord(session, req.Word); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = incrementTokenVersion(session.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Encode the updated session back into a new session_id (JWT token)
	newSessionID, err := encodeSession(session)
	if err != nil {
		http.Error(w, "Failed to encode session", http.StatusInternalServerError)
		return
	}

	// Create a PublicSession object (omit the target word)
	publicSession := &PublicSession{
		SessionID: newSessionID,
		Guesses:   session.Guesses,
	}

	// Return the public session (with the updated session_id and guesses)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publicSession)
}

func giveUpHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		SessionID string `json:"session_id"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	session, err := decodeSession(req.SessionID)
	if err != nil {
		http.Error(w, "Invalid session_id", http.StatusBadRequest)
		return
	}

	if session.InProgress {
		session.InProgress = false
		err = incrementTokenVersion(session.UserId)

		if err != nil {
			http.Error(w, "Failed to increment token version", http.StatusInternalServerError)
		}
	}

	type Response struct {
		SessionID string `json:"session_id"`
		Word      string `json:"word"`
	}

	newSessionID, err := encodeSession(session)
	if err != nil {
		http.Error(w, "Failed to encode session", http.StatusInternalServerError)
	}

	response := &Response{
		SessionID: newSessionID,
		Word:      session.Word,
	}

	// Return the public session (with the updated session_id and guesses)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterWordleRoutes registers the Wordle game routes
func RegisterWordleRoutes(r *mux.Router) {

	err := loadWords("./wordle/words.txt", &words)
	if err != nil {
		log.Fatal("Failed to load words:", err)
	}

	wordsMap = make(map[int][]string)
	for _, word := range words {
		wordsMap[len(word)] = append(wordsMap[len(word)], word)
	}

	// Register the Wordle game routes
	r.HandleFunc("/start", startSession).Methods("POST")
	r.HandleFunc("/session", getSession).Methods("POST")
	r.HandleFunc("/guess", guessWordHandler).Methods("POST")
	r.HandleFunc("/give_up", giveUpHandler).Methods("POST")
}
