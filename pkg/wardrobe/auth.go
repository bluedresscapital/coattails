package wardrobe

import (
	"fmt"
	"strconv"
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
	defer rows.Close()
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

func FetchUserById(id int) (*string, error) {
	rows, err := db.Query("SELECT username FROM users WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("unable to find user with id %d", id)
	}
	username := new(string)
	err = rows.Scan(username)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple users returned with id %d", id)
	}
	return username, nil
}

func CreateUser(username string, password [32]byte) error {
	_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password[:])
	if err != nil {
		return err
	}
	return nil
}

// TODO DEPRECATE THIS ONCE U GET TO SESSION STUFF
func FetchAuthToken(sessionToken string) (*string, error) {
	res, err := cache.Get(sessionToken).Result()
	if err != nil {
		return nil, err
	}
	tok := new(string)
	// Looks like this is the "correct" way to convert interface to string:
	// https://yourbasic.org/golang/interface-to-string/
	*tok = fmt.Sprintf("%s", res)
	return tok, nil
}

// Given cookie, verifies it by fetching in cache
func VerifyCookie(cookie string) (*int, error) {
	// Assume we map cookie to userId
	userId, err := cache.Get(cookie).Result()
	if err != nil {
		return nil, err
	}
	var i = new(int)
	*i, err = strconv.Atoi(userId)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Sets an expiring auth token into cache
func SetExpiringAuthToken(token string, userId *int) error {
	err := cache.SetNX(token, *userId, SessionTokenTtl).Err()
	if err != nil {
		return err
	}
	return nil
}

func ClearAuthToken(sessionToken string) error {
	err := cache.Del(sessionToken).Err()
	if err != nil {
		return err
	}
	return nil
}
