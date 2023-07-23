package api

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	//

	"github.com/fatih/color"
	"github.com/twiny/flog/v2"
	"github.com/twiny/spidy/v2/internal/pkg/spider/v1"
	"github.com/twiny/spidy/v2/internal/service/cache"
	"github.com/twiny/spidy/v2/internal/service/writer"

	//

	clog "github.com/charmbracelet/log"
	"github.com/twiny/domaincheck"
	"github.com/twiny/wbot"
)

//go:embed version
var Version string

var (
	red   = color.New(color.FgRed).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
)

type Spider struct {
	wg *sync.WaitGroup

	config *spider.CrawlerConfig

	bot   *wbot.WBot
	pages chan *spider.Page

	check *domaincheck.Checker
	store spider.Storage
	write spider.Writer

	uiprint chan *spider.Domain
	stats   *spider.Stats

	log *flog.Logger
}

func NewSpider(fp string, result string) (*Spider, error) {
	config, err := spider.ParseCrawlerConfig(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	rate, interval, err := spider.ParseRateLimit(config.RateLimit)
	if err != nil {
		rate, interval = 10, time.Second // default 10req/s
	}

	maxBodySize, err := spider.ParseSize(config.MaxBodySize)
	if err != nil {
		maxBodySize = 5 * 1024 * 1024 // default 5mb
	}

	opts := []wbot.Option{
		wbot.SetParallel(config.Worker),
		wbot.SetMaxDepth(config.MaxDepth),
		wbot.SetRateLimit(rate, interval),
		wbot.SetMaxBodySize(maxBodySize),
		wbot.SetUserAgents(config.UserAgents),
		wbot.SetProxies(config.Proxies),
	}

	bot := wbot.NewWBot(opts...)

	check, err := domaincheck.NewChecker()
	if err != nil {
		return nil, fmt.Errorf("failed to create domain checker: %w", err)
	}

	mainConfig, err := spider.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	store, err := cache.NewCache(mainConfig.Store.TTL, mainConfig.Store.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	logger, err := flog.NewLogger(mainConfig.Logs.Path, mainConfig.Logs.MaxAge, mainConfig.Logs.MaxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	write, err := writer.NewCSVWriter(result)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	return &Spider{
		wg:      new(sync.WaitGroup),
		config:  config,
		bot:     bot,
		pages:   make(chan *spider.Page, 2048),
		uiprint: make(chan *spider.Domain, 2048),
		stats:   new(spider.Stats),
		check:   check,
		store:   store,
		write:   write,
		log:     logger,
	}, nil
}

func (s *Spider) Start(links []string) error {
	clog.Infof("starting spider for %d links", len(links))

	// go crawl
	s.wg.Add(len(links))
	for _, link := range links {
		go func(l string) {
			defer s.wg.Done()
			if err := s.bot.Crawl(l); err != nil {
				s.log.Error(err.Error(), flog.NewField("url", l))
			}
		}(link)
	}

	// check domains
	s.wg.Add(s.config.Worker)
	for i := 0; i < s.config.Worker; i++ {
		go func() {
			defer s.wg.Done()
			for res := range s.bot.Stream() {
				atomic.AddUint32(&s.stats.URLs.Totals, 1)
				// if response is ok
				if res.Status != http.StatusOK {
					s.log.Info("bad HTTP status", flog.NewField("url", res.URL.String()), flog.NewField("status", res.URL.String()))
					continue
				}

				domains := spider.FindDomains(res.Body)

				for _, domain := range domains {
					atomic.AddUint32(&s.stats.Domains.Totals, 1)
					root := fmt.Sprintf("%s.%s", domain.Name, domain.TLD)
					// check if allowed extension
					if len(s.config.AllowedTLDs) > 0 {
						if ok := spider.IsAllowedTLD(s.config.AllowedTLDs, domain.TLD); !ok {
							s.log.Info("unsupported domain", flog.NewField("domain", root), flog.NewField("url", res.URL.String()))
							continue
						}
					}

					// skip if already checked
					if s.store.HasChecked(root) {
						// s.log.Info("already checked", flog.NewField("domain", root), flog.NewField("url", res.URL.String()))
						continue
					}

					atomic.AddUint32(&s.stats.Domains.Unique, 1)

					//
					func() {
						ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
						defer cancel()

						status, err := s.check.Check(ctx, root)
						if err != nil {
							s.log.Error(err.Error(), flog.NewField("domain", root), flog.NewField("url", res.URL.String()))
							return
						}

						// save domain
						if err := s.write.Write(&spider.Domain{
							URL:    res.URL.String(),
							Name:   domain.Name,
							TLD:    domain.TLD,
							Status: status.String(),
						}); err != nil {
							s.log.Error(err.Error(), flog.NewField("domain", root), flog.NewField("url", res.URL.String()))
							return
						}

						// terminal print
						switch status {
						case domaincheck.Available:
							fmt.Printf("%s %s\n", green("available"), root)
						case domaincheck.Registered:
							fmt.Printf("%s %s\n", red("taken"), root)
						default:
							fmt.Printf("%s %s\n", red("unknown"), root)
						}
					}()
				}
			}
		}()
	}

	s.wg.Wait()
	return nil
}

func (s *Spider) Shutdown() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	<-ctx.Done()
	clog.Info("shutting down ...")

	// 2nd signal kills program
	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		defer stop()
		<-ctx.Done()
		clog.Info("killing program ...")
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
