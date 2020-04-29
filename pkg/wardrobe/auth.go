package wardrobe

import (
	"fmt"
	"time"
)

var (
	SessionTokenTtl = time.Duration(24*7) * time.Hour
)

type AuthToken struct {
	UserId    int
	Token     string
	TtlSec    int
	CreatedAt time.Time
}

// Finds user row and returns its id
func FetchUser(username string, password [32]byte) (*int, error) {
	rows, err := db.Query("SELECT id FROM users WHERE username=$1 and password=$2", username, password[:])
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("invalid credentials for username %s", username)
	}
	id := new(int)
	err = rows.Scan(id)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple users returned with username %s", username)
	}
	return id, nil
}

func CreateUser(username string, password [32]byte) error {
	_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password[:])
	if err != nil {
		return err
	}
	return nil
}

// Fetches auth token given user_id. If no auth token is found, will return nil as auth token
func FetchAuthToken(sessionToken string) (*string, error) {
	res, err := cache.Do("GET", sessionToken)
	if err != nil {
		return nil, err
	}
	tok := new(string)
	// Looks like this is the "correct" way to convert interface to string:
	// https://yourbasic.org/golang/interface-to-string/
	*tok = fmt.Sprintf("%s", res)
	return tok, nil
}

// Sets an expiring auth token into cache
func SetExpiringAuthToken(username string, token string) error {
	_, err := cache.Do("SETEX", token, SessionTokenTtl.Seconds(), username)
	if err != nil {
		return err
	}
	return nil
}

func ClearAuthToken(sessionToken string) error {
	_, err := cache.Do("DEL", sessionToken)
	if err != nil {
		return err
	}
	return nil
}
