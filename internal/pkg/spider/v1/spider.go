package spider

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"golang.org/x/net/publicsuffix"
)

var (
	domainRe     = regexp.MustCompile(`(([[:alnum:]]-?)?([[:alnum:]]-?)+\.)+[[:alpha:]]{2,4}`)
	unicodeRe    = regexp.MustCompile(`\\u([0-9a-fA-F]{4})`) // The regex pattern to match \uXXXX sequences
	urlEncodedRe = regexp.MustCompile(`%([0-9a-fA-F]{2})`)   // The regex pattern to match %XX sequences
)

type (
	Storage interface {
		HasChecked(name string) bool
		Close() error
	}

	Writer interface {
		Write(*Domain) error
		Close() error
	}

	Domain struct {
		URL    string
		Name   string
		TLD    string
		Status string
	}

	Page struct {
		URL    *url.URL
		Status int
		Body   []byte
	}

	Stats struct {
		URLs struct {
			Available uint32
			Unique    uint32
			Totals    uint32
		}
		Domains struct {
			Available uint32
			Unique    uint32
			Totals    uint32
		}
	}
)

func (d Domain) CSVRow() []string {
	var row []string
	return append(row, d.URL, d.Name, d.TLD, d.Status)
}

var (
	dirtyInt = 1
)

func FindDomains(body []byte) (domains []Domain) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return
	}

	// // for debug save html in file
	// os.WriteFile("./dump/"+strconv.Itoa(dirtyInt)+".html", body, 0644)
	// dirtyInt++

	text := decodeHTMLChars(doc.Text())

	for _, domain := range domainRe.FindAllString(text, -1) {
		name, tld, ok := splitAndValidate(domain)
		if ok {
			domains = append(domains, Domain{
				Name: name,
				TLD:  tld,
			})
		}
	}

	return
}

func ParseRateLimit(input string) (int, time.Duration, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format")
	}

	reqs, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	switch parts[1] {
	case "s", "S":
		return reqs, time.Second, nil
	case "m":
		return reqs, time.Minute, nil
	case "h":
		return reqs, time.Hour, nil
	default:
		return 0, 0, fmt.Errorf("unknown duration specifier")
	}
}

func ParseSize(input string) (int64, error) {
	size, err := humanize.ParseBytes(input)
	if err != nil {
		return 0, fmt.Errorf("invalid format")
	}

	return int64(size), nil
}

func IsAllowedTLD(allowed []string, tld string) bool {
	for _, a := range allowed {
		if a == tld {
			return true
		}
	}

	return false
}

func splitAndValidate(d string) (name string, tld string, ok bool) {
	if len(d) > 253 {
		return "", "", false
	}

	d = strings.ToLower(d)

	root, err := publicsuffix.EffectiveTLDPlusOne(d)
	if err != nil {
		return "", "", false
	}

	parts := strings.SplitN(root, ".", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	name, tld = parts[0], parts[1]

	if _, ok := tlds[tld]; !ok {
		return "", "", false
	}

	return name, tld, true
}

func decodeHTMLChars(s string) string {
	// Decoder for unicode sequences
	unicodeDecoder := func(match string) string {
		hex := match[2:] // Extract the hexadecimal part from the match

		// Convert hex to integer
		codepoint, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return match
		}

		return string(rune(codepoint)) // Convert codepoint to string and return
	}

	// Decoder for URL-encoded sequences
	urlEncodedDecoder := func(match string) string {
		hex := match[1:] // Extract the hexadecimal part from the match

		// Convert hex to integer
		codepoint, err := strconv.ParseInt(hex, 16, 8) // Note: Changed to 8 bits because it represents a single byte
		if err != nil {
			return match
		}

		return string(rune(codepoint)) // Convert codepoint to string and return
	}

	// Replace all occurrences of unicode sequences with the actual unicode character
	s = unicodeRe.ReplaceAllStringFunc(s, unicodeDecoder)

	// Replace all occurrences of URL-encoded sequences with the actual character
	s = urlEncodedRe.ReplaceAllStringFunc(s, urlEncodedDecoder)

	return s
}
