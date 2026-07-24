package reviews

import "github.com/riverqueue/river"

// ReviewsIngestionArgs — ONE job per category run, not one per product.
//
// This replaces the earlier per-product design (ProductID/SourceASIN/
// SourceBatchDate) that catalog_ingestion used to enqueue inside its
// per-record loop. That design meant re-scanning the full reviews file
// once per product — O(products x file size) — which does not finish
// at 50k-product scale. This job streams the reviews file exactly once
// and aggregates all known products from it in a single pass.
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
