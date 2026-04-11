package label

import (
	gitea_sdk "code.gitea.io/sdk/gitea"
)

func slimLabel(l *gitea_sdk.Label) map[string]any {
	if l == nil {
		return nil
	}
	return map[string]any{
		"id":          l.ID,
		"name":        l.Name,
		"color":       l.Color,
		"description": l.Description,
		"exclusive":   l.Exclusive,
	}
}

func slimLabels(labels []*gitea_sdk.Label) []map[string]any {
	out := make([]map[string]any, 0, len(labels))
	for _, l := range labels {
		out = append(out, slimLabel(l))
	}
	return out
}
