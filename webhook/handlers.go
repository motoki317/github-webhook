package webhook

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/motoki317/github-webhook/icons"
	"github.com/motoki317/github-webhook/model"
	"net/http"
)

func issuesHandler(c echo.Context) error {
	payload := &model.PayloadIssue{}
	if err := c.Bind(payload); err != nil {
		fmt.Println(err)
		return err
	}

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
		return c.NoContent(http.StatusNoContent)
	}

	message := fmt.Sprintf(
		"### :%s: %s Issue %s by `%s`: [%s](%s)\n",
		icon,
		buildRepositoryBase(payload.Repository),
		payload.Action,
		payload.Sender.Login,
		payload.Issue.Title,
		payload.Issue.HTMLURL)

	if payload.Action == "opened" {
		message += "\n---\n"
		message += payload.Issue.Body
	}

	return postMessage(c, message)
}

func pushHandler(c echo.Context) error {
	payload := &model.PayloadPush{}
	if err := c.Bind(payload); err != nil {
		fmt.Println(err)
		return err
	}

	if len(payload.Commits) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	message := fmt.Sprintf(
		"### :%s: [[%s](%s)] %v new",
		icons.Pushed,
		payload.Repository.Name,
		payload.Repository.HTMLURL,
		len(payload.Commits))

	if len(payload.Commits) == 1 {
		message += " commit"
	} else {
		message += " commits"
	}
	message += fmt.Sprintf(" to %s\n", payload.Ref)
	message += "\n---\n"

	for _, commit := range payload.Commits {
		message += fmt.Sprintf(":0x%s: [`%s`](%s) : %s - `%s` @ %s\n", commit.ID[:6], commit.ID[:6], commit.URL, commit.Message, commit.Author.Name, commit.Timestamp.Format("2006/01/02 15:04:05"))
	}

	return postMessage(c, message)
}

func pullRequestHandler(c echo.Context) error {
	payload := &model.PayloadPullRequest{}
	if err := c.Bind(payload); err != nil {
		fmt.Println(err)
		return err
	}

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
		return c.NoContent(http.StatusNoContent)
	}

	message := fmt.Sprintf(
		"### :%s: %s Pull Request %s by `%s`: [%s](%s)\n",
		icon,
		buildRepositoryBase(payload.Repository),
		action,
		payload.Sender.Login,
		payload.PullRequest.Title,
		payload.PullRequest.HTMLURL)

	// send pull request body only on the first open
	if payload.Action == "opened" {
		message += "\n---\n"
		message += payload.PullRequest.Body
	}

	return postMessage(c, message)
}

// buildRepositoryBase Repositoryのベースメッセージを作成します
// 例: [[github-webhook](URL)]
func buildRepositoryBase(repo model.Repository) string {
	return fmt.Sprintf("[[%s](%s)]", repo.Name, repo.HTMLURL)
}
