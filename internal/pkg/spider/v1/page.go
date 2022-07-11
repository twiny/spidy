package spider

import "net/url"

// Page
type Page struct {
	URL    *url.URL
	Status int
	Body   []byte
}
