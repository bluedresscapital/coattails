package wardrobe

import (
	"database/sql"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/sundress"
)

// Creates a TD Portfolio - this will insert a tda_accounts object, as well as
// insert a portfolio object
func CreateTDPortfolio(userId int, name string, accountNum string, refreshToken string) error {
	refreshCipher, err := sundress.BdcEncrypt(refreshToken)
	if err != nil {
		return err
	}
	accountNumHash := sundress.Hash(accountNum)
	accountNumCipher, err := sundress.BdcEncrypt(accountNum)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO tda_accounts (user_id, account_num_hash, account_num_cipher, refresh_token_cipher)
		VALUES ($1, $2, $3, $4)
		`, userId, accountNumHash[:], accountNumCipher, refreshCipher)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO portfolios (user_id, name, type, tda_account_id)
		VALUES ($1, $2, 'tda', currval(pg_get_serial_sequence('tda_accounts', 'id')))
		`, userId, name)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Updates tda account
func UpdateRefreshToken(accountId int, refreshToken string) error {
	refreshCipher, err := sundress.BdcEncrypt(refreshToken)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		UPDATE tda_accounts
		SET refresh_token_cipher=$1
		WHERE id=$2`, refreshCipher, accountId)
	return err
}

type TDAccount struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	AccountNum   string `json:"account_num"`
	RefreshToken string `json:"refresh_token"`
}

func FetchTDAccount(id int) (*TDAccount, error) {
	rows, err := db.Query("SELECT id, user_id, account_num_cipher, refresh_token_cipher FROM tda_accounts WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no td account with id %d", id)
	}
	acc, err := fetchTDAccountFromRows(rows)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple td accounts four with id %d", id)
	}
	return acc, nil
}

func FetchTDAccountsByUserId(userId int) ([]TDAccount, error) {
	rows, err := db.Query("SELECT id, user_id, account_num_cipher, refresh_token_cipher FROM tda_accounts WHERE user_id=$1", userId)
	if err != nil {
		return nil, err
	}
	var accounts []TDAccount
	for rows.Next() {
		acc, err := fetchTDAccountFromRows(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *acc)
	}
	if accounts == nil {
		return make([]TDAccount, 0), nil
	}
	return accounts, nil
}

func fetchTDAccountFromRows(rows *sql.Rows) (*TDAccount, error) {
	var td TDAccount
	var refreshTokenCipher []byte
	var accountNumCipher []byte
	err := rows.Scan(&td.Id, &td.UserId, &accountNumCipher, &refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	refreshToken, err := sundress.BdcDecrypt(refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	accountNum, err := sundress.BdcDecrypt(accountNumCipher)
	if err != nil {
		return nil, err
	}
	td.RefreshToken = *refreshToken
	td.AccountNum = *accountNum
	return &td, nil
}
