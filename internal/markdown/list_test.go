package markdown

import (
	"strings"
	"testing"
)

func TestBlockParser_ContinueItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		src     string
		child   bool
		want    bool
		columns int
	}{
		{"ends an empty item at an empty line", "", false, false, 0},
		{"ends an empty item at a blank line below its indentation", " ", false, false, 0},
		{"ends an empty item at a blank line at its indentation", "  ", false, false, 0},
		{"ends an empty item at a blank line beyond its indentation", "    ", false, false, 0},
		{"ends an empty item at a blank line with a tab", "\t", false, false, 0},
		{"ends an empty item at a blank line with spaces and a tab", " \t ", false, false, 0},
		{"continues an item with a child at an empty line", "", true, true, 0},
		{"continues an item with a child at a blank line below its indentation", " ", true, true, 1},
		{"continues an item with a child at a blank line at its indentation", "  ", true, true, 2},
		{"continues an item with a child at a blank line beyond its indentation", "    ", true, true, 2},
		{"continues an item with a child at a blank line with a tab", "\t", true, true, 2},
		{"continues an empty item at indented content", "  a", false, true, 2},
		{"continues an empty item at tab-indented content", "\ta", false, true, 2},
		{"ends an empty item at content below its indentation", " a", false, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := blockParser{
				src: []byte(tt.src), l: line{end: count(len(tt.src))}, spaceFrom: 1,
				containers: []container{{kind: Document}, {kind: List, node: 1}},
			}
			if got := p.continueItem(container{kind: ListItem, indent: 2, child: tt.child}); got != tt.want {
				t.Fatalf("continueItem on %q with child %t = %t, want %t", tt.src, tt.child, got, tt.want)
			}
			if got := p.col + p.used; got != tt.columns {
				t.Fatalf("continueItem on %q consumes %d columns, want %d", tt.src, got, tt.columns)
			}
			if !tt.want && len(p.prefix) != 0 {
				t.Fatalf("continueItem on %q adds prefixes for a line that does not continue the item", tt.src)
			}
		})
	}
}

func TestParse_EmptyListItemBlankLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"ends an empty bullet item before a paragraph", "-\n  \n  a\n", "<ul>\n<li></li>\n</ul>\n<p>a</p>\n"},
		{"ends an empty bullet item before indented code", "-\n    \n    a\n", "<ul>\n<li></li>\n</ul>\n<pre><code>a\n</code></pre>\n"},
		{"ends an empty ordered item before a paragraph", "1.\n   \n   a\n", "<ol>\n<li></li>\n</ol>\n<p>a</p>\n"},
		{"ends an empty bullet item before tab-indented code", "-\n\t\n\ta\n", "<ul>\n<li></li>\n</ul>\n<pre><code>a\n</code></pre>\n"},
		{"makes a list loose after an empty item", "-\n  \n- a\n", "<ul>\n<li></li>\n<li>\n<p>a</p>\n</li>\n</ul>\n"},
		{"keeps an indented sibling outside an empty item", "-\n  \n  - a\n", "<ul>\n<li></li>\n<li>\n<p>a</p>\n</li>\n</ul>\n"},
		{"ends an empty item before a paragraph in a quote", "> -\n>   \n>   a\n", "<blockquote>\n<ul>\n<li></li>\n</ul>\n<p>a</p>\n</blockquote>\n"},
		{"ends an empty item before indented code in a quote", "> -\n>     \n>     a\n", "<blockquote>\n<ul>\n<li></li>\n</ul>\n<pre><code>a\n</code></pre>\n</blockquote>\n"},
		{"ends an empty item at a line without whitespace", "-\n\n  a\n", "<ul>\n<li></li>\n</ul>\n<p>a</p>\n"},
		{"keeps a blank line inside an item with content", "- a\n  \n  b\n", "<ul>\n<li>\n<p>a</p>\n<p>b</p>\n</li>\n</ul>\n"},
	}
	for _, tt := range tests {
		for _, ending := range []struct{ name, text string }{{"LF", "\n"}, {"CR", "\r"}, {"CRLF", "\r\n"}} {
			t.Run(tt.name+"/"+ending.name, func(t *testing.T) {
				t.Parallel()

				src := strings.ReplaceAll(tt.src, "\n", ending.text)
				tree := Parse([]byte(src))
				if err := tree.Verify(); err != nil {
					t.Fatalf("Parse(%q).Verify() = %v", src, err)
				}
				if got := renderHTML(tree, false); got != tt.want {
					t.Fatalf("Parse(%q) HTML = %q, want %q", src, got, tt.want)
				}
			})
		}
	}
}

func TestTree_ListStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		src     string
		start   int
		ordered bool
	}{
		{"reads a bullet list", "- a", 0, false},
		{"reads an ordered list", " 003) a", 3, true},
		{"reads a marker after a split tab", ">\t 1. a", 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tree := Parse([]byte(tt.src))
			start, ordered := tree.ListStart(firstOf(tree, List))
			if start != tt.start || ordered != tt.ordered {
				t.Fatalf("ListStart of %q = %d, %v, want %d, %v", tt.src, start, ordered, tt.start, tt.ordered)
			}
		})
	}
}

func TestTree_ListLoose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want bool
	}{
		{"reads a tight list", "- a\n- b", false},
		{"reads a loose list", "- a\n\n- b", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tree := Parse([]byte(tt.src))
			if got := tree.ListLoose(firstOf(tree, List)); got != tt.want {
				t.Fatalf("ListLoose of %q = %v, want %v", tt.src, got, tt.want)
			}
		})
	}
}
