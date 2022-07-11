package spider

// Domain
type Domain struct {
	URL    string
	Name   string
	TLD    string
	Status string
}

// CSVRow
func (d Domain) CSVRow() []string {
	var row []string
	return append(row, d.URL, d.Name, d.TLD, d.Status)
}
