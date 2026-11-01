package secretsanta

import (
	"fmt"
)

// Year is a represatation of a year
type Year struct {
	YearID      uint32
	Description string
	Amount      float32
	Pairing     []*Pairing
}

// Pairing describe a with user
type Pairing struct {
	BuyerUserID    uint32
	ReceiverUserID uint32
	Rating         float64
}

// YearSlice is a type to sort a year slices
type YearSlice struct {
	Years []*Year
}

func (ys YearSlice) Len() int {
	return len(ys.Years)
}

func (ys YearSlice) Less(i, j int) bool {
	return ys.Years[i].YearID < ys.Years[j].YearID
}

func (ys YearSlice) Swap(i, j int) {
	ys.Years[i], ys.Years[j] = ys.Years[j], ys.Years[i]
}

func (ss *SecretSanta) initYearTable() error {
	_, err := ss.db.Exec(`CREATE TABLE IF NOT EXISTS Year (
		    yearID INTEGER PRIMARY KEY,
		    description TEXT,
    		amount NUMERIC
		);`)
	if err != nil {
		return fmt.Errorf("failed to create yqr table: %w", err)
	}
	_, err = ss.db.Exec(`CREATE TABLE IF NOT EXISTS Year_User (
		    yearID INTEGER REFERENCES Year,
		    buyerID INTEGER REFERENCES User,
		    receiverID INTEGER REFERENCES User,
    		rating NUMERIC,
    		PRIMARY KEY(yearID, buyerID, receiverID)
		);`)
	return err
}

// AddYear add a year to the database
func (ss *SecretSanta) AddYear(year *Year) error {
	if year == nil {
		return fmt.Errorf("year couldn`t be nil")
	}
	if year.YearID == 0 {
		return fmt.Errorf("year id couldn`t be 0")
	}
	tx, err := ss.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO Year(yearID,description,amount)
		VALUES (?, ?, ?);`,
		year.YearID, year.Description, year.Amount)
	if err != nil {
		return fmt.Errorf("failed to insert year: %w", err)
	}
	for _, p := range year.Pairing {
		_, err = tx.Exec(`INSERT INTO Year_User(yearID,buyerID,receiverID,rating)
		VALUES (?, ?, ?, ?)`,
			year.YearID, p.BuyerUserID, p.ReceiverUserID, p.Rating)
		if err != nil {
			return fmt.Errorf("failed to insert pairing: %w", err)
		}
	}
	return tx.Commit()
}

// UpdateYear updates a year in the database
func (ss *SecretSanta) UpdateYear(year *Year) error {
	if year == nil {
		return fmt.Errorf("year couldn`t be nil")
	}
	if year.YearID == 0 {
		return fmt.Errorf("year id couldn`t be 0")
	}
	// If the year does not exist yet it is created, all pairings of the
	// year are replaced with the given ones
	tx, err := ss.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE Year SET description = ?, amount = ? WHERE yearID = ?`,
		year.Description, year.Amount, year.YearID)
	if err != nil {
		return fmt.Errorf("failed to update year: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		_, err = tx.Exec(`INSERT INTO Year(yearID,description,amount)
			VALUES (?, ?, ?)`,
			year.YearID, year.Description, year.Amount)
		if err != nil {
			return fmt.Errorf("failed to insert year: %w", err)
		}
	}
	_, err = tx.Exec(`DELETE FROM Year_User WHERE yearID = ?`, year.YearID)
	if err != nil {
		return fmt.Errorf("failed to delete pairings: %w", err)
	}
	for _, p := range year.Pairing {
		_, err = tx.Exec(`INSERT INTO Year_User(yearID,buyerID,receiverID,rating)
		VALUES (?, ?, ?, ?)`,
			year.YearID, p.BuyerUserID, p.ReceiverUserID, p.Rating)
		if err != nil {
			return fmt.Errorf("failed to insert pairing: %w", err)
		}
	}
	return tx.Commit()
}

// ListYears lists all years in the database
func (ss *SecretSanta) ListYears() ([]*Year, error) {
	rows, err := ss.db.Query("SELECT * FROM Year;")
	if err != nil {
		return nil, fmt.Errorf("failed to select year: %w", err)
	}
	defer rows.Close()
	var years []*Year
	for rows.Next() {
		year := &Year{
			Pairing: make([]*Pairing, 0),
		}
		if err := rows.Scan(&year.YearID, &year.Description, &year.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan year row: %w", err)
		}
		subRows, err := ss.db.Query("SELECT buyerID,receiverID,rating FROM Year_User WHERE yearID=?;", year.YearID)
		if err != nil {
			return nil, fmt.Errorf("failed to select year pairing: %w", err)
		}
		for subRows.Next() {
			var p = &Pairing{}
			if err := subRows.Scan(&p.BuyerUserID, &p.ReceiverUserID, &p.Rating); err != nil {
				subRows.Close()
				return nil, fmt.Errorf("failed to scan pairing row: %w", err)
			}
			year.Pairing = append(year.Pairing, p)
		}
		subRows.Close()
		years = append(years, year)
	}
	return years, nil
}

// GetYear gets a year by id
func (ss *SecretSanta) GetYear(id uint32) (*Year, error) {
	row := ss.db.QueryRow("SELECT * FROM Year WHERE yearID=?;", id)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed to select year: %w", row.Err())
	}
	year := &Year{
		Pairing: make([]*Pairing, 0),
	}
	if err := row.Scan(&year.YearID, &year.Description, &year.Amount); err != nil {
		return nil, fmt.Errorf("failed to scan year row: %w", err)
	}
	subRows, err := ss.db.Query("SELECT buyerID,receiverID,rating FROM Year_User WHERE yearID=?;", id)
	if err != nil {
		return nil, fmt.Errorf("failed to select year pairing: %w", err)
	}
	defer subRows.Close()
	for subRows.Next() {
		var p = &Pairing{}
		if err := subRows.Scan(&p.BuyerUserID, &p.ReceiverUserID, &p.Rating); err != nil {
			return nil, fmt.Errorf("failed to scan pairing row: %w", err)
		}
		year.Pairing = append(year.Pairing, p)
	}
	return year, nil
}

// RmYear removes a year by id
func (ss *SecretSanta) RmYear(id uint32) error {
	tx, err := ss.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.Exec("DELETE FROM Year WHERE yearID=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete year: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("failed to delete year: no year with id %d", id)
	}
	_, err = tx.Exec("DELETE FROM Year_User WHERE yearID=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete year pairing: %w", err)
	}
	return tx.Commit()
}
