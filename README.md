## Spidy
A tool that crawl websites to find domain names and checks thier availiabity.

### Install

```sh
git clone https://github.com/twiny/spidy.git
cd ./spidy

# build
go build -o bin/spidy -v cmd/spidy/main.go

# run
./bin/spidy -c config/config.yaml -u https://github.com
```

## Configuration

```yaml
# main crawler config
crawler:
    max_depth: 10 # max depth of pages to visit per website.
    # filter: [] # regexp filter
    rate_limit: "1/5s" # 1 request per 5 sec
    max_body_size: "20MB" # max page body size
    user_agents: # array of user-agents
      - "Spidy/2.1; +https://github.com/ twiny/spidy"
    # proxies: [] # array of proxy. http(s), SOCKS5
# Logs
log:
    rotate: 7 # log rotation
    path: "./log" # log directory
# Store
store:
    ttl: "24h" # keep cache for 24h 
    path: "./store" # store directory
# Results
result:
    path: ./result # result directory
parralle: 3 # number of concurrent workers 
timeout: "5m" # request timeout
tlds: ["biz", "cc", "com", "edu", "info", "net", "org", "tv"] # array of domain extension to check.
```


## TODO

- [ ] Add support to more `writers`.
- [ ] Add terminal logging.
- [ ] Add test cases.

## Issues

NOTE: This package is provided "as is" with no guarantee. Use it at your own risk and always test it yourself before using it in a production environment. If you find any issues, please create a new issue.