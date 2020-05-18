package wardrobe

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Transfer struct {
	Uid           string          `json:"uid"`
	PortId        int             `json:"port_id"`
	Amount        decimal.Decimal `json:"amount"`
	IsDeposit     bool            `json:"is_deposit"`
	ManuallyAdded bool            `json:"manually_added"`
	Date          time.Time       `json:"date"`
}

// Inserts transfer into db, ignores if uid already exists
func InsertIgnoreTransfer(t Transfer) error {
	_, err := db.Exec(`
		INSERT INTO transfers (uid, port_id, amount, is_deposit, manually_added, date, committed) 
		VALUES ($1,$2,$3,$4,$5,$6,false)
		ON CONFLICT (uid) DO NOTHING`,
		t.Uid, t.PortId, t.Amount.StringFixedBank(4), t.IsDeposit, t.ManuallyAdded, t.Date)
	return err
}

// Upserts transfer into db - function is idempotent
// WARNING: This should only be called by the manual upsert transfer handler.
// If an automated system calls this function, we will always have uncommitted orders
// and we'll be re-running alot of reloading data
func UpsertTransfer(t Transfer) error {
	_, err := db.Exec(`
		INSERT INTO transfers (uid, port_id, amount, is_deposit, manually_added, date, committed) 
		VALUES ($1,$2,$3,$4,$5,$6, false)
		ON CONFLICT (uid) DO UPDATE
		SET port_id=$2,amount=$3,is_deposit=$4,manually_added=$5,date=$6,committed=false`,
		t.Uid, t.PortId, t.Amount.StringFixedBank(4), t.IsDeposit, t.ManuallyAdded, t.Date)
	return err
}

func FetchTransfersbyUserId(userId int) ([]Transfer, error) {
	rows, err := db.Query(`
		SELECT uid, port_id, amount, is_deposit, manually_added, date 
		FROM transfers t
		JOIN portfolios p ON t.port_id=p.id
		WHERE p.user_id=$1
		ORDER BY date`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func FetchTransfersByPortfolioId(portId int) ([]Transfer, error) {
	rows, err := db.Query(`
		SELECT uid, port_id, amount, is_deposit, manually_added, date 
		FROM transfers t
		WHERE t.port_id=$1
		ORDER BY date`, portId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func SetTransfersCommitted(portId int) error {
	_, err := db.Exec(`UPDATE transfers SET committed=true WHERE port_id=$1`, portId)
	return err
}

func HasUncommittedTransfers(portId int) (bool, error) {
	rows, err := db.Query(`SELECT COUNT(*) FROM transfers WHERE committed=false AND port_id=$1`, portId)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, fmt.Errorf("no rows returned from count")
	}
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
