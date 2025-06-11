package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_resultExtract(t *testing.T) {
	cases := map[string]struct {
		input, want string
	}{
		"not think": {
			input: "hi this is a string.\nno thinking about it.\n",
			want:  "hi this is a string.\nno thinking about it.\n",
		},
		"think": {
			input: "<think>hi this is a string.\n</think>no thinking about it.\n",
			want:  "no thinking about it.\n",
		},
		"think multiline": {
			input: "<think>\nhi this is a string.\nmore\nand more\n</think>no thinking about it.\n",
			want:  "no thinking about it.\n",
		},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, removeThink.ReplaceAllString(tc.input, ""))
		})
	}
}

func Test_codeExtract(t *testing.T) {
	b, err := os.ReadFile("testdata/code.md")
	want := `fn something(i: i32) -> i32 {
  i
}

def something():
  return "something"

`
	require.NoError(t, err)
	require.Equal(t, want, codeExtract(string(b)))
}
