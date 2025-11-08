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
		    receiverID INTEGER,
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
	_, err := ss.db.Exec(`INSERT INTO Year(yearID,description,amount)
		VALUES (?, ?, ?);`,
		year.YearID, year.Description, year.Amount)
	if err != nil {
		return fmt.Errorf("failed to insert year: %w", err)
	}
	for _, p := range year.Pairing {
		_, err = ss.db.Exec(`INSERT INTO Year_User(yearID,buyerID,receiverID,rating)
		VALUES (?, ?, ?, ?)`,
			year.YearID, p.BuyerUserID, p.ReceiverUserID, p.Rating)
	}
	return err
}

// UpdateYear updates a year in the database
func (ss *SecretSanta) UpdateYear(year *Year) error {
	if year == nil {
		return fmt.Errorf("year couldn`t be nil")
	}
	if year.YearID == 0 {
		return fmt.Errorf("year id couldn`t be 0")
	}
	_, err := ss.db.Exec(`UPDATE Year SET description = ?, amount = ? WHERE yearID = ?`,
		year.Description, year.Amount, year.YearID)
	if err != nil {
		return fmt.Errorf("failed to upsert year: %w", err)
	}
	for _, p := range year.Pairing {
		_, err = ss.db.Exec(`INSERT INTO Year_User(yearID,buyerID,receiverID,rating)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(yearID, buyerID, receiverID) 
		DO 
		   UPDATE SET rating = ?
		   WHERE yearID = ?, buyerID = ?, receiverID = ?;`,
			year.YearID, p.BuyerUserID, p.ReceiverUserID, p.Rating,
			p.Rating, year.YearID, p.BuyerUserID, p.ReceiverUserID)
	}
	return err
}

// ListYears lists all years in the database
func (ss *SecretSanta) ListYears() ([]*Year, error) {
	rows, err := ss.db.Query("SELECT * FROM Year;")
	if err != nil {
		return nil, fmt.Errorf("failed to select year: %w", err)
	}
	var years []*Year
	for rows.Next() {
		year := &Year{
			Pairing: make([]*Pairing, 0),
		}
		rows.Scan(&year.YearID, &year.Description, &year.Amount)
		subRows, err := ss.db.Query("SELECT buyerID,receiverID,rating FROM Year_User WHERE yearID=?;", year.YearID)
		if err != nil {
			return nil, fmt.Errorf("failed to select year pairing: %w", err)
		}
		for subRows.Next() {
			var p = &Pairing{}
			subRows.Scan(&p.ReceiverUserID, &p.BuyerUserID, &p.Rating)
			year.Pairing = append(year.Pairing, p)
		}
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
	row.Scan(&year.YearID, &year.Description, &year.Amount)
	subRows, err := ss.db.Query("SELECT buyerID,receiverID,rating FROM Year_User WHERE yearID=?;", id)
	if err != nil {
		return nil, fmt.Errorf("failed to select year pairing: %w", err)
	}
	for subRows.Next() {
		var p = &Pairing{}
		subRows.Scan(&p.ReceiverUserID, &p.BuyerUserID, &p.Rating)
		year.Pairing = append(year.Pairing, p)
	}
	return year, nil
}

// RmYear removes a year by id
func (ss *SecretSanta) RmYear(id uint32) error {
	_, err := ss.db.Exec("DELETE FROM Year WHERE yearID=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete year: %w", err)
	}
	_, err = ss.db.Exec("DELETE FROM Year_User WHERE yearID=?", id)
	if err != nil {
		return fmt.Errorf("failed to delete year pairing: %w", err)
	}
	return nil
}
