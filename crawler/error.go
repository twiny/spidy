package crawler

import (
	"errors"
	"fmt"
	"time"
)

var (
	// Errors
	ErrBadUrl         = errors.New("bad url format")
	ErrUnsupportedTLD = errors.New("unsupported TLD")
	//
	OnTimeParse     = "OnTimeParse"
	OnProxyConnect  = "OnProxyConnect"
	OnDomainExtract = "OnDomainExtract"
	OnScrap         = "OnScrapping"
	OnWhois         = "OnWhois"
	OnDomain        = "OnDomain"
	OnMoz           = "OnMoz"
	OnTextExtract   = "OnTextExtract"
	//
	OnProcessor = "OnProcessor"
	OnStorage   = "OnStorage"
	OnPool      = "OnPool"
	OnBodyRead  = "OnBodyRead"
)

// Error struct
type Error struct {
	time   time.Time
	on     string
	url    string
	domain string
	err    error
}

/*
Error method
*/
func (e *Error) Error() string {
	return fmt.Sprint(e.time.Format(time.RFC822) + " - " + e.on + " - " + "URL: " + e.url + " - " + "Domain: " + e.domain + " - " + "Error: " + e.err.Error())
}
