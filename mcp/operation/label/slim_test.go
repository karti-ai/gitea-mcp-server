package label

import (
	"testing"

	gitea_sdk "code.gitea.io/sdk/gitea"
)

func TestSlimLabel(t *testing.T) {
	l := &gitea_sdk.Label{
		ID:          1,
		Name:        "bug",
		Color:       "#d73a4a",
		Description: "Something isn't working",
		Exclusive:   false,
	}

	m := slimLabel(l)
	if m["name"] != "bug" {
		t.Errorf("expected name bug, got %v", m["name"])
	}
	if m["color"] != "#d73a4a" {
		t.Errorf("expected color, got %v", m["color"])
	}
}
