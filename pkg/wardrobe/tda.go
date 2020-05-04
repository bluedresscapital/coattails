package wardrobe

import (
	"database/sql"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/sundress"
)

// Creates a TD Portfolio - this will insert a tda_accounts object, as well as
// insert a portfolio object
func CreateTDPortfolio(userId int, name string, clientId string, refreshToken string) error {
	refreshCipher, err := sundress.BdcEncrypt(refreshToken)
	if err != nil {
		return err
	}
	clientIdHash := sundress.Hash(clientId)
	clientIdCipher, err := sundress.BdcEncrypt(clientId)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO tda_accounts (user_id, client_id_hash, client_id_cipher, refresh_token_cipher)
		VALUES ($1, $2, $3, $4)
		`, userId, clientIdHash[:], clientIdCipher, refreshCipher)
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
func UpsertTDAccount(userId int, clientId string, refreshToken string) error {
	refreshCipher, err := sundress.BdcEncrypt(refreshToken)
	if err != nil {
		return err
	}
	clientIdHash := sundress.Hash(clientId)

	clientIdCipher, err := sundress.BdcEncrypt(clientId)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO tda_accounts (user_id, client_id_hash, client_id_cipher, refresh_token_cipher) 
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (user_id, client_id_hash) DO UPDATE
		SET client_id_cipher=$3,refresh_token_cipher=$4`,
		userId, clientIdHash[:], clientIdCipher, refreshCipher)
	return err
}

type TDAccount struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id"`
	ClientId     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

func FetchTDAccount(id int, userId int) (*TDAccount, error) {
	rows, err := db.Query("SELECT id, user_id, client_id_cipher, refresh_token_cipher FROM tda_accounts WHERE id=$1 and user_id=$2",
		id, userId)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no td account with id %d for user %d", id, userId)
	}
	acc, err := fetchTDAccountFromRows(rows)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple td accounts four with id %d and user %d", id, userId)
	}
	return acc, nil
}

func FetchTDAccountsByUserId(userId int) ([]TDAccount, error) {
	rows, err := db.Query("SELECT id, user_id, client_id_cipher, refresh_token_cipher FROM tda_accounts WHERE user_id=$1", userId)
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
	var clientIdCipher []byte
	var refreshTokenCipher []byte
	err := rows.Scan(&td.Id, &td.UserId, &clientIdCipher, &refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	clientId, err := sundress.BdcDecrypt(clientIdCipher)
	if err != nil {
		return nil, err
	}
	refreshToken, err := sundress.BdcDecrypt(refreshTokenCipher)
	if err != nil {
		return nil, err
	}
	td.ClientId = *clientId
	td.RefreshToken = *refreshToken
	return &td, nil
}
