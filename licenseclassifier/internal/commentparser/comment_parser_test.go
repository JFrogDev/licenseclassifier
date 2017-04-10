// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package commentparser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/licenseclassifier/licenseclassifier/internal/commentparser/language"
)

const (
	singleLineText = "single line text"
	multilineText  = `first line of text
second line of text
third line of text
`
)

func TestCommentParser_Lex(t *testing.T) {
	tests := []struct {
		description string
		lang        language.Language
		source      string
		want        Comments
	}{
		{
			description: "BCPL Single Line Comments",
			lang:        language.Go,
			source:      fmt.Sprintf("//%s\n", singleLineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   1,
					Text:      singleLineText,
				},
			},
		},
		{
			description: "Go Comment With Multiline String",
			lang:        language.Go,
			source:      fmt.Sprintf("var a = `A\nmultiline\\x20\nstring`\n//%s\n", singleLineText),
			want: []*Comment{
				{
					StartLine: 4,
					EndLine:   4,
					Text:      singleLineText,
				},
			},
		},
		{
			description: "Python Multiline String",
			lang:        language.Python,
			source:      fmt.Sprintf("#%s\n'''this is a multiline\nstring'''", singleLineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   1,
					Text:      singleLineText,
				},
			},
		},
		{
			description: "TR Command String",
			lang:        language.Python,
			source: fmt.Sprintf(`#%s
AUTH= \
| tr '"\n' \
| base64 -w
`, singleLineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   1,
					Text:      singleLineText,
				},
			},
		},
		{
			description: "Lisp Single Line Comments",
			lang:        language.Clojure,
			source:      fmt.Sprintf(";%s\n", singleLineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   1,
					Text:      singleLineText,
				},
			},
		},
		{
			description: "Shell Single Line Comments",
			lang:        language.Shell,
			source:      fmt.Sprintf("#%s\n", singleLineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   1,
					Text:      singleLineText,
				},
			},
		},
		{
			description: "BCPL Multiline Comments",
			lang:        language.C,
			source:      fmt.Sprintf("/*%s*/\n", multilineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   4,
					Text:      multilineText,
				},
			},
		},
		{
			description: "BCPL Multiline Comments no terminating newline",
			lang:        language.C,
			source:      fmt.Sprintf("/*%s*/", multilineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   4,
					Text:      multilineText,
				},
			},
		},
		{
			description: "Nested Multiline Comments",
			lang:        language.Swift,
			source:      "/*a /*\n  nested\n*/\n  comment\n*/\n",
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   5,
					Text:      "a /*\n  nested\n*/\n  comment\n",
				},
			},
		},
		{
			description: "Ruby Multiline Comments",
			lang:        language.Ruby,
			source:      fmt.Sprintf("=begin\n%s=end\n", multilineText),
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   5,
					Text:      "\n" + multilineText,
				},
			},
		},
		{
			description: "Multiple Single Line Comments",
			lang:        language.Shell,
			source: `# First line
# Second line
# Third line
`,
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   1,
					Text:      " First line",
				},
				{
					StartLine: 2,
					EndLine:   2,
					Text:      " Second line",
				},
				{
					StartLine: 3,
					EndLine:   3,
					Text:      " Third line",
				},
			},
		},
		{
			description: "Mixed Multiline / Single Line Comments",
			lang:        language.C,
			source: `/*
 * The first multiline line.
 * The second multiline line.
 */
 // The first single line comment.
 // The second single line comment.
`,
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   4,
					Text: `
 * The first multiline line.
 * The second multiline line.
 `,
				},
				{
					StartLine: 5,
					EndLine:   5,
					Text:      " The first single line comment.",
				},
				{
					StartLine: 6,
					EndLine:   6,
					Text:      " The second single line comment.",
				},
			},
		},
		{
			description: "Mixed Multiline / Single Line Comments",
			lang:        language.C,
			source: `/*
 * The first multiline line.
 * The second multiline line.
 */
 // The first single line comment.
 // The second single line comment.
`,
			want: []*Comment{
				{
					StartLine: 1,
					EndLine:   4,
					Text: `
 * The first multiline line.
 * The second multiline line.
 `,
				},
				{
					StartLine: 5,
					EndLine:   5,
					Text:      " The first single line comment.",
				},
				{
					StartLine: 6,
					EndLine:   6,
					Text:      " The second single line comment.",
				},
			},
		},
		{
			description: "HTML-like comments and quotes",
			lang:        language.HTML,
			source: `# This is an important topic
I don't want to go on all day here! <-- notice the quote in there!
<!-- Well, maybe I do... -->
`,
			want: []*Comment{
				{
					StartLine: 3,
					EndLine:   3,
					Text:      " Well, maybe I do... ",
				},
			},
		},
		{
			description: "JavaScript regex",
			lang:        language.JavaScript,
			source: `var re = /hello"world/;
// the comment
`,
			want: []*Comment{
				{
					StartLine: 2,
					EndLine:   2,
					Text:      " the comment",
				},
			},
		},
		{
			description: "Perl regex",
			lang:        language.Perl,
			source: `if (/hello"world/) {
  # the comment
  print "Yo!"
}
`,
			want: []*Comment{
				{
					StartLine: 2,
					EndLine:   2,
					Text:      " the comment",
				},
			},
		},
	}

	for _, tt := range tests {
		got := Parse([]byte(tt.source), tt.lang)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Mismatch(%q) = %+v, want %+v", tt.description, got, tt.want)
		}
	}
}

