package secretsanta

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestSecretSanta_Generate(t *testing.T) {
	ss, err := New("testdata/generate_test.db")
	if err != nil {
		t.Errorf("failed to open database: %v", err)
	}
	defer func() {
		ss.Close()
		os.Remove("testdata/generate_test.db")
	}()

	uids := []string{}
	for i := 0; i < 4; i++ {
		uid, err := ss.AddUser(&User{
			Email:     fmt.Sprintf("%d@example.de", i),
			Enabled:   true,
			Firstname: fmt.Sprintf("Firstname_%d", i),
			Lastname:  fmt.Sprintf("Lastname_%d", i),
		})
		if err != nil {
			t.Errorf("failed to create user: %v", err)
		}
		uids = append(uids, uid)
	}
	lastYear := &Year{
		YearID:      2017,
		Description: "Description",
		Amount:      15.90,
		Pairing:     []*Pairing{},
	}
	for i := 0; i < len(uids); i++ {
		j := i + 1
		if j >= len(uids) {
			j = 0
		}
		p := &Pairing{
			Rating:         1,
			ByerUserID:     uids[i],
			ReceiverUserID: uids[j],
		}
		lastYear.Pairing = append(lastYear.Pairing, p)
	}
	_, err = ss.AddYear(lastYear)
	if err != nil {
		t.Fatal("failed to save last year")
	}
	lastYear = &Year{
		YearID:      2016,
		Description: "Description",
		Amount:      15.90,
		Pairing:     []*Pairing{},
	}
	for i := 0; i < len(uids); i++ {
		j := i + 2
		if j == len(uids) {
			j = 0
		} else if j > len(uids) {
			j = 1
		}
		p := &Pairing{
			Rating:         1,
			ByerUserID:     uids[i],
			ReceiverUserID: uids[j],
		}
		lastYear.Pairing = append(lastYear.Pairing, p)
	}
	_, err = ss.AddYear(lastYear)
	if err != nil {
		t.Fatal("failed to save last year")
	}
	year := &Year{
		YearID:      2018,
		Description: "Alles neu",
		Amount:      15.50,
	}
	err = ss.Generate(year, true, 1541451566473193900)
	if err != nil {
		t.Fatalf("failed to generte year: %v", err)
	}
	buf := bytes.NewBufferString("")
	ss.PrintAll(buf, year)
	test := `+-------------+-----------+
| Year        |      2018 |
| Amount      |      15.5 |
| Description | Alles neu |
+-------------+-----------+
+-----+--------------------------------+--------------------------------+--------+
| NO  |              BYER              |            RECEIVER            | RATING |
+-----+--------------------------------+--------------------------------+--------+
|   0 | Lastname_0,Firstname_0         | Lastname_1,Firstname_1         |    0.5 |
|     | (0@example.de)                 | (1@example.de)                 |        |
|   1 | Lastname_2,Firstname_2         | Lastname_3,Firstname_3         |    0.5 |
|     | (2@example.de)                 | (3@example.de)                 |        |
|   2 | Lastname_3,Firstname_3         | Lastname_2,Firstname_2         |      1 |
|     | (3@example.de)                 | (2@example.de)                 |        |
|   3 | Lastname_1,Firstname_1         | Lastname_0,Firstname_0         |      1 |
|     | (1@example.de)                 | (0@example.de)                 |        |
+-----+--------------------------------+--------------------------------+--------+
`
	if test != buf.String() {
		t.Errorf("failed to generate year: want %s, got %s,", test, buf.String())
	}

	var b strings.Builder
	err = ss.SendInformation(year, "{{.ByerUser.Email}}->{{.ReceiverUser.Email}}", func(to, text string) error {
		b.WriteString(to)
		b.WriteString(":")
		b.WriteString(text)
		b.WriteString("\n")
		return nil
	})

	if err != nil {
		t.Errorf("failed to send information: %v", err)
	}

	test = `0@example.de:0@example.de->1@example.de
2@example.de:2@example.de->3@example.de
3@example.de:3@example.de->2@example.de
1@example.de:1@example.de->0@example.de
`

	if test != b.String() {
		t.Errorf("failed to generate year: want %s, got %s,", test, b.String())
	}
}
