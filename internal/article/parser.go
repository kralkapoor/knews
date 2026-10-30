package article

// Parser extracts an Article from a fetched HTML page.
type Parser interface {
	// CanParse reports whether the parser handles pages at articleURL.
	CanParse(articleURL string) bool

	// Parse extracts the article from html, which was fetched from articleURL.
	Parse(html []byte, articleURL string) (*Article, error)
}

// Registry picks a Parser for an article URL. The zero value is an empty
// registry ready to use.
type Registry struct {
	parsers []Parser
}

// Register adds p to the registry. Parsers are tried in registration order.
func (r *Registry) Register(p Parser) {
	r.parsers = append(r.parsers, p)
}

// parserFor returns the first registered parser that can handle articleURL,
// or nil if there is none.
func (r *Registry) parserFor(articleURL string) Parser {
	for _, p := range r.parsers {
		if p.CanParse(articleURL) {
			return p
		}
	}
	return nil
}
