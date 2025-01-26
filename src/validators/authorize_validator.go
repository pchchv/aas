package validators

type ValidateClientAndRedirectURIInput struct {
	RequestId   string
	ClientId    string
	RedirectURI string
}

type ValidateRequestInput struct {
	ResponseType        string
	ResponseMode        string
	CodeChallenge       string
	CodeChallengeMethod string
}
