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
		payload, err := hook.Parse(c.Request(),
			github.IssuesEvent,
			github.IssueCommentEvent,
			github.PushEvent,
			github.PullRequestEvent,
			github.PullRequestReviewEvent,
			github.PullRequestReviewCommentEvent)
		if err != nil {
			log.Println("Received invalid payload")
			return c.NoContent(http.StatusBadRequest)
		}

		go func() {
			var payloadType string
			var action string
			var err error
			switch payload.(type) {
			case github.IssuesPayload:
				payloadType = "issues"
				payload := payload.(github.IssuesPayload)
				action = payload.Action
				err = issuesHandler(payload)
			case github.IssueCommentPayload:
				payloadType = "issue comment"
				payload := payload.(github.IssueCommentPayload)
				action = payload.Action
				err = issueCommentHandler(payload)
			case github.PushPayload:
				payloadType = "push"
				payload := payload.(github.PushPayload)
				action = "push"
				err = pushHandler(payload)
			case github.PullRequestPayload:
				payloadType = "pull request"
				payload := payload.(github.PullRequestPayload)
				action = payload.Action
				err = pullRequestHandler(payload)
			case github.PullRequestReviewPayload:
				payloadType = "pull request review"
				payload := payload.(github.PullRequestReviewPayload)
				action = payload.Action
				err = pullRequestReviewHandler(payload)
			case github.PullRequestReviewCommentPayload:
				payloadType = "pull request review comment"
				payload := payload.(github.PullRequestReviewCommentPayload)
				action = payload.Action
				err = pullRequestReviewCommentHandler(payload)
			}

			log.Printf("Received event %s, action %s\n", payloadType, action)
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
