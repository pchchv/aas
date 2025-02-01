package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type From struct {
	Relays  any    `json:"Relays"`
	Mailbox string `json:"Mailbox"`
	Domain  string `json:"Domain"`
	Params  string `json:"Params"`
}

type To []struct {
	Relays  any    `json:"Relays"`
	Mailbox string `json:"Mailbox"`
	Domain  string `json:"Domain"`
	Params  string `json:"Params"`
}

type Headers struct {
	ContentTransferEncoding []string `json:"Content-Transfer-Encoding"`
	ContentType             []string `json:"Content-Type"`
	Date                    []string `json:"Date"`
	From                    []string `json:"From"`
	MIMEVersion             []string `json:"MIME-Version"`
	MessageID               []string `json:"Message-ID"`
	Received                []string `json:"Received"`
	ReturnPath              []string `json:"Return-Path"`
	Subject                 []string `json:"Subject"`
	To                      []string `json:"To"`
}

type Content struct {
	Headers Headers `json:"Headers"`
	Body    string  `json:"Body"`
	Size    int     `json:"Size"`
	Mime    any     `json:"MIME"`
}

type Raw struct {
	From string   `json:"From"`
	To   []string `json:"To"`
	Data string   `json:"Data"`
	Helo string   `json:"Helo"`
}

type Item struct {
	ID      string    `json:"ID"`
	From    From      `json:"From"`
	To      To        `json:"To"`
	Content Content   `json:"Content"`
	Created time.Time `json:"Created"`
	Mime    any       `json:"MIME"`
	Raw     Raw       `json:"Raw"`
}

type MailhogData struct {
	Total int    `json:"total"`
	Count int    `json:"count"`
	Start int    `json:"start"`
	Items []Item `json:"items"`
}

func AssertEmailSent(t *testing.T, to string, containing string) {
	destUrl := "http://mailhog:8025/api/v2/search?kind=to&query=" + to
	resp, err := http.Get(destUrl)
	require.NoError(t, err, "Failed to send GET request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var mailhogData MailhogData
	err = json.Unmarshal(body, &mailhogData)
	require.NoError(t, err, "Failed to unmarshal JSON")
	assert.Equal(t, 1, len(mailhogData.Items), "expecting to find 1 email")
	if len(mailhogData.Items) > 0 {
		assert.True(t, strings.Contains(mailhogData.Items[0].Content.Headers.To[0], to))
		assert.True(t, strings.Contains(mailhogData.Items[0].Content.Body, containing))
	} else {
		t.Errorf("No emails found for recipient: %s", to)
	}
}
