package reviews

import "github.com/riverqueue/river"

type ReviewsIngestionArgs struct {
	Category          string `json:"category"`
	ReviewsSourcePath string `json:"reviews_source_path"`
}

func (ReviewsIngestionArgs) Kind() string { return "reviews_ingestion" }

func (ReviewsIngestionArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 10,
		UniqueOpts:  river.UniqueOpts{ByArgs: true},
	}
}
