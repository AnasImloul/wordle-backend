package src

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

// JWT Secret for signing session tokens
var jwtSecret = []byte("my_secret_key")

func getTokenVersion(userId string) (int, error) {
	val, err := rdb.Get(ctx, fmt.Sprintf("session-%s", userId)).Result()
	if errors.Is(err, redis.Nil) {
		// Key does not exist, default to 0
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to get token version: %v", err)
	}

	tokenVersion, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to convert token version to int: %v", err)
	}

	return tokenVersion, nil
}

func incrementTokenVersion(userId string) error {
	tokenVersion, _ := getTokenVersion(userId)
	log.Println(fmt.Sprintf("session-%s", userId))
	err := rdb.Set(ctx, fmt.Sprintf("session-%s", userId), tokenVersion+1, time.Hour*5).Err()
	if err != nil {
		return fmt.Errorf("failed to increment token version: %v", err)
	}
	return nil
}

// Encode the session into a JWT token
func encodeSession(session *Session) (string, error) {

	tokenVersion, _ := getTokenVersion(session.UserId)

	// Create a JWT token with session data
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      session.UserId,
		"word":         session.Word,
		"in_progress":  session.InProgress,
		"guesses":      session.Guesses,
		"exp":          time.Now().Add(time.Hour * 24).Unix(), // 24-hour expiration
		"tokenVersion": tokenVersion,
	})

	// Sign and return the token
	return token.SignedString(jwtSecret)
}

func decodeSession(tokenString string) (*Session, error) {
	// Parse and verify the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract session data from token claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check if the token is expired
		if exp, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			if time.Now().After(expTime) {
				return nil, errors.New("token is expired")
			}
		}

		session := &Session{
			UserId:     claims["user_id"].(string),
			Word:       claims["word"].(string),
			InProgress: claims["in_progress"].(bool),
			Guesses:    []Guess{},
		}

		// Parse guesses from claims (as done before)
		if guesses, ok := claims["guesses"].([]interface{}); ok {
			for _, guessData := range guesses {
				guessMap := guessData.(map[string]interface{})
				guess := Guess{
					Word: guessMap["word"].(string),
				}
				letters := guessMap["letters"].([]interface{})
				for _, m := range letters {
					letterMap := m.(map[string]interface{})
					guessedLetter := GuessedLetter{
						Letter: letterMap["letter"].(string),
						Match:  letterMap["match"].(string),
					}
					guess.Letters = append(guess.Letters, guessedLetter)
				}
				session.Guesses = append(session.Guesses, guess)
			}
		}

		// Check tokenVersion
		if tokenVersionClaim, ok := claims["tokenVersion"].(float64); ok {
			tokenVersion := int(tokenVersionClaim)

			actualTokenVersion, err := getTokenVersion(session.UserId)
			if err != nil {
				return nil, fmt.Errorf("failed to get token version: %v", err)
			}

			if actualTokenVersion != tokenVersion {
				return nil, errors.New("token version mismatch")
			}
		}

		return session, nil
	}

	return nil, errors.New("invalid session")
}
