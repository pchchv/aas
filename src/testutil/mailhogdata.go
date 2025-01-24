package testutil

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
