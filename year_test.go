package secretsanta

import (
	"os"
	"reflect"
	"testing"
)

func TestSecretSanta_YearPairingRoundtrip(t *testing.T) {
	ss, err := New("testdata/year_pairing_test.db")
	if err != nil {
		t.Fatal("failed to open database: test.db")
	}
	defer ss.Close()
	defer os.Remove("testdata/year_pairing_test.db")

	bid, err := ss.AddUser(&User{
		Email:     "buyer@test.de",
		Firstname: "B",
		Lastname:  "Buyer",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("failed to create buyer: %v", err)
	}
	rid, err := ss.AddUser(&User{
		Email:     "receiver@test.de",
		Firstname: "R",
		Lastname:  "Receiver",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	year := &Year{
		YearID:      2019,
		Description: "Pairing",
		Amount:      10.0,
		Pairing: []*Pairing{
			{BuyerUserID: bid, ReceiverUserID: rid, Rating: 0.5},
		},
	}
	if err := ss.AddYear(year); err != nil {
		t.Fatalf("failed to save year: %v", err)
	}

	loaded, err := ss.GetYear(2019)
	if err != nil {
		t.Fatalf("failed to load year: %v", err)
	}
	if len(loaded.Pairing) != 1 {
		t.Fatalf("expected 1 pairing, got %d", len(loaded.Pairing))
	}
	p := loaded.Pairing[0]
	if p.BuyerUserID != bid || p.ReceiverUserID != rid {
		t.Errorf("pairing swapped on load: buyer=%d (want %d), receiver=%d (want %d)",
			p.BuyerUserID, bid, p.ReceiverUserID, rid)
	}
	if p.Rating != 0.5 {
		t.Errorf("rating mismatch: got %v, want 0.5", p.Rating)
	}
}

func TestSecretSanta_Year(t *testing.T) {
	ss, err := New("testdata/year_test.db")
	if err != nil {
		t.Fatal("failed to open database: test.db")
	}
	defer ss.Close()
	defer os.Remove("testdata/year_test.db")

	year := &Year{
		Amount:      15.50,
		Description: "Description",
		YearID:      2018,
		Pairing:     []*Pairing{},
	}
	err = ss.AddYear(year)
	if err != nil {
		t.Error("failed to create year")
	}

	year2, err := ss.GetYear(year.YearID)
	if err != nil {
		t.Errorf("failed to load year: %v", err)
	}
	if !reflect.DeepEqual(year, year2) {
		t.Error("failed to read year")
	}

	year.Amount = 30
	err = ss.UpdateYear(year)
	if err != nil {
		t.Errorf("failed to update year: %v", err)
	}

	years, err := ss.ListYears()
	if err != nil {
		t.Error("failed to list years")
	}
	found := false
	for _, v := range years {
		found = found || reflect.DeepEqual(year, v)
	}
	if !found {
		t.Error("failed to list year")
	}
}
