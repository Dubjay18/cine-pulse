package storage

type Content struct {
	Title     string   `json:"title"`
	Year      *int     `json:"year,omitempty"`
	Category  string   `json:"category"`
	ExtraInfo string   `json:"extra_info"`
	Type      string   `json:"type"` // "movie" or "series"
	Rating    *float64 `json:"rating,omitempty"`
	SourceURL *string  `json:"source_url,omitempty"`
}
