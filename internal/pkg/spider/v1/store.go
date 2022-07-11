package spider

// Storage
type Storage interface {
	HasChecked(name string) bool
	Close() error
}
