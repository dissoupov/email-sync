package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/effective-security/xpki/x/ctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	var c Cli

	assert.NotNil(t, c.ErrWriter())
	assert.NotNil(t, c.Writer())
	assert.NotNil(t, c.Reader())

	c.WithErrWriter(os.Stderr)
	c.WithReader(os.Stdin)
	c.WithWriter(os.Stdout)

	assert.NotNil(t, c.ErrWriter())
	assert.NotNil(t, c.Writer())
	assert.NotNil(t, c.Reader())

	out := bytes.NewBuffer([]byte{})
	c.WithWriter(out)
	err := c.WriteJSON(struct{}{})
	require.NoError(t, err)
	assert.Equal(t, "{}\n", out.String())

	out2 := bytes.NewBuffer([]byte{})
	c.WithWriter(out2)
	err = c.WriteYaml(struct{}{})
	require.NoError(t, err)
	assert.Equal(t, "{}\n", out2.String())
}

func TestParse(t *testing.T) {
	var cl struct {
		Cli

		Cmd struct {
			Ptr *bool `help:"test bool ptr"`
		} `kong:"cmd"`
	}

	p := mustNew(t, &cl)
	ctx, err := p.Parse([]string{"--cfg", "testdata/oauth2_clients.yaml", "cmd", "--ptr=false"})
	require.NoError(t, err)
	require.Equal(t, "cmd", ctx.Command())
	if assert.NotNil(t, cl.Cmd.Ptr) {
		assert.False(t, *cl.Cmd.Ptr)
	}
}

func mustNew(t *testing.T, cli interface{}, options ...kong.Option) *kong.Kong {
	t.Helper()
	options = append([]kong.Option{
		kong.Name("test"),
		kong.Exit(func(int) {
			t.Helper()
			t.Fatalf("unexpected exit()")
		}),
		ctl.BoolPtrMapper,
	}, options...)
	parser, err := kong.New(cli, options...)
	require.NoError(t, err)

	return parser
}
