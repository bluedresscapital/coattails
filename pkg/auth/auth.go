package auth

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/google/uuid"
)

// Logs user in with input credentials, and then returns (valid) auth token
func Login(username string, password [32]byte) (*string, error) {
	return _login(username, password)
}

func Register(username string, password [32]byte) (*string, error) {
	// Register is basically the same as login, except we need to create user first
	err := wardrobe.CreateUser(username, password)
	if err != nil {
		return nil, err
	}
	return _login(username, password)
}

func _login(username string, password [32]byte) (*string, error) {
	userId, err := wardrobe.FetchUser(username, password)
	if err != nil {
		return nil, err
	}
	authToken := uuid.New().String()
	err = wardrobe.SetExpiringAuthToken(authToken, userId)
	if err != nil {
		return nil, err
	}
	return &authToken, nil
}
