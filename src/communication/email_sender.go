package communication

type SendEmailInput struct {
	To       string
	Subject  string
	HtmlBody string
}

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}
