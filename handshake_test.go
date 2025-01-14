package ch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-faster/ch/cht"
	"github.com/go-faster/ch/proto"
)

func TestDial_Exception(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	server := cht.New(t)

	client, err := Dial(ctx, Options{
		Address:  server.TCP,
		Password: "invalid_password",
	})

	var e *Exception
	require.Nil(t, client)
	require.ErrorAs(t, err, &e)
	require.True(t, IsErr(err, proto.ErrAuthenticationFailed))
}
