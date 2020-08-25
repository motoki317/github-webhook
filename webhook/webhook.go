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
