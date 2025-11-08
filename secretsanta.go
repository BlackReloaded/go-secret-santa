package secretsanta

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const infoTmpl = `Hallo Wichtelfreunde,

hier kommen die Daten für dieses Jahr:

Jahr: {{.Year}}
Betrag: {{.Amount}}€
{{if .Desc}}
Besonderheiten: {{.Desc}}
{{end}}
Du beschenkst: {{.ReceiverUser.Firstname}} {{.ReceiverUser.Lastname}}
Rating: {{.Rating}}

Viel spaß :-)
`

// SecretSanta manage secret santa with a db. It is a Container for a BoltDB
type SecretSanta struct {
	db *sql.DB
}

// Close closed the bolt database
func (ss *SecretSanta) Close() error {
	if ss == nil || ss.db == nil {
		return fmt.Errorf("db not set")
	}
	return ss.db.Close()
}

// New open the bolt databse
func New(dbfile string) (*SecretSanta, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %q: %w", dbfile, err)
	}
	return NewWithDB(db)
}

// NewWithDB initilized the database and returns a container for secret santa
func NewWithDB(db *sql.DB) (*SecretSanta, error) {
	if db == nil {
		return nil, fmt.Errorf("missing database")
	}
	ss := &SecretSanta{db}
	err := ss.initUserTable()
	if err != nil {
		return nil, fmt.Errorf("failed to init user table: %w", err)
	}
	err = ss.initYearTable()
	if err != nil {
		return nil, fmt.Errorf("failed to init year table: %w", err)
	}
	return ss, nil
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

	users, err := ss.ListUsers(false)
	if err != nil {
		return err
	}
	userName := func(id uint32) string {
		user := getUser(users, id)
		if user != nil {
			return fmt.Sprintf("%s,%s (%s)", user.Lastname, user.Firstname, user.Email)
		}
		return ""
	}

	fmt.Println("")
	t = tablewriter.NewWriter(w)
	t.Header("No.", "Buyer", "Receiver", "Rating")
	for i, p := range year.Pairing {
		t.Append(
			fmt.Sprintf("%v", i),
			userName(p.BuyerUserID),
			userName(p.ReceiverUserID),
			fmt.Sprintf("%v", p.Rating),
		)
	}
	t.Render()
	return nil
}

func getUser(users []*User, id uint32) *User {
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
		return fmt.Errorf("year and info are required")
	}
	users, err := ss.ListUsers(false)
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}
	for _, v := range year.Pairing {
		bU := getUser(users, v.BuyerUserID)
		if bU == nil {
			return fmt.Errorf("no user with id %q found", v.BuyerUserID)
		}
		rU := getUser(users, v.ReceiverUserID)
		if rU == nil {
			return fmt.Errorf("no user with id %q found", v.ReceiverUserID)
		}
		if tmpl == "" {
			tmpl = infoTmpl
		}
		tmpl, err := template.New("mail").Parse(tmpl)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
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
		if err != nil {
			return fmt.Errorf("failed to run info method: %w", err)
		}
	}
	return nil
}
