package secretsanta

import (
	"bytes"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

func TestNewWithDB(t *testing.T) {
	db, err := bolt.Open("testdata/test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tests := []struct {
		name    string
		args    *bolt.DB
		wantErr bool
	}{
		{
			name:    "Nil",
			wantErr: true,
		},
		{
			name:    "Create",
			args:    db,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWithDB(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{
			name:    "NilInput",
			wantErr: true,
		},
		{
			name:    "TestDB",
			args:    "testdata/test.db",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, err := New(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ss != nil {
				if err := ss.Close(); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestSecretSanta_PrintAll(t *testing.T) {
	ss, err := New("testdata/secretsanta_test.db")
	if err != nil {
		t.Errorf("failed to open database: %v", err)
	}
	defer ss.Close()
	defer os.Remove("secretsanta_test.db")

	bid, err := ss.AddUser(&User{
		Email:     "byer@example.de",
		Enabled:   true,
		Firstname: "Max",
		Lastname:  "Musterman",
	})
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}
	rid, err := ss.AddUser(&User{
		Email:     "receiver@example.de",
		Enabled:   true,
		Firstname: "John",
		Lastname:  "Doe",
	})
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}

	y := &Year{
		Amount:      15.50,
		Description: "Description",
		YearID:      2018,
		Pairing: []*Pairing{
			{
				ByerUserID:     bid,
				ReceiverUserID: rid,
				Rating:         0.5,
			},
		},
	}
	buf := bytes.NewBufferString("")
	ss.PrintAll(buf, y)
	result := `+-------------+-------------+
| Year        |        2018 |
| Amount      |        15.5 |
| Description | Description |
+-------------+-------------+
+-----+--------------------------------+--------------------------------+--------+
| NO  |              BYER              |            RECEIVER            | RATING |
+-----+--------------------------------+--------------------------------+--------+
|   0 | Musterman,Max                  | Doe,John (receiver@example.de) |    0.5 |
|     | (byer@example.de)              |                                |        |
+-----+--------------------------------+--------------------------------+--------+
`
	if result != buf.String() {
		t.Errorf("output mismatch want:\n%s, got:\n%s,", result, buf.String())
	}
}

func TestSecretSanta_SendInformation(t *testing.T) {
	ss, err := New("testdata/secretsanta_test.db")
	if err != nil {
		t.Errorf("failed to open database: %v", err)
	}
	defer ss.Close()
	defer os.Remove("secretsanta_test.db")

	bid, err := ss.AddUser(&User{
		Email:     "byer@example.de",
		Enabled:   true,
		Firstname: "Max",
		Lastname:  "Musterman",
	})
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}
	rid, err := ss.AddUser(&User{
		Email:     "receiver@example.de",
		Enabled:   true,
		Firstname: "John",
		Lastname:  "Doe",
	})

	y := &Year{
		Amount:      15.50,
		Description: "Description",
		YearID:      2018,
		Pairing: []*Pairing{
			{
				ByerUserID:     bid,
				ReceiverUserID: rid,
				Rating:         0.5,
			},
		},
	}

	send := func(to, text string) error {
		if to != "byer@example.de" {
			t.Errorf("wrong to address want: byer@example.de, got: %s", to)
		}
		test := `{Max Musterman true byer@example.de}
{John Doe true receiver@example.de}
2018
Description
15.5
0.5`
		if text != test {
			t.Errorf("wrong text want: %s, got: %s", test, text)
		}
		return nil
	}
	tmpl := `{{.ByerUser}}
{{.ReceiverUser}}
{{.Year}}
{{.Desc}}
{{.Amount}}
{{.Rating}}`
	err = ss.SendInformation(y, tmpl, send)
	if err != nil {
		t.Errorf("faild to send information: %v", err)
	}
}
