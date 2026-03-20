// +build !ocr

package pipeline

import (
	"fmt"
	"os"
	"strings"

	"github.com/Y4NN777/doc-intel/internal/domain"
	"github.com/ledongthuc/pdf"
)

// Parser extracts text from PDF files
type Parser interface {
	Extract(path string) ([]PageText, domain.Language, error)
}

// PageText represents extracted text from a single page
type PageText struct {
	PageNumber int
	Text       string
	Source     domain.ChunkSource
}

// PDFParser implements Parser using ledongthuc/pdf for text extraction (no OCR)
type PDFParser struct {
	// Pure Go implementation - no OCR support
}

// NewPDFParser creates a new PDF parser without OCR support
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// Extract extracts text from a PDF file
func (p *PDFParser) Extract(path string) ([]PageText, domain.Language, error) {
	// Validate file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, domain.LanguageUnknown, fmt.Errorf("file not found: %s", path)
	}

	// Extract text from PDF
	pages, err := p.extractTextLayer(path)
	if err != nil {
		return nil, domain.LanguageUnknown, fmt.Errorf("failed to extract text: %w", err)
	}

	if len(pages) == 0 {
		return nil, domain.LanguageUnknown, fmt.Errorf("no pages extracted from PDF")
	}

	// Detect language from extracted text
	lang := p.detectLanguage(pages)

	return pages, lang, nil
}

// extractTextLayer extracts text from PDF text layer
func (p *PDFParser) extractTextLayer(path string) ([]PageText, error) {
	// Open PDF file
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()
	
	totalPages := r.NumPage()
	pages := make([]PageText, 0, totalPages)
	
	// Extract text from each page
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Log warning but continue with empty text
			text = ""
		}
		
		pages = append(pages, PageText{
			PageNumber: pageNum,
			Text:       strings.TrimSpace(text),
			Source:     domain.ChunkSourceTextLayer,
		})
	}
	
	return pages, nil
}

// detectLanguage detects the primary language from extracted text
func (p *PDFParser) detectLanguage(pages []PageText) domain.Language {
	// Combine first few pages for language detection
	var sample strings.Builder
	maxPages := 3
	if len(pages) < maxPages {
		maxPages = len(pages)
	}
	
	for i := 0; i < maxPages; i++ {
		sample.WriteString(pages[i].Text)
		sample.WriteString(" ")
	}
	
	text := strings.ToLower(sample.String())
	
	// Simple heuristic: count common words
	frenchWords := []string{"le", "la", "les", "de", "des", "un", "une", "et", "est", "dans", "pour", "que", "qui"}
	englishWords := []string{"the", "a", "an", "and", "is", "in", "to", "of", "for", "that", "this", "with"}
	
	frenchCount := 0
	englishCount := 0
	
	for _, word := range frenchWords {
		frenchCount += strings.Count(text, " "+word+" ")
	}
	
	for _, word := range englishWords {
		englishCount += strings.Count(text, " "+word+" ")
	}
	
	// Determine language based on counts
	if frenchCount > englishCount && frenchCount > 5 {
		return domain.LanguageFR
	} else if englishCount > frenchCount && englishCount > 5 {
		return domain.LanguageEN
	}
	
	return domain.LanguageUnknown
}
