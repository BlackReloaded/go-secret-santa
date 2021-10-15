package secretsanta

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	userBucket = "User"
)

// User is a represatation of participant
type User struct {
	ID        string
	Firstname string
	Lastname  string
	Enabled   bool
	Email     string
}

// UserSlice is a type to sort a year slices
type UserSlice struct {
	Users []*User
}

func (us UserSlice) Len() int {
	return len(us.Users)
}

func (us UserSlice) Less(i, j int) bool {
	return us.Users[i].Lastname < us.Users[j].Lastname
}

func (us UserSlice) Swap(i, j int) {
	us.Users[i], us.Users[j] = us.Users[j], us.Users[i]
}

// AddUser add a user to the database
func (ss *SecretSanta) AddUser(user *User) (string, error) {
	if user == nil {
		return "", errors.New("user is nil")
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "failed to create uuid")
	}
	user.ID = id.String()
	key, err := id.MarshalText()
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal uuid")
	}
	buf, err := json.Marshal(user)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal user to json")
	}
	err = ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))

		if err := b.Put(key, buf); err != nil {
			return errors.Wrap(err, "failed to save user to db")
		}

		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create user")
	}
	return id.String(), nil
}

// UdateUser updates a user in the database
func (ss *SecretSanta) UdateUser(user *User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	buf, err := json.Marshal(user)
	if err != nil {
		return errors.Wrap(err, "failed to marshal user to json")
	}
	uid, err := uuid.Parse(user.ID)
	if err != nil {
		return errors.Wrap(err, "failed to parse uuid")
	}
	key, err := uid.MarshalText()
	if err != nil {
		return errors.Wrap(err, "failed to marshal uuid")
	}
	err = ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))
		if len(b.Get(key)) == 0 {
			return errors.New("no user found to update")
		}
		if err := b.Put(key, buf); err != nil {
			return errors.Wrap(err, "failed to save user to db")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}
	return nil
}

// ListUsers lists all user in the database
func (ss *SecretSanta) ListUsers(disabled bool) ([]*User, error) {
	users := []*User{}
	err := ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))
		b.ForEach(func(k, v []byte) error {
			user := &User{}
			if err := json.Unmarshal(v, user); err != nil {
				return errors.Wrap(err, "failed to unmarshal user")
			}
			if !user.Enabled && !disabled {
				return nil
			}
			users = append(users, user)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list users")
	}
	return users, nil
}

// GetUser gets a user by id
func (ss *SecretSanta) GetUser(id string) (*User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uuid")
	}
	key, err := uid.MarshalText()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal uuid")
	}
	user := &User{}
	err = ss.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))
		v := b.Get(key)
		if err := json.Unmarshal(v, user); err != nil {
			return errors.Wrap(err, "failed to unmarshal user")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to load user")
	}
	return user, nil
}

// RemoveUser remove a user by id
func (ss *SecretSanta) RemoveUser(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.Wrap(err, "failed to parse uuid")
	}
	key, err := uid.MarshalText()
	if err != nil {
		return errors.Wrap(err, "failed to marshal uuid")
	}
	err = ss.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))
		return b.Delete(key)
	})
	return err
}
