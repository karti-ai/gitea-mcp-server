package user

import (
	"testing"

	gitea_sdk "code.gitea.io/sdk/gitea"
)

func TestSlimUserDetail(t *testing.T) {
	u := &gitea_sdk.User{
		ID:        42,
		UserName:  "alice",
		FullName:  "Alice Smith",
		Email:     "alice@example.com",
		AvatarURL: "https://gitea.com/avatars/42",
		HTMLURL:   "https://gitea.com/alice",
		IsAdmin:   true,
	}
	m := slimUserDetail(u)

	if m["id"] != int64(42) {
		t.Errorf("expected id 42, got %v", m["id"])
	}
	if m["login"] != "alice" {
		t.Errorf("expected login alice, got %v", m["login"])
	}
	if m["full_name"] != "Alice Smith" {
		t.Errorf("expected full_name Alice Smith, got %v", m["full_name"])
	}
	if m["is_admin"] != true {
		t.Errorf("expected is_admin true, got %v", m["is_admin"])
	}
}

func TestSlimUserDetail_Nil(t *testing.T) {
	if m := slimUserDetail(nil); m != nil {
		t.Errorf("expected nil for nil user, got %v", m)
	}
}
