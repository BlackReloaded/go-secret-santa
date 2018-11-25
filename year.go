package secretsanta

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	yearBucket = "Year"
)

// Year is a represatation of a year
type Year struct {
	ID          string
	YearID      uint32
	Description string
	Amount      float32
	Pairing     []*Pairing
}

// Pairing describe a with user
type Pairing struct {
	ByerUserID     string
	ReceiverUserID string
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

// AddYear add a year to the database
func (ss *SecretSanta) AddYear(year *Year) (string, error) {
	if year == nil {
		return "", errors.New("year is nil")
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "failed to create uuid")
	}
	year.ID = id.String()
	key, err := id.MarshalBinary()
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal uuid")
	}
	buf, err := json.Marshal(year)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal year to json")
	}
	err = ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(yearBucket))

		if err := b.Put(key, buf); err != nil {
			return errors.Wrap(err, "failed to save year to db")
		}

		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create year")
	}
	return id.String(), nil
}

// UdateYear updates a year in the database
func (ss *SecretSanta) UdateYear(year *Year) error {
	if year == nil {
		return errors.New("user is nil")
	}
	buf, err := json.Marshal(year)
	if err != nil {
		return errors.Wrap(err, "failed to marshal user to json")
	}
	uid, err := uuid.Parse(year.ID)
	if err != nil {
		return errors.Wrap(err, "failed to parse uuid")
	}
	key, err := uid.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "failed to marshal uuid")
	}
	err = ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(yearBucket))
		if len(b.Get(key)) == 0 {
			return errors.New("no year found to update")
		}
		if err := b.Put(key, buf); err != nil {
			return errors.Wrap(err, "failed to save year to db")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to update year")
	}
	return nil
}

// ListYears lists all years in the database
func (ss *SecretSanta) ListYears() ([]*Year, error) {
	years := []*Year{}
	err := ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(yearBucket))
		b.ForEach(func(k, v []byte) error {
			year := &Year{}
			if err := json.Unmarshal(v, year); err != nil {
				return errors.Wrap(err, "failed to unmarshal user")
			}
			years = append(years, year)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list years")
	}
	return years, nil
}

// GetYear gets a year by id
func (ss *SecretSanta) GetYear(id string) (*Year, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uuid")
	}
	key, err := uid.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal uuid")
	}
	year := &Year{}
	err = ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(yearBucket))
		v := b.Get(key)
		if err := json.Unmarshal(v, year); err != nil {
			return errors.Wrap(err, "failed to unmarshal year")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to load year")
	}
	return year, nil
}


// RmYear removes a year by id
func (ss *SecretSanta) RmYear(id string) (error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.Wrap(err, "failed to parse uuid")
	}
	key, err := uid.MarshalBinary()
	if err != nil {
		return errors.Wrap(err, "failed to marshal uuid")
	}
	err = ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(yearBucket))
		return b.Delete(key)
	})
	return err
}
