package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Agency(t *testing.T) {
	agency, err := loadConfig("testdata/agency.yaml")
	require.NoError(t, err)
	require.NoError(t, agency.Dispatch(context.TODO(), "planner", false, "testdata/*.md"))
}
