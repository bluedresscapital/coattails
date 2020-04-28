package wardrobe

import (
	"fmt"
)

func FetchUser(username string, password string) error {
	sql := fmt.Sprintf("INSERT INTO `%s` (`id`, `password`) VALUES (%s, %s)", "users", username, password)
	_, err := db.Query(sql)
	if err != nil {
		return err
	}
	return nil
}
