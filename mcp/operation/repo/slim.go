package repo

import (
	gitea_sdk "code.gitea.io/sdk/gitea"
)

func userLogin(u *gitea_sdk.User) string {
	if u == nil {
		return ""
	}
	return u.UserName
}

func slimRepo(r *gitea_sdk.Repository) map[string]any {
	if r == nil {
		return nil
	}
	m := map[string]any{
		"id":                r.ID,
		"full_name":         r.FullName,
		"description":       r.Description,
		"html_url":          r.HTMLURL,
		"clone_url":         r.CloneURL,
		"ssh_url":           r.SSHURL,
		"default_branch":    r.DefaultBranch,
		"private":           r.Private,
		"fork":              r.Fork,
		"archived":          r.Archived,
		"language":          r.Language,
		"stars_count":       r.Stars,
		"forks_count":       r.Forks,
		"open_issues_count": r.OpenIssues,
		"open_pr_counter":   r.OpenPulls,
		"created_at":        r.Created,
		"updated_at":        r.Updated,
	}
	if r.Owner != nil {
		m["owner"] = r.Owner.UserName
	}
	if len(r.Topics) > 0 {
		m["topics"] = r.Topics
	}
	return m
}

func slimRepos(repos []*gitea_sdk.Repository) []map[string]any {
	out := make([]map[string]any, 0, len(repos))
	for _, r := range repos {
		out = append(out, slimRepo(r))
	}
	return out
}

func slimBranch(b *gitea_sdk.Branch) map[string]any {
	if b == nil {
		return nil
	}
	m := map[string]any{
		"name":      b.Name,
		"protected": b.Protected,
	}
	if b.Commit != nil {
		m["commit_sha"] = b.Commit.ID
	}
	return m
}

func slimBranches(branches []*gitea_sdk.Branch) []map[string]any {
	out := make([]map[string]any, 0, len(branches))
	for _, b := range branches {
		out = append(out, slimBranch(b))
	}
	return out
}

func slimCommit(c *gitea_sdk.Commit) map[string]any {
	if c == nil {
		return nil
	}
	m := map[string]any{
		"sha":      c.SHA,
		"html_url": c.HTMLURL,
		"created":  c.Created,
	}
	if c.RepoCommit != nil {
		m["message"] = c.RepoCommit.Message
		if c.RepoCommit.Author != nil {
			m["author"] = map[string]any{
				"name":  c.RepoCommit.Author.Name,
				"email": c.RepoCommit.Author.Email,
				"date":  c.RepoCommit.Author.Date,
			}
		}
	}
	return m
}

func slimCommits(commits []*gitea_sdk.Commit) []map[string]any {
	out := make([]map[string]any, 0, len(commits))
	for _, c := range commits {
		out = append(out, slimCommit(c))
	}
	return out
}

func slimTag(t *gitea_sdk.Tag) map[string]any {
	if t == nil {
		return nil
	}
	m := map[string]any{
		"name":    t.Name,
		"message": t.Message,
	}
	if t.Commit != nil {
		m["commit_sha"] = t.Commit.SHA
	}
	return m
}

func slimTags(tags []*gitea_sdk.Tag) []map[string]any {
	out := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		m := map[string]any{
			"name": t.Name,
		}
		if t.Commit != nil {
			m["commit_sha"] = t.Commit.SHA
		}
		out = append(out, m)
	}
	return out
}

func slimRelease(r *gitea_sdk.Release) map[string]any {
	if r == nil {
		return nil
	}
	return map[string]any{
		"id":           r.ID,
		"tag_name":     r.TagName,
		"target":       r.Target,
		"title":        r.Title,
		"body":         r.Note,
		"draft":        r.IsDraft,
		"prerelease":   r.IsPrerelease,
		"html_url":     r.HTMLURL,
		"author":       userLogin(r.Publisher),
		"created_at":   r.CreatedAt,
		"published_at": r.PublishedAt,
	}
}

func slimReleases(releases []*gitea_sdk.Release) []map[string]any {
	out := make([]map[string]any, 0, len(releases))
	for _, r := range releases {
		out = append(out, slimRelease(r))
	}
	return out
}

func slimContents(c *gitea_sdk.ContentsResponse) map[string]any {
	if c == nil {
		return nil
	}
	m := map[string]any{
		"name": c.Name,
		"path": c.Path,
		"sha":  c.SHA,
		"type": c.Type,
		"size": c.Size,
	}
	if c.Content != nil {
		m["content"] = *c.Content
	}
	if c.Encoding != nil {
		m["encoding"] = *c.Encoding
	}
	if c.HTMLURL != nil {
		m["html_url"] = *c.HTMLURL
	}
	if c.DownloadURL != nil {
		m["download_url"] = *c.DownloadURL
	}
	return m
}

func slimDirEntries(entries []*gitea_sdk.ContentsResponse) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, c := range entries {
		if c == nil {
			continue
		}
		out = append(out, map[string]any{
			"name": c.Name,
			"path": c.Path,
			"type": c.Type,
			"size": c.Size,
		})
	}
	return out
}
