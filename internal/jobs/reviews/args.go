package reviews

type ReviewsIngestionArgs struct {
	ProducedID      string `json:"product_id"`
	SourceAsin      string `json:"source_asin"`
	SourceBatchDate string `json:"source_batch_date"`
}

func (ReviewsIngestionArgs) kind() string { return "reviews_ingestion" }
func (ReviewsIngestionArgs) Kind() string { return "reviews_ingestion" }
