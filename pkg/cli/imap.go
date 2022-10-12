package cli

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ImapCmd is the parent for email command
type EmailCmd struct {
	Auth     EmailAuthCmd     `cmd:"" help:"login to the server"`
	Messages EmailMessagesCmd `cmd:"" help:"list last messages"`
	Get      EmailGetCmd      `cmd:"" help:"get message"`
	Label    EmailLabelCmd    `cmd:"" help:"label the message"`
}

// AuthCmd authenticates to Email server
type EmailAuthCmd struct {
	Email string `kong:"arg" required:""`
}

// Run the command
func (a *EmailAuthCmd) Run(ctx *Cli) error {
	c, err := ctx.Connect(a.Email)
	if err != nil {
		return errors.WithStack(err)
	}

	// Don't forget to logout
	defer c.Close()

	w := ctx.Writer()

	mailboxes, err := c.Mailboxes()
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Fprintln(w, "Mailboxes:")
	for _, m := range mailboxes {
		fmt.Fprintln(w, "* "+m.Name)
	}

	return nil
}

// EmailMessagesCmd prints messages
type EmailMessagesCmd struct {
	Email string `kong:"arg" required:""`
	From  int    `help:"Offset from the last message"`
	Limit int    `default:"10"`
}

// Run the command
func (a *EmailMessagesCmd) Run(ctx *Cli) error {
	c, err := ctx.Connect(a.Email)
	if err != nil {
		return errors.WithStack(err)
	}

	// Don't forget to logout
	defer c.Close()

	list, err := c.Messages("INBOX", uint32(a.From), uint32(a.Limit))
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.WriteJSON(list)
	/*
		w := ctx.Writer()
		fmt.Fprintln(w, "Messages:")
		for _, m := range list {
			fmt.Fprintf(w, "%v\n", m)
		}
	*/

	return nil
}

// EmailGetCmd gets message
type EmailGetCmd struct {
	Email string `kong:"arg" required:""`
	ID    string `help:"Message ID"`
}

// Run the command
func (a *EmailGetCmd) Run(ctx *Cli) error {
	c, err := ctx.Connect(a.Email)
	if err != nil {
		return errors.WithStack(err)
	}

	// Don't forget to logout
	defer c.Close()

	id := strings.ReplaceAll(a.ID, "\\u003c", "<")
	id = strings.ReplaceAll(id, "\\u003e", ">")

	msg, err := c.GetMessage("INBOX", id)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.WriteJSON(msg)

	return nil
}

// EmailLabelCmd labels message
type EmailLabelCmd struct {
	Email string `kong:"arg" required:""`
	ID    string `help:"Message ID"`
}

// Run the command
func (a *EmailLabelCmd) Run(ctx *Cli) error {
	c, err := ctx.Connect(a.Email)
	if err != nil {
		return errors.WithStack(err)
	}

	// Don't forget to logout
	defer c.Close()

	id := strings.ReplaceAll(a.ID, "\\u003c", "<")
	id = strings.ReplaceAll(id, "\\u003e", ">")

	err = c.LabelMessage("INBOX", id, "\\able-downloaded")
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Fprintln(ctx.Writer(), "updated")

	return nil
}
