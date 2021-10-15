package secretsanta

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/pkg/errors"
)

var errNoPartner = errors.New("no partner to select")

type userCount struct {
	index int
	count float64
	user  *User
}

// Generate creates a year with pairing all enabled users, if seed is 0, unix time will be used
func (ss *SecretSanta) Generate(year *Year, autoSave bool, seed int64) error {
	years, err := ss.ListYears()
	if err != nil {
		return errors.Wrap(err, "failed to list years")
	}
	sort.Sort(sort.Reverse(YearSlice{years}))

	users, err := ss.ListUsers(false)
	if err != nil {
		return errors.Wrap(err, "failed to list users")
	}
	if len(users) == 0 {
		return errors.New("need more user")
	}
	sort.Sort(UserSlice{users})

	rand.Seed(seed)
	if seed <= 0 {
		rand.Seed(time.Now().UnixNano())
	}
	receiverUsers := make([]*User, len(users))
	copy(receiverUsers, users)
	p, err := pair(years, users, receiverUsers)
	if err != nil {
		return errors.Wrap(err, "failed to generate pairing")
	}
	year.Pairing = p

	if autoSave {
		_, err := ss.AddYear(year)
		if err != nil {
			return err
		}
	}
	return nil
}

func pair(years []*Year, byers []*User, receivers []*User) ([]*Pairing, error) {
	if len(byers) == 0 && len(receivers) == 0 {
		return []*Pairing{}, nil
	} else if len(byers) == 0 {
		return nil, errNoPartner
	}
	byerIndex := rand.Intn(len(byers))
	byer := byers[byerIndex]
	targetUsers := []*userCount{}
	max := 0.0
	for index, user := range receivers {
		if user == byer {
			continue
		}
		sub := 1.0
		for i, year := range years {
			for _, r := range year.Pairing {
				if r.ReceiverUserID == user.ID && r.ByerUserID == byer.ID {
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
		return nil, errNoPartner
	}
	multi := math.Pow10(int(math.Ceil(math.Log10(max))))
	random := math.Mod(rand.Float64()*multi, max)
	var partner *userCount
	for _, uc := range targetUsers {
		random -= uc.count
		if random <= 0 {
			partner = uc
			break
		}
	}
	pairing := &Pairing{
		ByerUserID:     byer.ID,
		ReceiverUserID: partner.user.ID,
		Rating:         partner.count,
	}

	byers = append(byers[:byerIndex], byers[byerIndex+1:]...)
	receivers = append(receivers[:partner.index], receivers[partner.index+1:]...)
	p, err := pair(years, byers, receivers)
	if err != nil {
		return nil, err
	}

	return append(p, pairing), nil
}
