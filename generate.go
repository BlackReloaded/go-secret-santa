package secretsanta

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

var errNoPartner = fmt.Errorf("no partner to select")

type userCount struct {
	index int
	count float64
	user  *User
}

// generateAttempts is how often a dead ended pairing is retried before an
// error is returned
const generateAttempts = 1000

// Generate creates a year with pairing all enabled users, if seed is 0 or
// negative, unix time will be used
func (ss *SecretSanta) Generate(year *Year, autoSave bool, seed int64) error {
	years, err := ss.ListYears()
	if err != nil {
		return fmt.Errorf("failed to list years: %w", err)
	}
	sort.Sort(sort.Reverse(YearSlice{years}))

	users, err := ss.ListUsers(false)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}
	if len(users) < 2 {
		return fmt.Errorf("need at least two users")
	}
	sort.Sort(UserSlice{users})

	if seed <= 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))

	// pair() can dead end on unlucky random choices, even though a valid
	// pairing exists, so retry with a fresh user copy until it succeeds
	var p []*Pairing
	for attempt := 0; attempt < generateAttempts; attempt++ {
		buyers := make([]*User, len(users))
		copy(buyers, users)
		receiverUsers := make([]*User, len(users))
		copy(receiverUsers, users)
		p, err = pair(years, buyers, receiverUsers, rng)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to generate pairing after %d attempts: %w", generateAttempts, err)
	}
	year.Pairing = p

	if autoSave {
		err := ss.UpdateYear(year)
		if err != nil {
			return err
		}
	}
	return nil
}

func pair(years []*Year, buyers []*User, receivers []*User, rng *rand.Rand) ([]*Pairing, error) {
	if len(buyers) == 0 && len(receivers) == 0 {
		return []*Pairing{}, nil
	} else if len(buyers) == 0 {
		return nil, fmt.Errorf("buyers could not be null: %w", errNoPartner)
	}
	buyerIndex := rng.Intn(len(buyers))
	buyer := buyers[buyerIndex]
	targetUsers := []*userCount{}
	max := 0.0
	for index, user := range receivers {
		if user == buyer {
			continue
		}
		sub := 1.0
		for i, year := range years {
			for _, r := range year.Pairing {
				if r.ReceiverUserID == user.ID && r.BuyerUserID == buyer.ID {
					sub -= 1.0 / (float64(i) + 1.0)
					break
				}
			}
		}
		if sub > 0 {
			targetUsers = append(targetUsers, &userCount{index, sub, user})
			max += sub
		}
	}
	if len(targetUsers) == 0 {
		return nil, fmt.Errorf("target user couldn`t be null%w", errNoPartner)
	}
	multi := math.Pow10(int(math.Ceil(math.Log10(max))))
	random := math.Mod(rng.Float64()*multi, max)
	var partner *userCount
	for _, uc := range targetUsers {
		random -= uc.count
		if random <= 0 {
			partner = uc
			break
		}
	}
	pairing := &Pairing{
		BuyerUserID:    buyer.ID,
		ReceiverUserID: partner.user.ID,
		Rating:         partner.count,
	}

	buyers = append(buyers[:buyerIndex], buyers[buyerIndex+1:]...)
	receivers = append(receivers[:partner.index], receivers[partner.index+1:]...)
	p, err := pair(years, buyers, receivers, rng)
	if err != nil {
		return nil, err
	}

	return append(p, pairing), nil
}
