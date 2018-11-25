package secretsanta

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

const infoTmpl = `Hallo Wichtelfreunde,

hier kommen die Daten für dieses Jahr:

Jahr: {{.Year}}
Betrag: {{.Amount}}€ 
Besonderheiten: {{.Desc}}
Du beschenkst: {{.ReceiverUser.Firstname}} {{.ReceiverUser.Lastname}} 
Rating: {{.Rating}}

Viel spaß :-)
`

// SecretSanta manage secret santa with a db. It is a Container for a BoltDB
type SecretSanta struct {
	db *bolt.DB
}

// Close closed the bolt database
func (ss *SecretSanta) Close() error {
	if ss == nil || ss.db == nil {
		return errors.New("db not set")
	}
	return ss.db.Close()
}

// New open the bolt databse
func New(dbfile string) (*SecretSanta, error) {
	db, err := bolt.Open(dbfile, 0600, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database %q", dbfile)
	}
	return NewWithDB(db)
}

// NewWithDB initilized the database and returns a container for secret santa
func NewWithDB(db *bolt.DB) (*SecretSanta, error) {
	if db == nil {
		return nil, errors.New("missing database")
	}
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(userBucket))
		if err != nil {
			return errors.Wrapf(err, "failed to create buckt: %s", userBucket)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(yearBucket))
		if err != nil {
			return errors.Wrapf(err, "failed to create buckt: %s", yearBucket)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &SecretSanta{db}, nil
}

// PrintAll write a year to the writer
func (ss *SecretSanta) PrintAll(w io.Writer, year *Year) error {
	t := tablewriter.NewWriter(w)
	t.Append([]string{
		"Year", fmt.Sprintf("%v", year.YearID),
	})
	t.Append([]string{
		"Amount", fmt.Sprintf("%v", year.Amount),
	})
	t.Append([]string{
		"Description", fmt.Sprintf("%s", year.Description),
	})
	t.Render()

	users, err := ss.ListUsers()
	if err != nil {
		return err
	}
	userName := func(id string) string {
		user := getUser(users, id)
		if user != nil {
			return fmt.Sprintf("%s,%s (%s)", user.Lastname, user.Firstname, user.Email)
		}
		return ""
	}

	fmt.Println("")
	t = tablewriter.NewWriter(w)
	t.SetHeader([]string{"No.", "Byer", "Receiver", "Rating"})
	for i, p := range year.Pairing {
		t.Append([]string{
			fmt.Sprintf("%v", i),
			userName(p.ByerUserID),
			userName(p.ReceiverUserID),
			fmt.Sprintf("%v", p.Rating),
		})
	}
	t.Render()
	return nil
}

func getUser(users []*User, id string) *User {
	for _, user := range users {
		if user.ID == id {
			return user
		}
	}
	return nil
}

// SendInfo set the information
type SendInfo func(to, text string) error

// SendInformation sends information about a year and paring
func (ss *SecretSanta) SendInformation(year *Year, tmpl string, info SendInfo) error {
	if year == nil || info == nil {
		return errors.New("year and info are required")
	}
	users, err := ss.ListUsers()
	if err != nil {
		return errors.Wrap(err, "failed to load users")
	}
	for _, v := range year.Pairing {
		bU := getUser(users, v.ByerUserID)
		if bU == nil {
			return errors.Errorf("no user with id %q found", v.ByerUserID)
		}
		rU := getUser(users, v.ReceiverUserID)
		if rU == nil {
			return errors.Errorf("no user with id %q found", v.ReceiverUserID)
		}
		if tmpl == "" {
			tmpl = infoTmpl
		}
		tmpl, err := template.New("mail").Parse(tmpl)
		if err != nil {
			return errors.Wrap(err, "failed to parse template")
		}
		w := &strings.Builder{}
		tmpl.Execute(w, struct {
			ByerUser     *User
			ReceiverUser *User
			Year         uint32
			Desc         string
			Amount       float32
			Rating       float64
		}{
			ByerUser:     bU,
			ReceiverUser: rU,
			Year:         year.YearID,
			Desc:         year.Description,
			Amount:       year.Amount,
			Rating:       v.Rating,
		})
		err = info(bU.Email, w.String())
		if err!=nil {
			return errors.Wrap(err, "failed to run info method")
		}
	}
	return nil
}
