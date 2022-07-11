package spider

// Writer
type Writer interface {
	Write(*Domain) error
}