func TestCommentParser_ChunkIterator(t *testing.T) {
	tests := []struct {
		description string
		comments    Comments
		want        []Comments
	}{
		{
			description: "Empty Comments",
			comments:    Comments{},
			want:        nil,
		},
		{
			description: "Single Line Comment Chunk",
			comments: Comments{
				{StartLine: 1, EndLine: 1, Text: "Block 1 line 1"},
				{StartLine: 2, EndLine: 2, Text: "Block 1 line 2"},
			},
			want: []Comments{
				Comments{
					{StartLine: 1, EndLine: 1, Text: "Block 1 line 1"},
					{StartLine: 2, EndLine: 2, Text: "Block 1 line 2"},
				},
			},
		},
		{
			description: "Multiline Comment Chunk",
			comments: Comments{
				{StartLine: 1, EndLine: 3, Text: "Multiline 1\n2\n3"},
			},
			want: []Comments{
				Comments{{StartLine: 1, EndLine: 3, Text: "Multiline 1\n2\n3"}},
			},
		},
		{
			description: "Multiple Single Line Comment Chunks",
			comments: Comments{
				{StartLine: 1, EndLine: 1, Text: "Block 1 line 1"},
				{StartLine: 2, EndLine: 2, Text: "Block 1 line 2"},
				{StartLine: 5, EndLine: 5, Text: "Block 2 line 1"},
				{StartLine: 6, EndLine: 6, Text: "Block 2 line 2"},
				{StartLine: 10, EndLine: 10, Text: "Block 3 line 1"},
				{StartLine: 11, EndLine: 11, Text: "Block 3 line 2"},
				{StartLine: 13, EndLine: 13, Text: "Block 3 line 3"},
			},
			want: []Comments{
				Comments{
					{StartLine: 1, EndLine: 1, Text: "Block 1 line 1"},
					{StartLine: 2, EndLine: 2, Text: "Block 1 line 2"},
				},
				Comments{
					{StartLine: 5, EndLine: 5, Text: "Block 2 line 1"},
					{StartLine: 6, EndLine: 6, Text: "Block 2 line 2"},
				},
				Comments{
					{StartLine: 10, EndLine: 10, Text: "Block 3 line 1"},
					{StartLine: 11, EndLine: 11, Text: "Block 3 line 2"},
					{StartLine: 13, EndLine: 13, Text: "Block 3 line 3"},
				},
			},
		},
		{
			description: "Multiline Comment Chunk",
			comments: Comments{
				{StartLine: 1, EndLine: 3, Text: "Multiline 1\n2\n3"},
				{StartLine: 4, EndLine: 6, Text: "Multiline 1\n2\n3"},
			},
			want: []Comments{
				Comments{{StartLine: 1, EndLine: 3, Text: "Multiline 1\n2\n3"}},
				Comments{{StartLine: 4, EndLine: 6, Text: "Multiline 1\n2\n3"}},
			},
		},
		{
			description: "Multiline and Single Line Comment Chunks",
			comments: Comments{
				{StartLine: 1, EndLine: 3, Text: "Multiline 1\n2\n3"},
				{StartLine: 4, EndLine: 4, Text: "Block 2 line 1"},
				{StartLine: 5, EndLine: 5, Text: "Block 2 line 2"},
			},
			want: []Comments{
				Comments{
					{StartLine: 1, EndLine: 3, Text: "Multiline 1\n2\n3"},
				},
				Comments{
					{StartLine: 4, EndLine: 4, Text: "Block 2 line 1"},
					{StartLine: 5, EndLine: 5, Text: "Block 2 line 2"},
				},
			},
		},
	}

	for _, tt := range tests {
		i := 0
		for got := range tt.comments.ChunkIterator() {
			if i >= len(tt.want) {
				t.Errorf("Mismatch(%q) more comment chunks than expected = %v, want %v",
					tt.description, i+1, len(tt.want))
				break
			}
			if !reflect.DeepEqual(got, tt.want[i]) {
				t.Errorf("Mismatch(%q) = %+v, want %+v", tt.description, got, tt.want[i])
			}
			i++
		}
		if i != len(tt.want) {
			t.Errorf("Mismatch(%q) not enough comment chunks = %v, want %v",
				tt.description, i, len(tt.want))
		}
	}
}