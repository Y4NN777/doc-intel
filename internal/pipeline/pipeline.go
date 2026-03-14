package pipeline

// Pipeline orchestrates document ingestion
// Enforces INV-03: atomic ingestion, INV-05: complete reset on reprocess
type Pipeline interface {
	// Ingest processes all pending documents in a workspace
	Ingest(workspaceID string) (*Report, error)

	// Reprocess purges and re-indexes a specific document
	Reprocess(documentID string) (*Report, error)
}

// Report summarizes ingestion results
type Report struct {
	DocumentsProcessed int
	DocumentsFailed    int
	TotalChunks        int
	Errors             []error
}
