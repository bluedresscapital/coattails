package wardrobe

func UpsertStock(ticker string) error {
	_, err := db.Exec(`INSERT INTO stocks (ticker) VALUES ($1) ON CONFLICT (ticker) DO NOTHING`, ticker)
	return err
}
