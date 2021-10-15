package secretsanta

import (
	"os"
	"reflect"
	"testing"
)

func TestSecretSant_User(t *testing.T) {
	ss, err := New("testdata/user_test.db")
	if err != nil {
		t.Fatal("failed to open database: test.db")
	}
	defer ss.Close()
	defer os.Remove("user_test.db")

	user := &User{
		Email:     "test@test.com",
		Firstname: "Vorname",
		Lastname:  "Nachname",
		Enabled:   true,
	}
	uid, err := ss.AddUser(user)
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}
	user2, err := ss.GetUser(uid)
	if err != nil {
		t.Errorf("failed to load user: %v", err)
	}
	if !reflect.DeepEqual(user, user2) {
		t.Error("failed user!=user2")
	}

	user.Enabled = false
	err = ss.UdateUser(user)
	if err != nil {
		t.Error("failed to update user")
	}

	users, err := ss.ListUsers(false)
	if err != nil {
		t.Error("failed to list user")
	}
	if len(users) != 0 {
		t.Error("user filter is not working")
	}
	users, err = ss.ListUsers(true)
	if err != nil {
		t.Error("failed to list user")
	}
	if len(users) != 1 {
		t.Error("user filter is not working")
	}
	found := false
	for _, v := range users {
		found = found || (user.ID == uid && reflect.DeepEqual(user, v))
	}
	if !found {
		t.Errorf("failed to list user: %v -> %v", user, users)
	}
}
