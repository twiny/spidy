## Spidy
Spidy is a tool that crawl web pages from a given list of websites, it match all domains on each page and find expired domains among them.

# Usage
compile the package
`
go build .
`
then run
`
./Spidy -config /path/to/setting.yaml
`

# Output/Results
results will be saved in ./log folder:

  errors.txt: errors while scraping will be stored here. helpful for debugging.
  
  domains.txt: list of all unique domain checked.

  found.txt: list of the available domains found.

  visited.txt: list of all unique visited URLs.


## Engine Setting:
main app setting:

  **- worker :number of threads**

  example: worker:10 => scrap 10 urls at once.

  **- depth: page scraping depth**

  example: depth:5 => visit the link from
  the 1st page and follow link found in 2nd page
  till the 5th page

  **- parallel: number of processor**

  example: parallel:5 => on the scraped page process
  5 link at once.

  **- urls: path to a .txt file.**

  path to the input.txt which will have a URLs 
  a new URL in each line.

  **- proxies: an array of proxy. accepts only HTTP proxies.**

  if no proxy is added. proxy scraping will be disabled.
  if one proxy is added. all scraping will be through one proxy.
  if more then two proxies added. scraping will be rotated.
  example:

  proxies: ["http://username:password@1.1.1.1:2345","http://username:password1.1.1.1:2345","http://username:password1.1.1.1:2345"]

  to disable able proxy, use empty array, like:
  proxies: []


  **- tlds: an array of tld.**

  example: [com, net, org]

  an empty array will match all the 122 TLD in crawler/tld.go

  **- random_delay: time duration**

  a random time duration between requests
  example: 10s

  **- timeout: time duration**

  set timeout for HTTP requests
  example: 60s

  # Big Thanks
  Colly V2 => https://github.com/gocolly/colly

[![Donate with Ethereum](https://en.cryptobadges.io/badge/small/0x94a003520Ad7F9aFF613c1cb6798a96256217EC9)](https://en.cryptobadges.io/donate/0x94a003520Ad7F9aFF613c1cb6798a96256217EC9)
