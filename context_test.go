package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_scrape(t *testing.T) {
	t.Skip()
	got, err := scrape("https://nix-community.github.io/home-manager/index.xhtml")
	require.NoError(t, err)
	t.Log(got)
}
