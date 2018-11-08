package secretsanta

import (
	"os"
	"reflect"
	"testing"
)

func TestSecretSanta_Year(t *testing.T) {
	ss, err := New("testdata/year_test.db")
	if err != nil {
		t.Fatal("failed to open database: test.db")
	}
	defer ss.Close()
	defer os.Remove("year_test.db")

	year := &Year{
		Amount:      15.50,
		Description: "Description",
		YearID:      2018,
	}
	uid, err := ss.AddYear(year)
	if err != nil {
		t.Error("failed to create year")
	}

	year2, err := ss.GetYear(uid)
	if err != nil {
		t.Errorf("failed to load year: %v", err)
	}
	if !reflect.DeepEqual(year, year2) {
		t.Error("failed to read year")
	}

	year.Amount = 30
	err = ss.UdateYear(year)
	if err != nil {
		t.Errorf("failed to update year: %v", err)
	}

	years, err := ss.ListYears()
	if err != nil {
		t.Error("failed to list years")
	}
	found := true
	for _, v := range years {
		found = found || reflect.DeepEqual(year, v)
	}
	if !found {
		t.Error("failed to list year")
	}
}
