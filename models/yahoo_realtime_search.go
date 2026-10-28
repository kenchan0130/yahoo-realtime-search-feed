package models

import (
	"fmt"
	"strconv"
	"time"
)

type unixTimestamp struct {
	time.Time
}

func (t *unixTimestamp) UnmarshalJSON(data []byte) error {
	v, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("strconv.Atoi(): %v", err)
	}
	t.Time = time.Unix(int64(v), 0)

	return nil
}

type PageData struct {
	Timeline    *Timeline    `json:"timeline,omitempty"`
	SearchError *SearchError `json:"searchError,omitempty"`
}

type SearchError struct {
	// ErrorType is e.g. "zeromatch" or "longQuery", and empty when there is no error.
	ErrorType *string `json:"errorType,omitempty"`
}

type PageProps struct {
	PageData *PageData `json:"pageData,omitempty"`
}

type Props struct {
	PageProps *PageProps `json:"pageProps,omitempty"`
}

type Timeline struct {
	Entry *[]TimelineEntry `json:"entry,omitempty"`
}

type TimelineEntry struct {
	ID                   *string        `json:"id,omitempty"`
	URL                  *string        `json:"url,omitempty"`
	DisplayText          *string        `json:"displayText,omitempty"`
	DisplayTextBody      *string        `json:"displayTextBody,omitempty"`
	DisplayTextFragments *string        `json:"displayTextFragments,omitempty"`
	CreatedAt            *unixTimestamp `json:"createdAt,omitempty"`
	ScreenName           *string        `json:"screenName,omitempty"`
}
