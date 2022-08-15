package spider

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

var (
	//  domain regexp
	domainRegexp = regexp.MustCompile(`(([[:alnum:]]-?)?([[:alnum:]]-?)+\.)+[[:alpha:]]{2,4}`)
)

// FindDomains
func FindDomains(body []byte) (domains []Domain) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return
	}

	var s = UnescapeHTML.Replace(doc.Text())

	for _, domain := range domainRegexp.FindAllString(s, -1) {
		name, tld, ok := splitDomain(domain)
		if ok {
			domains = append(domains, Domain{
				Name: name,
				TLD:  tld,
			})
		}
	}

	return
}

// SplitDomain
func splitDomain(d string) (name string, tld string, ok bool) {
	// get domain tld
	root, err := publicsuffix.EffectiveTLDPlusOne(d)
	if err != nil {
		return
	}

	//convert to domain name, and tld
	i := strings.Index(root, ".")
	tld = root[i+1:]

	if _, ok = tlds[tld]; !ok {
		return
	}

	root = strings.ToLower(root)
	tld = strings.ToLower(tld)
	name = strings.TrimSuffix(root, "."+tld)

	return
}
