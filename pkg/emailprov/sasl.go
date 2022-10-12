package emailprov

import "github.com/sqs/go-xoauth2"

type saslClient struct {
	email       string
	accessToken string
	tokenType   string
}

// Begins SASL authentication with the server. It returns the
// authentication mechanism name and "initial response" data (if required by
// the selected mechanism). A non-nil error causes the client to abort the
// authentication attempt.
//
// A nil ir value is different from a zero-length value. The nil value
// indicates that the selected mechanism does not use an initial response,
// while a zero-length value indicates an empty initial response, which must
// be sent to the server.
func (s saslClient) Start() (mech string, ir []byte, err error) {
	// returns unencoded string
	sasl := xoauth2.OAuth2String(s.email, s.accessToken)

	//logger.KV(xlog.DEBUG, "sasl", sasl)
	ir = []byte(sasl)
	mech = "XOAUTH2"
	return
}

// Continues challenge-response authentication. A non-nil error causes
// the client to abort the authentication attempt.
func (s saslClient) Next(challenge []byte) (response []byte, err error) {
	return
	//return nil, errors.Errorf("not supported")
}
