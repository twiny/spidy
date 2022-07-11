package spider

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

var (
	//  domain regexp
	domainRegexp = regexp.MustCompile(`(([[:alnum:]]-?)?([[:alnum:]]-?)+\.)+[[:alpha:]]{2,4}`)
)

var count int32 = 0

var log = func(b []byte) error {
	var dir = "./dump"
	dir = strings.TrimSuffix(dir, "/")

	var fn = fmt.Sprintf("%s/%d.html", dir, count)

	if _, err := os.Stat(fn); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		return err
	}

	atomic.AddInt32(&count, 1)

	return nil
}

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

// // FindDomains
// // this func extract any string that match `domainRegexp`
// // this will lead to having a lot of invalid domains.
// func FindDomains(body []byte) []Domain {
// 	s := unescapeHTML.Replace(string(body))

// 	var domains = []Domain{}
// 	for _, domain := range domainRegexp.FindAllString(s, -1) {
// 		name, tld, ok := splitDomain(domain)
// 		if ok {
// 			domains = append(domains, Domain{
// 				Name: name,
// 				TLD:  tld,
// 			})
// 		}
// 	}

// 	return domains
// }
