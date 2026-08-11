// Package htmlscan pulls the page title out of an HTML document, standing in
// for the bit of this provider that screen-scrapes a title/status banner off
// the BTP service marketplace page when a structured API isn't available.
//
// ExtractTitle parses the whole document rather than streaming tokens: the
// marketplace pages this reads are small, and the parser handles malformed
// markup that the raw tokenizer would surface as caller-visible errors.
package htmlscan

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// ExtractTitle parses an HTML document and returns the text content of its
// first <title> element.
func ExtractTitle(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", fmt.Errorf("htmlscan: parse document: %w", err)
	}
	title, ok := findTitle(doc)
	if !ok {
		return "", fmt.Errorf("htmlscan: no <title> element found")
	}
	return title, nil
}

// findTitle walks the parsed node tree depth-first looking for the first
// <title> element with text content.
func findTitle(n *html.Node) (string, bool) {
	if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
		return strings.TrimSpace(n.FirstChild.Data), true
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title, ok := findTitle(c); ok {
			return title, ok
		}
	}
	return "", false
}
