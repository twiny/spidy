package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/gocolly/colly/v2/proxy"

	domainCheck "github.com/twiny/domain"
)

var (
	// list of support tlds
	supportedTLD = &Store{
		data: make(map[string]struct{}),
		mu:   &sync.Mutex{},
	}

	// domain name regex
	reDomain = regexp.MustCompile(`(([[:alnum:]]-?)?([[:alnum:]]-?)+\.)+[[:alnum:]]{2,4}`)

	// whois dialer
	dialer = net.Dialer{
		Timeout: 30 * time.Second,
	}
)

// Writer struct
type Writer struct {
	visits  chan string
	domains chan string
	valids  chan string
}

// Spider struct
type Spider struct {
	collector *colly.Collector
	setting   *Setting
	//
	domainDB  *Store
	domains   chan string
	responses chan []byte
	//
	writer *Writer
	err    chan *Error
}

/*
NewSpider with setting
*/
func NewSpider(setting *Setting) (*Spider, error) {
	// load & save supported TLDs
	switch {
	case len(setting.Engine.TLDs) == 0:
		for tld := range tldList {
			supportedTLD.Save(tld)
		}
	case len(setting.Engine.TLDs) > 0:
		for _, tld := range setting.Engine.TLDs {
			supportedTLD.Save(tld)
		}
	}

	// timeout
	timeout, err := time.ParseDuration(setting.Engine.Timeout)
	if err != nil {
		return nil, &Error{
			time.Now(),
			OnTimeParse,
			"",
			"",
			err,
		}
	}

	// random delay
	randomDelay, err := time.ParseDuration(setting.Engine.RandomDelay)
	if err != nil {
		return nil, &Error{
			time:   time.Now(),
			on:     OnTimeParse,
			url:    "",
			domain: "",
			err:    err,
		}
	}

	// collector
	collector := colly.NewCollector()

	// option
	collector.MaxDepth = setting.Engine.Depth
	collector.Async = true
	collector.CheckHead = true
	collector.ParseHTTPErrorResponse = true
	collector.AllowURLRevisit = false

	// limiter
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: setting.Engine.Parallel,
		Delay:       5 * time.Second,
		RandomDelay: randomDelay,
	})

	// default transport
	collector.WithTransport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	// error channel
	errs := make(chan *Error, setting.Engine.Worker)

	// set proxies
	if len(setting.Engine.ProxyList) > 0 {
		proxies, err := proxy.RoundRobinProxySwitcher(
			setting.Engine.ProxyList...,
		)
		if err != nil {
			e := &Error{
				time.Now(),
				OnProxyConnect,
				"",
				"",
				err,
			}
			errs <- e
		}

		//
		collector.WithTransport(&http.Transport{
			Proxy: proxies,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: timeout,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		})
	}

	// return spider
	return &Spider{
		collector: collector,
		setting:   setting,
		//
		domainDB:  NewStore(), // unique domain storage
		domains:   make(chan string, setting.Engine.Worker),
		responses: make(chan []byte, setting.Engine.Worker),
		//
		writer: &Writer{
			visits:  make(chan string, setting.Engine.Worker),
			domains: make(chan string, setting.Engine.Worker),
			valids:  make(chan string, setting.Engine.Worker),
		},
		//
		err: errs,
	}, nil
}

/*
collect: collects domains from HTML of input url
*/
func (s *Spider) collect(inputURL string) {
	// clone main spider
	spider := s.collector.Clone()

	// extract main domain
	mainDomain, _, _ := extractDomainLink(inputURL)

	// User-Agent & referer rotator
	extensions.RandomUserAgent(spider)
	extensions.Referer(spider)

	// start
	// Set error handler
	spider.OnError(func(r *colly.Response, err error) {
		if err != nil {
			s.err <- &Error{
				time.Now(),
				OnScrap,
				r.Request.URL.String(),
				r.Request.URL.Host,
				err,
			}
		}
		//
		r.Request.Abort()
	})

	// On Response extract root domains from response body
	// by matching regex
	spider.OnResponse(func(resp *colly.Response) {
		if resp.StatusCode == http.StatusOK {
			// response body
			respBody := resp.Body

			// parse html doc
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(respBody))
			if err != nil {
				s.err <- &Error{
					time.Now(),
					OnBodyRead,
					"",
					"",
					err,
				}
			}

			// find all links
			doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
				link, exists := sel.Attr("href")
				link = resp.Request.AbsoluteURL(link)
				if exists {
					childDomain, _, _ := extractDomainLink(link)
					if childDomain == mainDomain {
						if visited, _ := spider.HasVisited(link); !visited && linkCheck(link) {
							resp.Request.Visit(link)
							s.writer.visits <- link
						}
					}
				}
			})

			// send response body
			s.responses <- respBody
		}
	})

	// visit
	spider.Visit(inputURL)
	spider.Wait()
}

