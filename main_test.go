package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Agency(t *testing.T) {
	t.Skip()
	agents, err := loadConfig("testdata/agency.yaml")
	require.NoError(t, err)
	require.Len(t, agents, 1)
	agent := agents[0]
	require.Equal(t, agent.AgentName, "planner")
	out, err := agent.do(t.Context(), "testdata/example.md")
	require.NoError(t, err)
	require.NotEmpty(t, out)
}

func Test_getType(t *testing.T) {
	cases := map[string]struct {
		input string
		want  InputType
	}{
		"file": {
			input: "testdata/agency.yaml",
			want:  FileType,
		},
		"glob": {
			input: "testdata/*.yaml",
			want:  FileType,
		},
		"url": {
			input: "https://localhost:3000",
			want:  UrlType,
		},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, getType(tc.input))
		})
	}
}
