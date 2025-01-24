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
