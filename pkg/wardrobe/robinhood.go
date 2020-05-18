package wardrobe

import (
	"fmt"

	"github.com/bluedresscapital/coattails/pkg/secrets"
)

type RHAccount struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceTok  string `json:"device_token"`
	RefreshTok string `json:"refresh_token"`
}

func CreateRHPortfolio(userId int, name string, username string, password string, deviceTok string, refreshTok string) error {
	usernameHash := secrets.Hash(username)
	usernameCipher, err := secrets.BdcEncrypt(username)
	if err != nil {
		return err
	}
	passwordCipher, err := secrets.BdcEncrypt(password)
	if err != nil {
		return err
	}
	deviceTokCipher, err := secrets.BdcEncrypt(deviceTok)
	if err != nil {
		return err
	}
	refreshTokCipher, err := secrets.BdcEncrypt(refreshTok)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO rh_accounts (user_id, username_hash, username_cipher, password_cipher, device_token_cipher, refresh_token_cipher)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userId, usernameHash[:], usernameCipher, passwordCipher, deviceTokCipher, refreshTokCipher)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO portfolios (user_id, name, type, rh_account_id)
		VALUES ($1, $2, 'rh', currval(pg_get_serial_sequence('rh_accounts', 'id')))
		`, userId, name)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func UpdateRHRefreshToken(id int, refreshTok string) error {
	refreshTokCipher, err := secrets.BdcEncrypt(refreshTok)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE rh_accounts SET refresh_token_cipher=$1 WHERE id=$2`, refreshTokCipher, id)
	return err
}

func FetchRHAccount(id int) (*RHAccount, error) {
	rows, err := db.Query(`SELECT id, user_id, username_cipher, password_cipher, device_token_cipher, refresh_token_cipher FROM rh_accounts WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("no rh account found with id %d", id)
	}
	var rhAcc RHAccount
	var usernameCipher []byte
	var passwordCipher []byte
	var deviceTokenCipher []byte
	var refreshTokenCipher []byte
	err = rows.Scan(&rhAcc.Id, &rhAcc.UserId, &usernameCipher, &passwordCipher, &deviceTokenCipher, &refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	username, err := secrets.BdcDecrypt(usernameCipher)
	if err != nil {
		return nil, err
	}
	password, err := secrets.BdcDecrypt(passwordCipher)
	if err != nil {
		return nil, err
	}
	deviceTok, err := secrets.BdcDecrypt(deviceTokenCipher)
	if err != nil {
		return nil, err
	}
	refreshTok, err := secrets.BdcDecrypt(refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	rhAcc.Username = *username
	rhAcc.Password = *password
	rhAcc.DeviceTok = *deviceTok
	rhAcc.RefreshTok = *refreshTok
	return &rhAcc, nil
}

func GetStockFromInstrumentId(instrumentId string) (*string, error) {
	res, err := cache.Get(instrumentId).Result()
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func SetStockFromInstrument(instrument string, stock string) error {
	_, err := cache.Set(instrument, stock, 0).Result()
	return err
}
