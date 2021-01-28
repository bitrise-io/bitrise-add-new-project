package phases

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationParams(t *testing.T) {
	type testCase struct {
		organizationSlug string
		isExpectedInJSON bool
	}
	testCases := []testCase{
		{
			"",
			false,
		},
		{
			"notempty",
			true,
		},
	}

	for _, test := range testCases {
		// Given
		progress := Progress{}
		progress.OrganizationSlug = test.organizationSlug

		// When
		params, err := toRegistrationParams(progress)
		bytes, err := json.Marshal(params)

		// Then
		require.NoError(t, err)
		assert.Equal(t, test.isExpectedInJSON, strings.Contains(string(bytes), "organization_slug"))
	}
}
