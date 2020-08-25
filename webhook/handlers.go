package webhook

import (
	"fmt"
	"github.com/motoki317/github-webhook/icons"
	"gopkg.in/go-playground/webhooks.v5/github"
	"strings"
	"time"
)

func issuesHandler(payload github.IssuesPayload) error {
	var icon string
	switch payload.Action {
	case "opened":
		icon = icons.IssueOpened
	case "deleted":
		icon = icons.IssueClosed
	case "closed":
		icon = icons.IssueClosed
	case "reopened":
		icon = icons.IssueOpened
	default:
		return nil
	}

	issueName := fmt.Sprintf("[#%d %s](%s)", payload.Issue.Number, payload.Issue.Title, payload.Issue.HTMLURL)
	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] Issue %s %s by `%s`\n",
		icon,
		payload.Repository.Name, payload.Repository.HTMLURL,
		issueName,
		strings.Title(payload.Action),
		payload.Sender.Login)

	if payload.Action == "opened" {
		message += "\n---\n"
		message += payload.Issue.Body
	}

	return nil
}

func issueCommentHandler(payload github.IssueCommentPayload) error {
	issueName := fmt.Sprintf("[#%d %s](%s)", payload.Issue.Number, payload.Issue.Title, payload.Issue.HTMLURL)
	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] [Issue Comment](%s) %s by `%s`\n"+
			"%s\n",
		icons.Comment,
		payload.Repository.Name, payload.Repository.HTMLURL,
		payload.Comment.HTMLURL,
		strings.Title(payload.Action),
		payload.Sender.Login,
		issueName)

	if payload.Action == "created" {
		message += "\n---\n"
		message += payload.Comment.Body
	}

	return postMessage(message)
}

func pushHandler(payload github.PushPayload) error {
	if len(payload.Commits) == 0 {
		return nil
	}

	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] %v New",
		icons.Pushed,
		payload.Repository.Name, payload.Repository.HTMLURL,
		len(payload.Commits))

	if len(payload.Commits) == 1 {
		message += " Commit"
	} else {
		message += " Commits"
	}
	message += fmt.Sprintf(
		" to `%s` by `%s`\n",
		payload.Ref,
		payload.Sender.Login)
	message += "\n---\n"

	for _, commit := range payload.Commits {
		formattedTime, err := formatTimeISO8601(commit.Timestamp, "2006/01/02 15:04:05")
		if err != nil {
			return err
		}
		message += fmt.Sprintf(
			":0x%s: [`%s`](%s) : %s - `%s` @ %s\n",
			commit.ID[:6], commit.ID[:6],
			commit.URL,
			commit.Message,
			commit.Author.Name,
			formattedTime)
	}

	return nil
}

func pullRequestHandler(payload github.PullRequestPayload) error {
	// If action == "closed" and Merged is true, then the pull request was merged
	var action string
	var icon string
	switch payload.Action {
	case "opened":
		action = payload.Action
		icon = icons.PullRequestOpened
	case "closed":
		if payload.PullRequest.Merged {
			action = "merged"
			icon = icons.PullRequestMerged
		} else {
			action = "closed"
			icon = icons.PullRequestClosed
		}
	case "reopened":
		action = payload.Action
		icon = icons.PullRequestOpened
	default:
		return nil
	}

	prName := fmt.Sprintf("[#%d %s](%s)", payload.PullRequest.Number, payload.PullRequest.Title, payload.PullRequest.HTMLURL)
	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] Pull Request %s %s by `%s`\n",
		icon,
		payload.Repository.Name, payload.Repository.HTMLURL,
		prName,
		strings.Title(action),
		payload.Sender.Login)

	// send pull request body only on the first open
	if payload.Action == "opened" {
		message += "\n---\n"
		message += payload.PullRequest.Body
	}

	return postMessage(message)
}

func pullRequestReviewHandler(payload github.PullRequestReviewPayload) error {
	if payload.Action != "submitted" {
		return nil
	}

	var action string
	var icon string
	switch payload.Review.State {
	case "approved":
		action = "approved"
		icon = icons.PullRequestApproved
	case "commented":
		action = "commented"
		icon = icons.Comment
	case "changes_requested":
		action = "changes requested"
		icon = icons.Comment
	default:
		return nil
	}

	prName := fmt.Sprintf("[#%d %s](%s)", payload.PullRequest.Number, payload.PullRequest.Title, payload.PullRequest.HTMLURL)
	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] Pull Request %s %s by `%s`\n"+
			"\n"+
			"---\n"+
			"%s",
		icon,
		payload.Repository.Name, payload.Repository.HTMLURL,
		prName,
		strings.Title(action),
		payload.Sender.Login,
		payload.Review.Body)

	return postMessage(message)
}

func formatTimeISO8601(from string, format string) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05Z", from)
	if err != nil {
		return "", err
	}
	return t.Format(format), nil
}
