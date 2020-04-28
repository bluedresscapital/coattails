package auth

import "log"

func Login(username string, password string) {
	log.Printf("Trying to login! %s, %s", username, password)
}
