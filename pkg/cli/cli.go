package cli

import (
	"io"
	"os"
	"time"

	"github.com/ableorg/email-sync/pkg/emailprov"
	"github.com/alecthomas/kong"
	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/xhttp/marshal"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/x/ctl"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
)

var logger = xlog.NewPackageLogger("github.com/ableorg/email-sync/pkg", "cli")

// Cli provides CLI context to run commands
type Cli struct {
	//Cfg      string          `help:"Location of client config file" default:"~/.ableai/.email-sync.yaml" type:"path"`
	Debug    bool            `short:"D" help:"Enable debug mode"`
	LogLevel string          `short:"l" help:"Set the logging level (debug|info|warn|error)" default:"info"`
	Version  ctl.VersionFlag `name:"version" help:"Print version information and quit" hidden:""`

	Server  string        `help:"Email server address" default:"outlook.office365.com:993"`
	Storage string        `help:"Folder location for the cache" default:"~/.ableai/email"`
	Timeout time.Duration `help:"Timeout for email operations" default:"30s"`

	// Stdin is the source to read from, typically set to os.Stdin
	stdin io.Reader
	// Output is the destination for all output from the command, typically set to os.Stdout
	output io.Writer
	// ErrOutput is the destinaton for errors.
	// If not set, errors will be written to os.StdError
	errOutput io.Writer

	ctx context.Context
}

// Context for requests
func (c *Cli) Context() context.Context {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	return c.ctx
}

// Reader is the source to read from, typically set to os.Stdin
func (c *Cli) Reader() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	return os.Stdin
}

// WithReader allows to specify a custom reader
func (c *Cli) WithReader(reader io.Reader) *Cli {
	c.stdin = reader
	return c
}

// Writer returns a writer for control output
func (c *Cli) Writer() io.Writer {
	if c.output != nil {
		return c.output
	}
	return os.Stdout
}

// WithWriter allows to specify a custom writer
func (c *Cli) WithWriter(out io.Writer) *Cli {
	c.output = out
	return c
}

// ErrWriter returns a writer for control output
func (c *Cli) ErrWriter() io.Writer {
	if c.errOutput != nil {
		return c.errOutput
	}
	return os.Stderr
}

// WithErrWriter allows to specify a custom error writer
func (c *Cli) WithErrWriter(out io.Writer) *Cli {
	c.errOutput = out
	return c
}

// AfterApply hook loads config
func (c *Cli) AfterApply(app *kong.Kong, vars kong.Vars) error {
	if c.Debug {
		xlog.SetGlobalLogLevel(xlog.DEBUG)
	} else {
		xlog.SetGlobalLogLevel(xlog.ERROR)
	}

	return nil
}

// WriteJSON prints response to out
func (c *Cli) WriteJSON(value interface{}) error {
	json, err := marshal.EncodeBytes(marshal.PrettyPrint, value)
	if err != nil {
		return errors.WithMessage(err, "failed to encode")
	}
	c.Writer().Write(json)
	c.Writer().Write([]byte{'\n'})

	return nil
}

// WriteYaml prints response to out
func (c *Cli) WriteYaml(value interface{}) error {
	y, err := yaml.Marshal(value)
	if err != nil {
		return errors.WithMessage(err, "failed to encode")
	}
	c.Writer().Write(y)

	return nil
}

// WriteObject prints response to out
func (c *Cli) WriteObject(format string, value interface{}) error {
	if format == "yaml" {
		return c.WriteYaml(value)
	}
	return c.WriteJSON(value)
}

// Connect to email
func (c *Cli) Connect(email string) (*emailprov.Provider, error) {
	t, err := retriable.LoadAuthToken(retriable.ExpandStorageFolder(c.Storage))
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to load auth token, use auth command")
	}

	accessToken, tokenType, _, err := retriable.ParseAuthToken(t)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to parse auth token")
	}
	if tokenType != "Bearer" {
		return nil, errors.Errorf("unsupported token type: %s", tokenType)
	}

	p, err := emailprov.New(c.Server, email, accessToken, emailprov.WithTimeout(c.Timeout))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return p, nil
}
