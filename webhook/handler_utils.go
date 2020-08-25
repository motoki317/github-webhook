package webhook

import (
	"gopkg.in/go-playground/webhooks.v5/github"
	"strings"
	"time"
)

func formatTime(from string, format string) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", from)
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}

func getAssigneeNames(payload interface{}) (ret string) {
	var assignees []*github.Assignee
	switch payload.(type) {
	case github.IssuesPayload:
		payload := payload.(github.IssuesPayload)
		assignees = payload.Issue.Assignees
	case github.IssueCommentPayload:
		payload := payload.(github.IssueCommentPayload)
		assignees = payload.Issue.Assignees
	case github.PullRequestPayload:
		payload := payload.(github.PullRequestPayload)
		assignees = payload.PullRequest.Assignees
	case github.PullRequestReviewPayload:
		payload := payload.(github.PullRequestReviewPayload)
		// not a slice of pointers
		reviewAssignees := payload.PullRequest.Assignees
		assignees = make([]*github.Assignee, 0, len(reviewAssignees))
		for _, assignee := range reviewAssignees {
			assignee := assignee
			assignees = append(assignees, &assignee)
		}
	case github.PullRequestReviewCommentPayload:
		payload := payload.(github.PullRequestReviewCommentPayload)
		assignees = payload.PullRequest.Assignees
	default:
		return
	}

	if assignees == nil {
		return
	}

	formatted := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		formatted = append(formatted, "`"+assignee.Login+"`")
	}
	return strings.Join(formatted, ", ")
}

func getRequestedReviewers(payload github.PullRequestPayload) string {
	formatted := make([]string, 0, len(payload.PullRequest.RequestedReviewers))
	for _, reviewer := range payload.PullRequest.RequestedReviewers {
		formatted = append(formatted, "`"+reviewer.Login+"`")
	}
	return strings.Join(formatted, ", ")
}

func getLabelNames(payload interface{}) (ret string) {
	// gopkg.in/go-playground/webhooks.v5/github/payload.go
	var labels []struct {
		ID          int64  `json:"id"`
		NodeID      string `json:"node_id"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		Color       string `json:"color"`
		Default     bool   `json:"default"`
	}
	switch payload.(type) {
	case github.IssuesPayload:
		payload := payload.(github.IssuesPayload)
		labels = payload.Issue.Labels
	case github.IssueCommentPayload:
		payload := payload.(github.IssueCommentPayload)
		labels = payload.Issue.Labels
	case github.PullRequestPayload:
		payload := payload.(github.PullRequestPayload)
		labels = payload.PullRequest.Labels
	// case github.PullRequestReviewPayload:
	// case github.PullRequestReviewCommentPayload:
	// no labels
	default:
		return
	}

	if labels == nil {
		return
	}

	formatted := make([]string, 0, len(labels))
	for _, label := range labels {
		formatted = append(formatted, ":0x"+label.Color+": `"+label.Name+"`")
	}
	return strings.Join(formatted, ", ")
}
