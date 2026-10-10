package format

import (
	"bytes"
	"strings"
	"testing"
)

func TestSource_EmptyListItemDialect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"keeps a tab blank and the following paragraph", "*\n\t\n  0\n", "*\n\t\n  0\n"},
		{"keeps multiple following blocks", "-\n  \n  a\n\n  b\n\n#  h\n", "-\n  \n  a\n\n  b\n\n#  h\n"},
		{"keeps a lazy continuation", "*\n  \n  a\nb\n", "*\n  \n  a\nb\n"},
		{"keeps outside code", "-\n    \n    a\n", "-\n    \n    a\n"},
		{"keeps an indented sibling", "*\n  \n  * a\n", "*\n  \n  * a\n"},
		{"keeps an ordered item", "3.\n   \n   a\n", "3.\n   \n   a\n"},
		{"keeps a boundary in a quote", "> *\n>   \n>   a\n", "> *\n>   \n>   a\n"},
		{"keeps a boundary in nested lists", "* *\n    \n    a\n", "* *\n    \n    a\n"},
		{"keeps a split tab after a quote marker", "> *\n>\t \n>\t a\n", "> *\n>\t \n>\t a\n"},
		{"keeps whitespace-only gaps and trailing blanks", "*\n  \n    ```\n  \t \n\n  ```\n \t\n", "*\n  \n    ```\n  \t \n\n  ```\n \t\n"},
		{"keeps the first ambiguous boundary", "*\n  \n  a\n\n*\n\t\n  b\n", "*\n  \n  a\n\n*\n\t\n  b\n"},
		{"formats the blocks before the protected suffix", "#  h\n\n*a*\n\n*\n\t\n  b\n", "# h\n\n_a_\n\n*\n\t\n  b\n"},
		{"keeps a definition after the boundary", "*\n  \n  [a]: /u\n\n[a]\n", "*\n  \n  [a]: /u\n\n[a]\n"},
		{"keeps a boundary in a footnote definition", "[^a]: *\n      \n      a\n\n[^a]\n", "[^a]: *\n      \n      a\n\n[^a]\n"},
		{"keeps front matter before the formatted prefix", "---\na: 1\n---\n\n#  h\n\n*\n\t\n  b\n", "---\na: 1\n---\n\n# h\n\n*\n\t\n  b\n"},
		{"formats a prefix after a BOM", "\ufeff#  h\n\n*\n\t\n  b\n", "# h\n\n*\n\t\n  b\n"},
		{"adds a final line ending", "*\n\t\n  0", "*\n\t\n  0\n"},
		{"formats an empty item after a shorter blank", "*\n \n  a\n\n#  h\n", "-\n\na\n\n# h\n"},
		{"formats an item that has content", "* a\n  \n  *b*\n\n#  h\n", "- a\n\n  _b_\n\n# h\n"},
	}
	for _, tt := range tests {
		for _, ending := range []struct{ name, text string }{{"LF", "\n"}, {"CR", "\r"}, {"CRLF", "\r\n"}} {
			t.Run(tt.name+"/"+ending.name, func(t *testing.T) {
				t.Parallel()

				src := strings.ReplaceAll(tt.src, "\n", ending.text)
				if got := string(checkSource(t, []byte(src))); got != tt.want {
					t.Fatalf("Source(%q) = %q, want %q", src, got, tt.want)
				}
			})
		}
	}

	t.Run("does not change or alias the input", func(t *testing.T) {
		t.Parallel()

		src := []byte("*\r\n\t\r\n  a")
		original := bytes.Clone(src)
		out, err := Source(src)
		if err != nil {
			t.Fatal(err)
		}
		out[0] = '+'
		if !bytes.Equal(src, original) {
			t.Fatalf("Source changed or aliased its input: got %q, want %q", src, original)
		}
	})
}
