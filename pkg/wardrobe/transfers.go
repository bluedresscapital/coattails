package wardrobe

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type Transfer struct {
	Uid           string          `json:"uid"`
	PortId        int             `json:"port_id"`
	Amount        decimal.Decimal `json:"amount"`
	IsDeposit     bool            `json:"is_deposit"`
	ManuallyAdded bool            `json:"manually_added"`
	Date          time.Time       `json:"date"`
}

// Upserts transfer into db - function is idempotent
func UpsertTransfer(uid string, portId int, amount decimal.Decimal, isDeposit bool, manuallyAdded bool, date time.Time) error {
	_, err := db.Exec(`
		INSERT INTO transfers (uid, port_id, amount, is_deposit, manually_added, date) 
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (uid) DO UPDATE
		SET port_id=$2,amount=$3,is_deposit=$4,manually_added=$5,date=$6`,
		uid, portId, amount.StringFixedBank(4), isDeposit, manuallyAdded, date)
	return err
}

// NOTE(ma): See if we still need this function, currently unused
func FetchTransferByUid(uid string) (*Transfer, error) {
	rows, err := db.Query(`
		SELECT uid, port_id, amount, is_deposit, manually_added, date 
		FROM transfers 
		WHERE uid=$1`,
		uid)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("unable to find transfer with uid %s", uid)
	}
	var transfer Transfer
	err = rows.Scan(&transfer.Uid, &transfer.PortId, &transfer.Amount, &transfer.IsDeposit, &transfer.ManuallyAdded, &transfer.Date)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple transfers found with uid %s", uid)
	}
	return &transfer, nil
}

func FetchTransfersbyUserId(userId int) ([]Transfer, error) {
	rows, err := db.Query(`
		SELECT uid, port_id, amount, is_deposit, manually_added, date 
		FROM transfers t
		JOIN portfolios p ON t.port_id=p.id
		WHERE p.user_id=$1`, userId)
	if err != nil {
		return nil, err
	}
	var transfers []Transfer
	for rows.Next() {
		var t Transfer
		err = rows.Scan(&t.Uid, &t.PortId, &t.Amount, &t.IsDeposit, &t.ManuallyAdded, &t.Date)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}
	if transfers == nil {
		return make([]Transfer, 0), nil
	}
	return transfers, nil
}

func DeleteTransfer(uid string, portId int) error {
	_, err := db.Exec(`DELETE FROM transfers WHERE uid=$1 AND port_id=$2`, uid, portId)
	return err
}
