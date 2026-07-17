package reviews

// ReviewsIngestionArgs is enqueued by catalog_ingestion, one per product.
// Idempotency key per execution_phase 2.3: product_id + source_batch_date.
type ReviewsIngestionArgs struct {
	ProductID       string `json:"product_id"`
	SourceASIN      string `json:"source_asin"`
	SourceBatchDate string `json:"source_batch_date"`
}

func (ReviewsIngestionArgs) Kind() string { return "reviews_ingestion" }
