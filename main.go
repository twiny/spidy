package main

import (
	"Spidy/crawler"
	"flag"
	"fmt"
	"os"

	"github.com/pkg/profile"
)

const msg = `
Spidy - Fast Bulk Expired Domain Scraper

Usage:
./Spidy -config /path/to/setting.yaml
`

// usage
func usage() {
	print(msg)
	os.Exit(0)
}

// main
func main() {
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	// go func() {
	// 	p := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	// 	time.Sleep(30 * time.Minute)
	// 	p.Stop()
	// 	os.Exit(1)
	// }()
	// load configs
	config := flag.String("config", "", "config: path to setting.yaml.")
	flag.Usage = usage
	flag.Parse()
	//
	if len(*config) == 0 {
		fmt.Println("enter config path.")
		return
	}
	//
	fmt.Println("Welcome, Spidy is running.")
	//
	setting, err := crawler.SettingFromFile(*config)
	if err != nil {
		fmt.Println(err)
		return
	}

	// make directory for cache & log if not exisit
	if _, err := os.Stat("./log"); os.IsNotExist(err) {
		os.Mkdir("./log", os.ModePerm)
	}
	//
	tool, err := crawler.NewSpider(setting)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Logger
	tool.Logger()

	// Run
	tool.Run()

	// finialy
	fmt.Println("done :)")
}
