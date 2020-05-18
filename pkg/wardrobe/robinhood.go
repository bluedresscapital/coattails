package wardrobe

import (
	"fmt"

	"github.com/bluedresscapital/coattails/pkg/secrets"
)

type RHAccount struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id"`
	RefreshTok string `json:"refresh_token"`
}

func CreateRHPortfolio(userId int, name string, refreshTok string) error {
	refreshTokCipher, err := secrets.BdcEncrypt(refreshTok)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO rh_accounts (user_id, name, refresh_token_cipher)
		VALUES ($1, $2, $3)
	`, userId, name, refreshTokCipher)
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
	rows, err := db.Query(`SELECT id, user_id, refresh_token_cipher FROM rh_accounts WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("no rh account found with id %d", id)
	}
	var rhAcc RHAccount
	var refreshTokenCipher []byte
	err = rows.Scan(&rhAcc.Id, &rhAcc.UserId, &refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	refreshTok, err := secrets.BdcDecrypt(refreshTokenCipher)
	if err != nil {
		return nil, err
	}
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
