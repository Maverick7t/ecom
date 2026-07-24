package reviews

import (
	"bufio"
	"compress/gzip"
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/Maverick7t/ecom/internal/jobs/features"
	"github.com/Maverick7t/ecom/internal/platform/storage"
)

// rawReview mirrors the Amazon Reviews 2023 review-record schema.
//
// VERIFY this against your actual sample review record before trusting
// it — the same caveat that applied to the metadata schema. Notably:
// there is no total-vote-count field in this dataset, only helpful_vote.
// That means a true "helpfulness ratio" cannot be computed — see the
// placeholder helpfulnessScore calculation in persistOne below.
type rawReview struct {
	Rating           float64 `json:"rating"`
	Title            string  `json:"title"`
	Text             string  `json:"text"`
	ASIN             string  `json:"asin"`
	ParentASIN       string  `json:"parent_asin"`
	UserID           string  `json:"user_id"`
	Timestamp        int64   `json:"timestamp"` // assumed Unix ms — verify
	HelpfulVote      int     `json:"helpful_vote"`
	VerifiedPurchase bool    `json:"verified_purchase"`
}

// topHeap is a min-heap on HelpfulVote, capped at topN. Keeps the N most
// helpful reviews per product in memory without buffering every review
// for every product — required for a single-pass scan over 50k products'
// worth of reviews to stay memory-bounded.
type topHeap []rawReview

func (h topHeap) Len() int           { return len(h) }
func (h topHeap) Less(i, j int) bool { return h[i].HelpfulVote < h[j].HelpfulVote }
func (h topHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *topHeap) Push(x any)        { *h = append(*h, x.(rawReview)) }
func (h *topHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

const topN = 20

type reviewAgg struct {
	productID  string
	count      int
	ratingSum  float64
	dist       map[int]int
	firstTS    int64
	lastTS     int64
	helpfulSum int
	top20      *topHeap
}

func newAgg(productID string) *reviewAgg {
	th := &topHeap{}
	heap.Init(th)
	return &reviewAgg{
		productID: productID,
		dist:      map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0},
		top20:     th,
	}
}

func (a *reviewAgg) add(r rawReview) {
	a.count++
	a.ratingSum += r.Rating

	if star := int(r.Rating); star >= 1 && star <= 5 {
		a.dist[star]++
	}
	if a.firstTS == 0 || r.Timestamp < a.firstTS {
		a.firstTS = r.Timestamp
	}
	if r.Timestamp > a.lastTS {
		a.lastTS = r.Timestamp
	}
	a.helpfulSum += r.HelpfulVote

	if a.top20.Len() < topN {
		heap.Push(a.top20, r)
	} else if a.top20.Len() > 0 && (*a.top20)[0].HelpfulVote < r.HelpfulVote {
		heap.Pop(a.top20)
		heap.Push(a.top20, r)
	}
}

// topSortedDesc returns the aggregated top reviews ordered by HelpfulVote
// descending. The heap's backing slice already IS the top-N set — only
// the display order needs fixing.
func (a *reviewAgg) topSortedDesc() []rawReview {
	out := make([]rawReview, len(*a.top20))
	copy(out, *a.top20)
	sort.Slice(out, func(i, j int) bool { return out[i].HelpfulVote > out[j].HelpfulVote })
	return out
}

type Worker struct {
	river.WorkerDefaults[ReviewsIngestionArgs]
	db      *pgxpool.Pool
	storage storage.Storage
	logger  *slog.Logger
}

func NewWorker(db *pgxpool.Pool, st storage.Storage, logger *slog.Logger) *Worker {
	return &Worker{db: db, storage: st, logger: logger}
}

// Timeout: single pass over the full reviews file plus one storage
// upload and one DB transaction per known product. Same rationale as
// catalog_ingestion.Timeout — River's client default will not survive
// this at 50k-product scale.
func (w *Worker) Timeout(job *river.Job[ReviewsIngestionArgs]) time.Duration {
	return 60 * time.Minute
}

