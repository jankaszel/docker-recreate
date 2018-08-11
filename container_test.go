package recreate

import (
	"github.com/fsouza/go-dockerclient"
	"testing"
)

// <https://stackoverflow.com/a/15312097>
func testEquality(a []string, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestMergeContainerEnv(t *testing.T) {
	env := make(map[string]string)
	env["FOO"] = "BAR"
	env["BAR"] = "baz123"

	config := docker.CreateContainerOptions{
		Config: &docker.Config{
			Env: []string{"BAZ=bar"},
		},
	}

	expected := []string{
		"BAZ=bar",
		"FOO=BAR",
		"BAR=baz123",
	}

	received := mergeContainerEnv(config, env)

	if !testEquality(expected, received) {
		t.Errorf("Merged environenment variables do not equal:\nExpected: %v\nReceived: %v\n",
			expected,
			received,
		)
	}
}