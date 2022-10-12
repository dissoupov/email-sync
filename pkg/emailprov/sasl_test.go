package emailprov

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaslClient(t *testing.T) {
	s := saslClient{
		email:       "denis@ableia.io",
		accessToken: "at",
		tokenType:   "Bearer",
	}

	mech, _, err := s.Start()
	require.NoError(t, err)
	assert.Equal(t, "XOAUTH2", mech)

	_, err = s.Next([]byte{})
	require.NoError(t, err)
}
