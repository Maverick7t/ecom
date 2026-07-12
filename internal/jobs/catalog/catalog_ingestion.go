package catalog


import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
 
	"github.com/Maverick7t/ecom/internal/jobs/reviews"
	"github.com/Maverick7t/ecom/internal/platform/database/dbgen"
)

type CatalogIngestionJob struct {
	SourcePath string 'json:"source_path"'
	Category string 'json:"category"'
	Limit int 'json':"limit"'


func (CatalogIngestionArgs) King() string { return "catalog_ingestion"}
type metadataRecord struct {
	MainCategory string 'json:"main_category"'
	Categories []string 'json:"categories"'
	Title string 'json:"title"'
	Store string 'json:"store"'
	Description []string 'json:"description"'
	Price *string 'json:"price"'
	ParentAsin string 'json:"parent_asin"'
	Images []struct {
		Large string 'json:"large"'
	} 'json:"images"'
}