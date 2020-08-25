package webhook

import (
	"fmt"
	"github.com/motoki317/github-webhook/icons"
	"gopkg.in/go-playground/webhooks.v5/github"
	"strings"
)

func issuesHandler(payload github.IssuesPayload) error {
	var icon string
	switch payload.Action {
	case "opened":
		icon = icons.IssueOpened
	case "edited":
		icon = icons.Edit
	case "deleted":
		icon = icons.IssueClosed
	case "closed":
		icon = icons.IssueClosed
	case "reopened":
		icon = icons.IssueOpened
	case "pinned":
		icon = icons.Pin
	case "unpinned":
		icon = icons.Pin
	case "labeled":
		icon = icons.Tag
	case "unlabeled":
		icon = icons.Tag
	case "locked":
		icon = icons.Lock
	case "unlocked":
		icon = icons.Unlock
	case "transferred":
		icon = icons.Transfer
	case "milestoned":
		icon = icons.Milestone
	case "demilestoned":
		icon = icons.Milestone
	case "assigned":
		icon = icons.Assignment
	case "unassigned":
		icon = icons.Assignment
	default:
		return nil
	}

	issueName := fmt.Sprintf("[#%d %s](%s)", payload.Issue.Number, payload.Issue.Title, payload.Issue.HTMLURL)
	var message string
	switch payload.Action {
	case "assigned":
		fallthrough
	case "unassigned":
		message = fmt.Sprintf(
			"### :%s: [[%s](%s)] Issue %s %s to `%s` by `%s`\n",
			icon,
			payload.Repository.Name, payload.Repository.HTMLURL,
			issueName,
			strings.Title(payload.Action),
			payload.Assignee.Login,
			payload.Sender.Login)
	default:
		message = fmt.Sprintf(
			"### :%s: [[%s](%s)] Issue %s %s by `%s`\n",
			icon,
			payload.Repository.Name, payload.Repository.HTMLURL,
			issueName,
			strings.Title(payload.Action),
			payload.Sender.Login)
	}

	if assignees := getAssigneeNames(payload); assignees != "" {
		message += "Assignees: " + assignees + "\n"
	}
	if labels := getLabelNames(payload); labels != "" {
		message += "Labels: " + labels + "\n"
	}

	if payload.Action == "opened" || payload.Action == "edited" {
		message += "\n---\n"
		message += payload.Issue.Body
	}

	return postMessage(message)
}

func issueCommentHandler(payload github.IssueCommentPayload) error {
	var icon string
	switch payload.Action {
	case "created":
		icon = icons.Comment
	case "edited":
		icon = icons.Edit
	case "deleted":
		icon = icons.Retrieved
	default:
		return nil
	}

	issueName := fmt.Sprintf("[#%d %s](%s)", payload.Issue.Number, payload.Issue.Title, payload.Issue.HTMLURL)
	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] [Comment](%s) %s by `%s`\n"+
			"%s\n",
		icon,
		payload.Repository.Name, payload.Repository.HTMLURL,
		payload.Comment.HTMLURL,
		strings.Title(payload.Action),
		payload.Sender.Login,
		issueName)

	if assignees := getAssigneeNames(payload); assignees != "" {
		message += "Assignees: " + assignees + "\n"
	}
	if labels := getLabelNames(payload); labels != "" {
		message += "Labels: " + labels + "\n"
	}

	if payload.Action == "created" || payload.Action == "edited" {
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
		formattedTime, err := formatTime(commit.Timestamp, "2006/01/02 15:04:05")
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

	return postMessage(message)
}

