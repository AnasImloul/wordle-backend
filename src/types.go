package src

type Guess struct {
	Word    string          `json:"word"`
	Letters []GuessedLetter `json:"letters"`
}

type GuessedLetter struct {
	Letter string `json:"letter"`
	Match  string `json:"match"`
}

type Session struct {
	UserId     string  `json:"-"`       // used to track JWT version
	Word       string  `json:"-"`       // The target word (not returned to the client)
	InProgress bool    `json:"-"`       // used to track if player gave up
	Guesses    []Guess `json:"guesses"` // List of guesses
}

type PublicSession struct {
	SessionID string  `json:"session_id"` // The JWT token representing the session
	Guesses   []Guess `json:"guesses"`    // List of guesses (no target word)
}
