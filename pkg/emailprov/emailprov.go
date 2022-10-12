package emailprov

import (
	"time"

	"github.com/effective-security/xlog"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/ableorg/email-sync", "emailprov")

// WithTimeout allows to specify timeout
func WithTimeout(timeout time.Duration) Option {
	return newFuncOption(func(o *options) {
		o.timeout = timeout
	})
}

type Provider struct {
	dops   options
	server string
	email  string
	client *client.Client
}

// New returns Provider
func New(server, email, accessToken string, ops ...Option) (*Provider, error) {
	p := &Provider{
		dops:   options{},
		server: server,
		email:  email,
	}
	for _, op := range ops {
		op.apply(&p.dops)
	}

	// Connect to server
	logger.KV(xlog.DEBUG, "connecting", p.server)

	c, err := client.DialTLS(p.server, nil)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to dial: %s", p.server)
	}
	if p.dops.timeout > 0 {
		c.Timeout = p.dops.timeout
	}
	p.client = c

	sc := saslClient{
		email:       email,
		accessToken: accessToken,
		tokenType:   "Bearer",
	}

	logger.KV(xlog.DEBUG, "authenticating", email)
	err = p.client.Authenticate(sc)
	if err != nil {
		return nil, errors.WithMessagef(err, "authentication failed")
	}

	return p, nil
}

// Close the connection
func (p *Provider) Close() error {
	if p.client != nil {
		err := p.client.Logout()
		p.client = nil
		return errors.WithStack(err)
	}
	return nil
}

// Mailboxes returns mailboxes
func (p *Provider) Mailboxes() ([]*imap.MailboxInfo, error) {
	logger.KV(xlog.DEBUG, "list", "Mailboxes")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- p.client.List("", "*", mailboxes)
	}()

	list := make([]*imap.MailboxInfo, 0, 10)
	for m := range mailboxes {
		list = append(list, m)
	}
	if err := <-done; err != nil {
		return nil, errors.WithMessage(err, "failed to list mailboxes")
	}

	return list, nil
}

// Messages returns messages
func (p *Provider) Messages(mailbox string, from, limit uint32) ([]*imap.Message, error) {
	mbox, err := p.client.Select(mailbox, true)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to select mailbox: %s", mailbox)
	}
	logger.KV(xlog.DEBUG, "mailbox", mailbox, "messages", mbox.Messages)

	// Get the last messages
	if from > mbox.Messages {
		return nil, nil
	}
	from = mbox.Messages - from
	to := from + limit
	if to > mbox.Messages {
		// We're using unsigned integers here, only subtract if the result is > 0
		to = mbox.Messages
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- p.client.Fetch(seqset, imap.FetchAll.Expand(), messages)
	}()

	list := make([]*imap.Message, 0, limit)
	for msg := range messages {
		list = append(list, msg)
	}

	if err := <-done; err != nil {
		return nil, errors.WithMessage(err, "failed to list messages")
	}

	return list, err
}

// GetMessage returns message by ID
func (p *Provider) GetMessage(mailbox string, id string) (*imap.Message, error) {
	_, err := p.client.Select(mailbox, true)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to select mailbox: %s", mailbox)
	}

	// Set search criteria
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	criteria.Header.Set("Message-ID", id)

	ids, err := p.client.Search(criteria)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to find message: %s", id)
	}

	if len(ids) == 0 {
		return nil, errors.Errorf("message not found: %s", id)
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- p.client.Fetch(seqset, imap.FetchFull.Expand(), messages)
	}()

	var res *imap.Message
	for msg := range messages {
		res = msg
		break
	}

	if err := <-done; err != nil {
		return nil, errors.WithMessagef(err, "failed to get message: %s", id)
	}

	return res, err
}

// LabelMessage sets message label
func (p *Provider) LabelMessage(mailbox string, id, label string) error {
	_, err := p.client.Select(mailbox, true)
	if err != nil {
		return errors.WithMessagef(err, "unable to select mailbox: %s", mailbox)
	}

	// Set search criteria
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	criteria.Header.Set("Message-ID", id)

	ids, err := p.client.Search(criteria)
	if err != nil {
		return errors.WithMessagef(err, "unable to find message: %s", id)
	}

	if len(ids) == 0 {
		return errors.Errorf("message not found: %s", id)
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{label}
	err = p.client.Store(seqset, item, flags, nil)
	if err != nil {
		return errors.WithMessagef(err, "failed to update message: %s", id)
	}

	return err
}

// Option configures how we set up the client
type Option interface {
	apply(*options)
}

type options struct {
	timeout time.Duration
}

type funcOption struct {
	f func(*options)
}

func (fo *funcOption) apply(o *options) {
	fo.f(o)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}