/*
extract func
*/
func (s *Spider) extract(body []byte) {
	// Find domains in body
	domains := reDomain.FindAllString(unescapeHTML.Replace(string(body)), -1)

	// check length
	if len(domains) > 0 {
		for _, domain := range domains {
			s.domains <- domain
		}
	}
}

/*
validate func
*/
func (s *Spider) validate(domain string) {
	rootDomain, tld, _ := extractDomain(domain)
	//
	if len(rootDomain) > 1 {
		if supported := supportedTLD.Found(tld); supported {
			if found := s.domainDB.Found(rootDomain); !found {
				s.domainDB.Save(rootDomain)
				s.writer.domains <- rootDomain

				// check available
				valid, err := domainCheck.AvailableCheck(rootDomain, dialer.Dial)
				if err != nil {
				}
				switch valid {
				case true:
					fmt.Println(valid, " ", rootDomain, "\n")
					s.writer.valids <- rootDomain
				case false:
					fmt.Println(valid, " ", rootDomain, "\n")
				}
			}
		}
	}
}

// wait groups
var wg1, wg2, wg3 = &sync.WaitGroup{}, &sync.WaitGroup{}, &sync.WaitGroup{}

/*
Run func
*/
func (s *Spider) Run() {
	// open file
	input, err := os.Open(s.setting.Engine.URLPath)
	if err != nil {
		fmt.Println("could not open URLs file")
		return
	}

	// read file
	scanner := bufio.NewScanner(input)
	urls := make(chan string, s.setting.Engine.Worker)
	go func() {
		for scanner.Scan() {
			urls <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
		//
		input.Close()
		close(urls)
	}()

	// run collector
	wg1.Add(s.setting.Engine.Worker)
	for i := 0; i < s.setting.Engine.Worker; i++ {
		go func(u chan string) {
			for url := range u {
				s.collect(url)
			}
			wg1.Done()
		}(urls)
	}

	// close
	go func() {
		wg1.Wait()
		close(s.domains)
	}()

	// extract domains
	wg2.Add(s.setting.Engine.Worker)
	for i := 0; i < s.setting.Engine.Worker; i++ {
		go func(response chan []byte) {
			for body := range response {
				s.extract(body)
			}
			wg2.Done()
		}(s.responses)
	}
	go func() {
		wg2.Wait()
		close(s.responses)
	}()
	// validate
	wg3.Add(20)
	for i := 0; i < 20; i++ {
		go func(d chan string) {
			for domain := range d {
				s.validate(domain)
			}
			wg3.Done()
		}(s.domains)
	}
	//
	wg3.Wait()
	close(s.writer.domains)
	close(s.writer.valids)
	close(s.writer.visits)
	close(s.err)
	//
	s.domainDB.Close()
}

/*
Logger func
*/
func (s *Spider) Logger() error {
	// write visited
	if err := writeData("./log/visited.txt", s.writer.visits); err != nil {
		return err
	}

	// write domains
	if err := writeData("./log/domains.txt", s.writer.domains); err != nil {
		return err
	}

	// write valid
	if err := writeData("./log/valids.txt", s.writer.valids); err != nil {
		return err
	}

	// write errors
	errFile, err := os.Create("./log/errors.log")
	if err != nil {
		return err
	}
	if err := os.Chmod("./log/errors.log", 0777); err != nil {
		return err
	}
	go func() {
		for e := range s.err {
			errFile.WriteString(e.Error() + "\n")
		}
		// close file
		errFile.Close()
	}()

	return nil
}
