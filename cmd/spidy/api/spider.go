package api

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	//
	"github.com/twiny/spidy/v2/internal/pkg/spider/v1"
	"github.com/twiny/spidy/v2/internal/service/cache"
	"github.com/twiny/spidy/v2/internal/service/writer"

	//
	"github.com/twiny/domaincheck"
	"github.com/twiny/flog"
	"github.com/twiny/wbot"
)

//go:embed version
var Version string

// Spider
type Spider struct {
	wg      *sync.WaitGroup
	setting *spider.Setting
	bot     *wbot.WBot
	pages   chan *spider.Page
	check   *domaincheck.Checker
	store   spider.Storage
	write   spider.Writer
	log     *flog.Logger
}

// NewSpider
func NewSpider(fp string) (*Spider, error) {
	// get settings
	setting := spider.ParseSetting(fp)

	// crawler opts
	opts := []wbot.Option{
		wbot.SetParallel(setting.Parralle),
		wbot.SetMaxDepth(setting.Crawler.MaxDepth),
		wbot.SetRateLimit(setting.Crawler.Limit.Rate, setting.Crawler.Limit.Interval),
		wbot.SetMaxBodySize(setting.Crawler.MaxBodySize),
		wbot.SetUserAgents(setting.Crawler.UserAgents),
		wbot.SetProxies(setting.Crawler.Proxies),
	}

	bot := wbot.NewWBot(opts...)

	check, err := domaincheck.NewChecker()
	if err != nil {
		return nil, err
	}

	// store
	store, err := cache.NewCache(setting.Store.TTL, setting.Store.Path)
	if err != nil {
		return nil, err
	}

	// logger
	log, err := flog.NewLogger(setting.Log.Path, "spidy", setting.Log.Rotate)
	if err != nil {
		return nil, err
	}

	write, err := writer.NewCSVWriter(setting.Result.Path)
	if err != nil {
		return nil, err
	}

	return &Spider{
		wg:      &sync.WaitGroup{},
		setting: setting,
		bot:     bot,
		pages:   make(chan *spider.Page, setting.Parralle),
		check:   check,
		store:   store,
		write:   write,
		log:     log,
	}, nil
}

// Start
func (s *Spider) Start(links []string) error {
	// go crawl
	s.wg.Add(len(links))
	for _, link := range links {
		go func(l string) {
			defer s.wg.Done()
			//
			if err := s.bot.Crawl(l); err != nil {
				s.log.Error(err.Error(), map[string]string{"url": l})
			}
		}(link)
	}

	// check domains
	s.wg.Add(s.setting.Parralle)
	for i := 0; i < s.setting.Parralle; i++ {
		go func() {
			defer s.wg.Done()
			// results
			for res := range s.bot.Stream() {
				// if response is ok
				if res.Status != http.StatusOK {
					s.log.Info("bad HTTP status", map[string]string{
						"url":    res.URL.String(),
						"status": strconv.Itoa(res.Status),
					})
					continue
				}

				// extract domains
				domains := spider.FindDomains(res.Body)

				// check availability
				for _, domain := range domains {
					root := fmt.Sprintf("%s.%s", domain.Name, domain.TLD)

					// check if allowed extension
					if len(s.setting.TLDs) > 0 {
						if ok := s.setting.TLDs[domain.TLD]; !ok {
							s.log.Info("unsupported domain", map[string]string{
								"domain": root,
								"url":    res.URL.String(),
							})
							continue
						}
					}

					// skip if already checked
					if s.store.HasChecked(root) {
						s.log.Info("already checked", map[string]string{
							"domain": root,
							"url":    res.URL.String(),
						})
						continue
					}

					//
					ctx, cancel := context.WithTimeout(context.Background(), s.setting.Timeout)
					defer cancel()

					status, err := s.check.Check(ctx, root)
					if err != nil {
						s.log.Error(err.Error(), map[string]string{
							"domain": root,
							"url":    res.URL.String(),
						})
						continue
					}

					// save domain
					if err := s.write.Write(&spider.Domain{
						URL:    res.URL.String(),
						Name:   domain.Name,
						TLD:    domain.TLD,
						Status: status.String(),
					}); err != nil {
						s.log.Error(err.Error(), map[string]string{
							"domain": root,
							"url":    res.URL.String(),
						})
						continue
					}

					// terminal print
					fmt.Printf("[Spidy] == domain: %s - status %s\n", root, status.String())
				}
			}
		}()
	}

	s.wg.Wait()
	return nil
}

// Shutdown
func (s *Spider) Shutdown() error {
	// attempt graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigs
	log.Println("shutting down ...")

	// 2nd ctrl+c kills program
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sigs
		log.Println("killing program ...")
		os.Exit(0)
	}()

	s.bot.Close()
	s.log.Close()
	if err := s.store.Close(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