func pullRequestHandler(payload github.PullRequestPayload) error {
	// If action == "closed" and Merged is true, then the pull request was merged
	action := payload.Action
	var icon string
	switch payload.Action {
	case "opened":
		icon = icons.PullRequestOpened
	case "edited":
		icon = icons.Edit
	case "closed":
		if payload.PullRequest.Merged {
			action = "merged"
			icon = icons.PullRequestMerged
		} else {
			action = "closed"
			icon = icons.PullRequestClosed
		}
	case "reopened":
		icon = icons.PullRequestOpened
	case "assigned":
		icon = icons.Assignment
	case "unassigned":
		icon = icons.Assignment
	case "review_requested":
		action = "review requested"
		icon = icons.Assignment
	case "review_request_removed":
		action = "review request removed"
		icon = icons.Assignment
	case "ready_for_review":
		action = "marked as ready for review"
		icon = icons.Assignment
	case "labeled":
		icon = icons.Tag
	case "unlabeled":
		icon = icons.Tag
	// case "synchronize": on push event
	case "locked":
		icon = icons.Lock
	case "unlocked":
		icon = icons.Unlock
	default:
		return nil
	}

	prName := fmt.Sprintf("[#%d %s](%s)", payload.PullRequest.Number, payload.PullRequest.Title, payload.PullRequest.HTMLURL)

	var message string
	switch payload.Action {
	case "assigned":
		fallthrough
	case "unassigned":
		message = fmt.Sprintf(
			"### :%s: [[%s](%s)] Pull Request %s %s to `%s` by `%s`\n",
			icon,
			payload.Repository.Name, payload.Repository.HTMLURL,
			prName,
			strings.Title(action),
			payload.Assignee.Login,
			payload.Sender.Login)
	case "review_requested":
		message = fmt.Sprintf(
			"### :%s: [[%s](%s)] Pull Request %s %s to `%s` by `%s`\n",
			icon,
			payload.Repository.Name, payload.Repository.HTMLURL,
			prName,
			strings.Title(action),
			payload.RequestedReviewer.Login,
			payload.Sender.Login)
	default:
		message = fmt.Sprintf(
			"### :%s: [[%s](%s)] Pull Request %s %s by `%s`\n",
			icon,
			payload.Repository.Name, payload.Repository.HTMLURL,
			prName,
			strings.Title(action),
			payload.Sender.Login)
	}

	if assignees := getAssigneeNames(payload); assignees != "" {
		message += "Assignees: " + assignees + "\n"
	}
	if reviewers := getRequestedReviewers(payload); reviewers != "" {
		message += "Reviewers: " + reviewers + "\n"
	}
	if labels := getLabelNames(payload); labels != "" {
		message += "Labels: " + labels + "\n"
	}

	// send pull request body only on the first open or on edited
	if payload.Action == "opened" || payload.Action == "edited" {
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
		"### :%s: [[%s](%s)] Pull Request %s %s by `%s`\n",
		icon,
		payload.Repository.Name, payload.Repository.HTMLURL,
		prName,
		strings.Title(action),
		payload.Sender.Login)

	if assignees := getAssigneeNames(payload); assignees != "" {
		message += "Assignees: " + assignees + "\n"
	}

	message += "\n---\n"
	message += payload.Review.Body

	return postMessage(message)
}

func pullRequestReviewCommentHandler(payload github.PullRequestReviewCommentPayload) error {
	var icon string
	switch payload.Action {
	case "created":
		icon = icons.Comment
	case "edited":
		icon = icons.Edit
	case "deleted":
		icon = icons.Retrieved
	default:
		return nil
	}

	prName := fmt.Sprintf("[#%d %s](%s)", payload.PullRequest.Number, payload.PullRequest.Title, payload.PullRequest.HTMLURL)
	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] [Review Comment](%s) %s by `%s`\n"+
			"%s\n",
		icon,
		payload.Repository.Name, payload.Repository.HTMLURL,
		payload.Comment.HTMLURL,
		strings.Title(payload.Action),
		payload.Sender.Login,
		prName)

	if assignees := getAssigneeNames(payload); assignees != "" {
		message += "Assignees: " + assignees + "\n"
	}

	message += "\n---\n"
	message += payload.Comment.Body

	return postMessage(message)
}
