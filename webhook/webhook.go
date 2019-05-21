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
		event := c.Request().Header.Get("X-GitHub-Event")
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
	req, err := http.NewRequest("POST",
		"https://q.trap.jp/api/1.0/webhooks/"+os.Getenv("TRAQ_WEBHOOK_ID"),
		strings.NewReader(message))
	if err != nil {
		return err
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	req.Header.Set("X-TRAQ-Signature", generateSignature(message))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Message sent! %s", message)
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

	switch payload.Action {
	case "opened":
		return postMessage(c, "### Issue Opened")
	case "closed":
		return postMessage(c, "### Issue Closed")
	case "reopened":
		return postMessage(c, "### Issue Re-opened")
	default:
		return c.NoContent(http.StatusNoContent)
	}
}

func pushHandler(c echo.Context) error {
	payload := &model.PayloadPush{}
	if err := c.Bind(payload); err != nil {
		fmt.Println(err)
		return err
	}

	message := fmt.Sprintf("### [[%s](%s)] %v new ", payload.Repository.Name, payload.Repository.URL, len(payload.Commits))
	if len(payload.Commits) == 1 {
		message += " commit\n"
	} else {
		message += " commits\n"
	}
	for _, commit := range payload.Commits {
		message += fmt.Sprintf("[`%s`](%s) : %s - `%s` @ %s\n", commit.ID[:6], commit.URL, commit.Message, commit.Author.Name, commit.Timestamp.Format("2006/01/02 15:04:05"))
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
