package testutil

import "time"

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
