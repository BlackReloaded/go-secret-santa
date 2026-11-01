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

	var uids []uint32
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
			BuyerUserID:    uids[i],
			ReceiverUserID: uids[j],
		}
		lastYear.Pairing = append(lastYear.Pairing, p)
	}
	err = ss.AddYear(lastYear)
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
			BuyerUserID:    uids[i],
			ReceiverUserID: uids[j],
		}
		lastYear.Pairing = append(lastYear.Pairing, p)
	}
	err = ss.AddYear(lastYear)
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
	test := `┌─────────────┬───────────┐
│ Year        │ 2018      │
│ Amount      │ 15.5      │
│ Description │ Alles neu │
└─────────────┴───────────┘
┌──────┬───────────────────────────────────────┬───────────────────────────────────────┬────────┐
│ NO . │                 BUYER                 │               RECEIVER                │ RATING │
├──────┼───────────────────────────────────────┼───────────────────────────────────────┼────────┤
│ 0    │ Lastname_0,Firstname_0 (0@example.de) │ Lastname_3,Firstname_3 (3@example.de) │ 1      │
│ 1    │ Lastname_2,Firstname_2 (2@example.de) │ Lastname_1,Firstname_1 (1@example.de) │ 1      │
│ 2    │ Lastname_3,Firstname_3 (3@example.de) │ Lastname_2,Firstname_2 (2@example.de) │ 1      │
│ 3    │ Lastname_1,Firstname_1 (1@example.de) │ Lastname_0,Firstname_0 (0@example.de) │ 1      │
└──────┴───────────────────────────────────────┴───────────────────────────────────────┴────────┘
`
	if test != buf.String() {
		t.Errorf("failed to generate year: want %s, got %s,", test, buf.String())
	}

	var b strings.Builder
	err = ss.SendInformation(year, "{{.BuyerUser.Email}}->{{.ReceiverUser.Email}}", func(to, text string) error {
		b.WriteString(to)
		b.WriteString(":")
		b.WriteString(text)
		b.WriteString("\n")
		return nil
	})

	if err != nil {
		t.Errorf("failed to send information: %v", err)
	}

	test = `0@example.de:0@example.de->3@example.de
2@example.de:2@example.de->1@example.de
3@example.de:3@example.de->2@example.de
1@example.de:1@example.de->0@example.de
`

	if test != b.String() {
		t.Errorf("failed to generate year: want %s, got %s,", test, b.String())
	}
}
