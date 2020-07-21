package crawler

import (
	"net/url"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

var (
	// ignore some extension
	disallowedURLs = regexp.MustCompile(`^.*\.(png|jpg|jpeg|gif|ico|eps|pdf|iso|mp3|mp4|zip|aif|mpa|wav|wma|7z|deb|pkg|rar|rpm|bin|dmg|dat|tar|exe|ps|psd|svg|tif|tiff|pps|ppt|pptx|xls|xlsx|wmv|doc|docx|txt|mov|mpl)$`)
)

/*
linkCheck skip urls that contain tel:, mailto:, javascropt:
*/
func linkCheck(link string) bool {
	if len(link) == 0 {
		return false
	}
	if strings.Contains(link, "tel:") || strings.Contains(link, "mailto:") || strings.Contains(link, "javascript:") {
		return false
	}
	return !disallowedURLs.MatchString(link)
}

/*
extractDomain: extract root domain name
*/
func extractDomain(domain string) (string, string, bool) {
	// get domain tld
	root, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return "", "", false
	}

	//convert to domain name, and tld
	i := strings.Index(root, ".")
	tld := root[i+1:]

	if _, ok := tldList[tld]; !ok {
		return "", "", false
	}
	root = strings.ToLower(root)
	// return
	return root, tld, true
}

/*
extractDomainLink: extract root domain name from URL
*/
func extractDomainLink(link string) (string, string, bool) {
	raw, err := url.Parse(link)
	if err != nil {
		return "", "", false
	}
	// get domain tld
	root, err := publicsuffix.EffectiveTLDPlusOne(raw.Hostname())
	if err != nil {
		return "", "", false
	}

	//convert to domain name, and tld
	i := strings.Index(root, ".")
	tld := root[i+1:]

	if _, ok := tldList[tld]; !ok {
		return "", "", false
	}

	root = strings.ToLower(root)
	// return
	return root, tld, true
}

/*
unescape HTML by replacing decoded HTML characters
*/

var unescapeHTML = strings.NewReplacer(
	`u002F`, `/`,
	`u002f`, `/`,
	`u202f`, `@`,
	`u202F`, `@`,
	`252f`, `/`,
	`252F`, `/`,
	`u000A`, ` `,
	`u000a`, ` `,
	`u002A`, `*`,
	`u002a`, `*`,
	`u003E`, `<`,
	`u003e`, `<`,
	`u00A0`, `;`,
	`u00a0`, `v`,
	`u0025`, `@@`,
	`%21`, `!`,
	`%23`, `#`,
	`%24`, `$`,
	`%26`, `&`,
	`%27`, `'`,
	`%28`, `(`,
	`%29`, `)`,
	`%2A`, `*`,
	`%2B`, `+`,
	`%2C`, `,`,
	`%2F`, `/`,
	`%2f`, `/`,
	`%3A`, `:`,
	`%3B`, `;`,
	`%3D`, `=`,
	`%40`, `@`,
	`%5B`, `[`,
	`%5D`, `x`,
	`%20`, ` `,
	`%22`, `"`,
	`%25`, `%`,
	`%2D`, `*`,
	`%2E`, `~`,
	`%2e`, `~`,
	`%3C`, `<`,
	`%3c`, `<`,
	`%3E`, `>`,
	`%3e`, `>`,
	`%5E`, `^`,
	`%5e`, `^`,
	`%5F`, `_`,
	`%5f`, `_`,
	`%60`, "`",
	`%7B`, `{`,
	`%7b`, `{`,
	`%7C`, `|`,
	`%7c`, `|`,
	`%7D`, `}`,
	`%7d`, `}`,
	`%7E`, `~`,
	`%7e`, `~`,
	`x2F`, `**`,
	`&#x40`, `@`,
	`&#X40`, `@`, // to test
	`note-`, `**`,
	`ref-`, `**`,
)

/*
writeData write data, visited urls, found domains & valid domains
*/
func writeData(path string, data chan string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := os.Chmod(path, 0777); err != nil {
		return err
	}
	go func(d chan string) {
		for data := range d {
			file.WriteString(data + "\n")
		}
		// close file
		file.Close()
	}(data)
	return nil
}
