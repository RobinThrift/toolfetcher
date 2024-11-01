package toolfile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToolFile(t *testing.T) {
	tt := []struct {
		name     string
		contents string
		err      error
		entries  Entries
	}{
		{name: "Empty Tool File", contents: "", entries: Entries{}},

		{
			name: "Valid Tool File",
			contents: `# Go Dev Tools
staticcheck: go://honnef.co/go/tools/cmd/staticcheck@2024.1.1

# OpenAPI
oapi-codegen: github-releases://oapi-codegen/oapi-codegen@2.4.1
`,
			entries: Entries{"staticcheck": {Name: "staticcheck", Version: "2024.1.1"}, "oapi-codegen": {Name: "oapi-codegen", Version: "2.4.1"}},
		},

		{
			name: "Invalid Tool File/Missing Name",
			contents: `go://honnef.co/go/tools/cmd/staticcheck@2024.1.1
oapi-codegen: github-releases://oapi-codegen/oapi-codegen@2.4.1
`,
			err: ErrMissingName,
		},

		{
			name: "Invalid Tool File/Missing Version",
			contents: `staticcheck: go://honnef.co/go/tools/cmd/staticcheck@2024.1.1
oapi-codegen: github-releases://oapi-codegen/oapi-codegen
`,
			err: ErrMissingVersion,
		},

		{
			name: "Invalid Tool File/Incomplete Line",
			contents: `staticcheck: go://honnef.co/go/tools/cmd/staticcheck@2024.1.1
oapi-codegen:
`,
			err: ErrMissingVersion,
		},

		{
			name: "Invalid Tool File/Missing Colon",
			contents: `staticcheck: go://honnef.co/go/tools/cmd/staticcheck@2024.1.1
oapi-codegen  github-releases://oapi-codegen/oapi-codegen@2.4.1
`,
			err: ErrMissingName,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.contents)
			entries, err := ParseToolFile(r)
			if tt.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.entries, entries)
			} else {
				assert.ErrorIs(t, err, tt.err)
				assert.Len(t, entries, 0)
			}
		})
	}
}
