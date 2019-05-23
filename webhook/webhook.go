package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/motoki317/github-webhook/model"
)

// MakeWebhookHandler WebhookHandlerを返します
func MakeWebhookHandler() func(c echo.Context) error {
	return func(c echo.Context) error {
		event := c.Request().Header.Get("x-github-event")
		fmt.Printf("Received %s event\n", event)
		switch event {
		case "":
			return c.NoContent(http.StatusBadRequest)
		case "issues":
			return issuesHandler(c)
		case "push":
			return pushHandler(c)
		case "pull_request":
			return pullRequestHandler(c)
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// postMessage Webhookにメッセージを投稿します
func postMessage(c echo.Context, message string) error {
	url := "https://q.trap.jp/api/1.0/webhooks/" + os.Getenv("TRAQ_WEBHOOK_ID")
	req, err := http.NewRequest("POST",
		url,
		strings.NewReader(message))
	if err != nil {
		return err
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	req.Header.Set("X-TRAQ-Signature", generateSignature(message))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response := make([]byte, 512)
	resp.Body.Read(response)

	fmt.Printf("Message sent to %s, message: %s, response: %s\n", url, message, response)

	return c.NoContent(http.StatusNoContent)
}

func generateSignature(message string) string {
	mac := hmac.New(sha1.New, []byte(os.Getenv("TRAQ_WEBHOOK_SECRET")))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func issuesHandler(c echo.Context) error {
	payload := &model.PayloadIssue{}
	if err := c.Bind(payload); err != nil {
		fmt.Println(err)
		return err
	}

	message := fmt.Sprintf("### %s Issue %s: [%s](%s)\n", buildRepositoryBase(payload.Repository), payload.Action, payload.Issue.Title, payload.Issue.URL)
	switch payload.Action {
	case "opened":
		message += payload.Issue.Body
	case "deleted":
	case "closed":
	case "reopened":
	default:
		return nil
	}

	return postMessage(c, message)
}

func pushHandler(c echo.Context) error {
	payload := &model.PayloadPush{}
	if err := c.Bind(payload); err != nil {
		fmt.Println(err)
		return err
	}

	message := fmt.Sprintf("### %s %v new", buildRepositoryBase(payload.Repository), len(payload.Commits))
	if len(payload.Commits) == 1 {
		message += " commit"
	} else {
		message += " commits"
	}
	message += fmt.Sprintf(" to %s\n", payload.Ref)
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

	switch payload.Action {
	default:
		return c.NoContent(http.StatusNoContent)
	}
}

// buildRepositoryBase Repositoryのベースメッセージを作成します
// 例: [[github-webhook](URL)]
func buildRepositoryBase(repo mode.Repository) string {
	return fmt.Sprintf("### [[%s](%s)]", repo.Name, repo.URL)
}