func (w *Worker) Work(ctx context.Context, job *river.Job[ReviewsIngestionArgs]) error {
	args := job.Args
	if strings.TrimSpace(args.Category) == "" {
		return fmt.Errorf("reviews_ingestion: category is required")
	}
	if strings.TrimSpace(args.ReviewsSourcePath) == "" {
		return fmt.Errorf("reviews_ingestion: reviews_source_path is required")
	}

	syncRunID, err := w.createSyncRun(ctx)
	if err != nil {
		return fmt.Errorf("create sync_run: %w", err)
	}

	knownProducts, err := w.loadKnownProducts(ctx, args.Category)
	if err != nil {
		return w.failRun(ctx, syncRunID, fmt.Errorf("load known products: %w", err))
	}
	w.logger.Info("reviews_ingestion started",
		slog.String("category", args.Category),
		slog.Int("known_products", len(knownProducts)))

	aggs, linesScanned, err := w.streamAndAggregate(args.ReviewsSourcePath, knownProducts)
	if err != nil {
		return w.failRun(ctx, syncRunID, fmt.Errorf("stream reviews: %w", err))
	}

	processed := 0
	for parentASIN, agg := range aggs {
		if err := w.persistOne(ctx, agg); err != nil {
			w.logger.Error("persist review agg failed",
				slog.String("parent_asin", parentASIN), slog.Any("error", err))
			continue // one bad product should not fail the whole run
		}
		processed++
	}

	w.logger.Info("reviews_ingestion completed",
		slog.String("category", args.Category),
		slog.Int("products_with_reviews", processed),
		slog.Int("products_with_no_reviews", len(knownProducts)-len(aggs)))

	return w.completeRun(ctx, syncRunID, linesScanned, processed)
}

// loadKnownProducts restricts aggregation to products this ingestion run
// actually created — reviews for parent_asins outside the current
// category ingestion are skipped, not aggregated into nothing.
func (w *Worker) loadKnownProducts(ctx context.Context, category string) (map[string]string, error) {
	rows, err := w.db.Query(ctx, `
		SELECT p.parent_asin, p.id::text
		FROM products p
		JOIN product_categories pc ON pc.product_id = p.id
		JOIN categories c ON c.id = pc.category_id
		WHERE c.slug = $1 OR c.name = $1
	`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var asin, id string
		if err := rows.Scan(&asin, &id); err != nil {
			return nil, err
		}
		out[asin] = id
	}
	return out, rows.Err()
}

func (w *Worker) streamAndAggregate(path string, known map[string]string) (map[string]*reviewAgg, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open reviews file: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, 0, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // review text can run long

	aggs := make(map[string]*reviewAgg, len(known))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		var r rawReview
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			w.logger.Warn("skipping malformed review line", slog.Int("line", lineNum))
			continue
		}
		productID, ok := known[r.ParentASIN]
		if !ok {
			continue // review belongs to a product outside this run
		}
		agg, ok := aggs[r.ParentASIN]
		if !ok {
			agg = newAgg(productID)
			aggs[r.ParentASIN] = agg
		}
		agg.add(r)
	}
	if err := scanner.Err(); err != nil {
		return nil, lineNum, fmt.Errorf("scan reviews file: %w", err)
	}
	return aggs, lineNum, nil
}

