package notification

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToGitHubDescription(t *testing.T) {
	cases := []struct {
		input          string
		expectedResult string
	}{
		{
			input:          "",
			expectedResult: "",
		},
		{
			input:          "foobar",
			expectedResult: "foobar",
		},
		{
			input:          "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et ma",
			expectedResult: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et ma",
		},
		{
			input:          "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qu",
			expectedResult: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et ma",
		},
	}

	for _, c := range cases {
		result := toGitHubDescription(c.input)
		require.Equal(t, c.expectedResult, *result)
	}
}
