package reviews

type ReviewsIngestionArgs struct {
	ProducedID string `json:"product_id"`
	SourceAsin string `json:"source_asin"`
	SourceBatchDate string `json:"source_batch_date"`
}

func (ReviewIngestinArgs) kind() string { return "reviews_ingestion"}
}