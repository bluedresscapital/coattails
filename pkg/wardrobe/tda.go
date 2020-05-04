package wardrobe

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/sundress"
)

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
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, client_id_hash) DO UPDATE
		SET client_id_cipher=$3,refresh_token_cipher=$4`,
		userId, clientIdHash[:], clientIdCipher, refreshCipher)
	return err
}

type TDAccount struct {
	Id           int
	UserId       int
	ClientId     string
	RefreshToken string
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
	var td TDAccount
	var clientIdCipher []byte
	var refreshTokenCipher []byte
	err = rows.Scan(&td.Id, &td.UserId, &clientIdCipher, &refreshTokenCipher)
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
