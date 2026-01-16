package pagination

import (
	"bytes"

	"github.com/opentelekomcloud/gophertelekomcloud/internal/extract"
)

// PageWithInfo is a page with marker information inside `page_info`
type PageWithInfo struct {
	MarkerPageBase
}

type pageInfo struct {
	PreviousMarker string `json:"previous_marker"`
	NextMarker     string `json:"next_marker"`
	CurrentCount   int    `json:"current_count"`
}

func (p PageWithInfo) LastMarker() (string, error) {
	var info pageInfo
	err := extract.IntoStructPtr(bytes.NewReader(p.Body), &info, "page_info")
	if err != nil {
		return "", err
	}
	return info.NextMarker, nil
}

// NextPageURL generates the URL for the page of results after this one.
func (p PageWithInfo) NextPageURL() (string, error) {
	currentURL := p.URL

	mark, err := p.Owner.LastMarker()
	if err != nil {
		return "", err
	}
	if mark == "" {
		return "", nil
	}

	q := currentURL.Query()
	q.Set("marker", mark)
	currentURL.RawQuery = q.Encode()

	return currentURL.String(), nil
}

func NewPageWithInfo(r PageResult) PageWithInfo {
	p := PageWithInfo{MarkerPageBase: MarkerPageBase{
		PageResult: r,
	}}
	p.Owner = &p
	return p
}