func (w *Worker) persistOne(ctx context.Context, agg *reviewAgg) error {
	top := agg.topSortedDesc()
	blob, err := json.Marshal(top)
	if err != nil {
		return fmt.Errorf("marshal top20: %w", err)
	}
	storagePath := fmt.Sprintf("reviews/top20/%s.json", agg.productID)

	uploadCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := w.storage.Upload(uploadCtx, storagePath, blob, "application/json"); err != nil {
		return fmt.Errorf("upload top20: %w", err)
	}

	avgRating := 0.0
	if agg.count > 0 {
		avgRating = agg.ratingSum / float64(agg.count)
	}

	// review_velocity: reviews/month across the DATASET's own review
	// window (first review -> last review), not against wall-clock now.
	// This is a static historical dataset — measuring against "now"
	// would silently shrink velocity further every day the pipeline
	// sits unrun, which is meaningless for a portfolio snapshot.
	monthsSpan := 1.0
	if agg.lastTS > agg.firstTS {
		days := time.UnixMilli(agg.lastTS).Sub(time.UnixMilli(agg.firstTS)).Hours() / 24
		if m := days / 30.44; m > 1 {
			monthsSpan = m
		}
	}
	velocity := float64(agg.count) / monthsSpan

	// PLACEHOLDER SEMANTICS: no total-vote field exists in this dataset,
	// so this is a capped normalization of avg helpful_vote per review —
	// NOT a true ratio. Confirm this is acceptable before feature_
	// generation depends on it. Consider a migration renaming
	// helpfulness_ratio -> helpfulness_score so the column name stops
	// implying a numerator/denominator that doesn't exist.
	avgHelpful := 0.0
	if agg.count > 0 {
		avgHelpful = float64(agg.helpfulSum) / float64(agg.count)
	}
	const helpfulCap = 50.0
	helpfulnessScore := avgHelpful / helpfulCap
	if helpfulnessScore > 1.0 {
		helpfulnessScore = 1.0
	}

	distJSON, err := json.Marshal(agg.dist)
	if err != nil {
		return fmt.Errorf("marshal rating distribution: %w", err)
	}

	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO product_reviews_meta
			(product_id, review_count, avg_rating, rating_distribution,
			 review_velocity, helpfulness_ratio, raw_reviews_storage_path, computed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (product_id) DO UPDATE SET
			review_count             = EXCLUDED.review_count,
			avg_rating               = EXCLUDED.avg_rating,
			rating_distribution      = EXCLUDED.rating_distribution,
			review_velocity          = EXCLUDED.review_velocity,
			helpfulness_ratio        = EXCLUDED.helpfulness_ratio,
			raw_reviews_storage_path = EXCLUDED.raw_reviews_storage_path,
			computed_at              = NOW()
	`, agg.productID, agg.count, avgRating, distJSON, velocity, helpfulnessScore, storagePath); err != nil {
		return fmt.Errorf("upsert product_reviews_meta: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE products SET avg_rating = $1, review_count = $2, updated_at = NOW()
		WHERE id = $3
	`, avgRating, agg.count, agg.productID); err != nil {
		return fmt.Errorf("update product aggregates: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	// Enqueue feature_generation — non-fatal, matching the existing
	// catalog -> reviews handoff pattern. features.Worker is not built
	// yet (Phase 2.4), so expect "job kind not registered" warnings
	// here until it exists. That is expected, not a bug.
	if client, ok := river.ClientFromContext[pgx.Tx](ctx); ok {
		if _, err := client.Insert(ctx, features.FeatureGenerationArgs{
			ProductID: agg.productID,
		}, nil); err != nil {
			w.logger.Warn("failed to enqueue feature_generation (non-fatal, expected until Phase 2.4)",
				slog.String("product_id", agg.productID), slog.Any("error", err))
		}
	}

	return nil
}

func (w *Worker) createSyncRun(ctx context.Context) (string, error) {
	var id string
	err := w.db.QueryRow(ctx, `
		INSERT INTO sync_runs (job_type, status) VALUES ('reviews_ingestion', 'RUNNING')
		RETURNING id::text
	`).Scan(&id)
	return id, err
}

func (w *Worker) completeRun(ctx context.Context, syncRunID string, linesScanned, productsProcessed int) error {
	_, err := w.db.Exec(ctx, `
		UPDATE sync_runs
		SET status = 'COMPLETED', records_in = $2, records_out = $3, completed_at = NOW()
		WHERE id = $1::uuid
	`, syncRunID, linesScanned, productsProcessed)
	return err
}

func (w *Worker) failRun(ctx context.Context, syncRunID string, cause error) error {
	if _, err := w.db.Exec(ctx, `
		UPDATE sync_runs
		SET status = 'FAILED', error_message = $2, completed_at = NOW()
		WHERE id = $1::uuid
	`, syncRunID, cause.Error()); err != nil {
		w.logger.Error("failed to mark sync_run FAILED", slog.Any("error", err))
	}
	return cause
}
