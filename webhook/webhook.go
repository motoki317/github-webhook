package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// MakeWebhookHandler WebhookHandlerを返します
func MakeWebhookHandler(githubSecret string) func(c echo.Context) error {
	hook, _ := github.New(github.Options.Secret(githubSecret))

	return func(c echo.Context) error {
		payload, err := hook.Parse(c.Request(), github.ReleaseEvent, github.PullRequestEvent)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		go func() {
			var payloadType string
			var err error
			switch payload.(type) {
			case github.IssuesPayload:
				payloadType = "issues"
				err = issuesHandler(payload.(github.IssuesPayload))
			case github.IssueCommentPayload:
				payloadType = "issue comment"
				err = issueCommentHandler(payload.(github.IssueCommentPayload))
			case github.PushPayload:
				payloadType = "push"
				err = pushHandler(payload.(github.PushPayload))
			case github.PullRequestPayload:
				payloadType = "pull request"
				err = pullRequestHandler(payload.(github.PullRequestPayload))
			}

			log.Printf("Received event %s\n", payloadType)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
			}
		}()

		return c.NoContent(http.StatusNoContent)
	}
}

// postMessage Webhookにメッセージを投稿します
func postMessage(message string) error {
	url := "https://q.trap.jp/api/v3/webhooks/" + os.Getenv("TRAQ_WEBHOOK_ID")
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
	defer func() {
		_ = resp.Body.Close()
	}()

	response := make([]byte, 512)
	_, _ = resp.Body.Read(response)

	fmt.Printf("Message sent to %s, message: %s, response: %s\n", url, message, response)
	return nil
}

func generateSignature(message string) string {
	mac := hmac.New(sha1.New, []byte(os.Getenv("TRAQ_WEBHOOK_SECRET")))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
