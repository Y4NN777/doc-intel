package domain

import "time"

// DocStatus represents the processing state of a document
type DocStatus string

const (
	DocStatusPending DocStatus = "pending"
	DocStatusIndexed DocStatus = "indexed"
	DocStatusFailed  DocStatus = "failed"
)

// Language represents supported document languages
type Language string

const (
	LanguageEN      Language = "en"
	LanguageFR      Language = "fr"
	LanguageUnknown Language = "unknown"
)

// Document represents a PDF file in a workspace
type Document struct {
	ID          string
	WorkspaceID string
	Path        string
	Language    Language
	Status      DocStatus
	PageCount   int
	IsRead      bool
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// IsPending returns true if document is awaiting processing
func (d *Document) IsPending() bool {
	return d.Status == DocStatusPending
}

// IsIndexed returns true if document has been successfully processed
func (d *Document) IsIndexed() bool {
	return d.Status == DocStatusIndexed
}
