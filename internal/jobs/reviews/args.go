package reviews

import "github.com/riverqueue/river"

// ReviewsIngestionArgs — one job per category run, not per product.
// See design correction: streaming the reviews file once and grouping
// in-memory, rather than re-scanning per product.
type ReviewsIngestionArgs struct {
	Category   string `json:"category"`
	SourcePath string `json:"source_path"` // local path to reviews jsonl.gz
}

func (ReviewsIngestionArgs) Kind() string { return "reviews_ingestion" }

func (ReviewsIngestionArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: 10}
}
