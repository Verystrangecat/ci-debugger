package workflow_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/murataslan1/ci-debugger/internal/workflow"
)

func TestResolveImage_Defaults(t *testing.T) {
	img, err := workflow.ResolveImage([]string{"ubuntu-latest"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io/catthehacker/ubuntu:act-latest", img)
}

func TestResolveImage_Override(t *testing.T) {
	overrides := map[string]string{
		"ubuntu-latest": "my-custom:image",
	}
	img, err := workflow.ResolveImage([]string{"ubuntu-latest"}, overrides)
	require.NoError(t, err)
	assert.Equal(t, "my-custom:image", img)
}

func TestResolveImage_DirectDockerImage(t *testing.T) {
	img, err := workflow.ResolveImage([]string{"ghcr.io/my/image:tag"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io/my/image:tag", img)
}

func TestResolveImage_Unknown(t *testing.T) {
	_, err := workflow.ResolveImage([]string{"windows-latest"}, nil)
	assert.Error(t, err)
}

func TestResolveImage_Empty(t *testing.T) {
	_, err := workflow.ResolveImage([]string{}, nil)
	assert.Error(t, err)
}
