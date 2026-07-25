package reviews

import "github.com/riverqueue/river"

// ReviewsIngestionArgs — ONE job per catalog_ingestion run, not one per
// product and not scoped by category name.
//
// Earlier versions scoped this job by category name (`Category string`)
// and matched it against categories.slug/name in the DB. That was a real
// bug: catalog_ingestion filters metadata records against the DATASET's
// category taxonomy (e.g. "Electronics", the top-level main_category),
// but products get linked only to their LEAF category (e.g. "USB
// Cables") via resolveCategoryChain. Those two strings never match, so
// loadKnownProducts returned zero rows and every review was silently
// skipped — the job "completed" with nothing aggregated.
//
// Fix: catalog_ingestion already knows exactly which parent_asins it
// just upserted. Pass that list directly instead of trying to
// reconstruct the set from a category name.
type ReviewsIngestionArgs struct {
	ParentASINs       []string `json:"parent_asins"`
	ReviewsSourcePath string   `json:"reviews_source_path"`
}

func (ReviewsIngestionArgs) Kind() string { return "reviews_ingestion" }

func (ReviewsIngestionArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 10,
		UniqueOpts:  river.UniqueOpts{ByArgs: true},
	}
}
