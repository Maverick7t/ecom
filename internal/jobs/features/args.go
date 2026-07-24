package features

// FeatureGenerationArgs — Phase 2.4, not yet built. This file exists
// only so reviews_ingestion can reference the job Kind() when enqueuing
// downstream work. There is no features.Worker yet, so any Insert() of
// this args type will fail with "job kind is not registered" until
// Phase 2.4 is implemented — that failure is caught and logged as
// non-fatal in reviews_ingestion.go, matching the existing catalog ->
// reviews handoff pattern.
type FeatureGenerationArgs struct {
	ProductID string `json:"product_id"`
}

func (FeatureGenerationArgs) Kind() string { return "feature_generation" }
