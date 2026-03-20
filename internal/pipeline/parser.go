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

// extractTextLayer extracts text from PDF text layer with proper spacing
func (p *PDFParser) extractTextLayer(path string) ([]PageText, error) {
	// Open PDF file
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()
	
	totalPages := r.NumPage()
	pages := make([]PageText, 0, totalPages)
	
	// Extract text from each page using styled text for better spacing
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		
		// Try to get styled text with positions for better spacing
		texts := page.Content().Text
		if len(texts) == 0 {
			pages = append(pages, PageText{
				PageNumber: pageNum,
				Text:       "",
				Source:     domain.ChunkSourceTextLayer,
			})
			continue
		}
		
		// Reconstruct text with proper spacing based on positions
		var result strings.Builder
		var lastX, lastY float64
		lastX = -1000 // Initialize to impossible value
		
		for i, t := range texts {
			if i == 0 {
				result.WriteString(t.S)
				lastX = t.X + float64(len(t.S))*t.FontSize*0.5 // Approximate end position
				lastY = t.Y
			} else {
				// Add space if there's a horizontal gap or line break
				xGap := t.X - lastX
				yGap := t.Y - lastY
				
				// New line if Y position changed significantly
				if yGap > t.FontSize*0.3 || yGap < -t.FontSize*0.3 {
					result.WriteString("\n")
				} else if xGap > t.FontSize*0.2 { // Horizontal gap suggests space
					result.WriteString(" ")
				}
				
				result.WriteString(t.S)
				lastX = t.X + float64(len(t.S))*t.FontSize*0.5
				lastY = t.Y
			}
		}
		
		pages = append(pages, PageText{
			PageNumber: pageNum,
			Text:       strings.TrimSpace(result.String()),
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
	
	// Simple heuristic: count common words using word boundaries
	frenchWords := []string{"le", "la", "les", "de", "des", "du", "un", "une", "et", "est", "sont", "dans", "pour", "que", "qui", "avec", "par", "sur", "ce", "cette", "ces", "il", "elle", "nous", "vous", "ils", "elles", "mais", "ou", "donc", "car", "si", "ne", "pas", "plus", "tout", "tous", "toute", "toutes", "mon", "ma", "mes", "ton", "ta", "tes", "son", "sa", "ses"}
	englishWords := []string{"the", "a", "an", "and", "is", "are", "in", "to", "of", "for", "that", "this", "with", "on", "at", "by", "from", "it", "he", "she", "we", "you", "they", "but", "or", "if", "not", "all", "my", "your", "his", "her", "our", "their", "can", "will", "would", "should", "could", "have", "has", "had", "do", "does", "did", "be", "been", "being"}
	
	frenchCount := 0
	englishCount := 0
	
	// Split text into words and check for matches
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'à' && r <= 'ÿ') || r == '\'' || r == '-')
	})
	
	wordSet := make(map[string]bool)
	for _, word := range words {
		wordSet[word] = true
	}
	
	for _, word := range frenchWords {
		if wordSet[word] {
			frenchCount++
		}
	}
	
	for _, word := range englishWords {
		if wordSet[word] {
			englishCount++
		}
	}
	
	// Determine language based on counts
	if frenchCount > englishCount && frenchCount >= 3 {
		return domain.LanguageFR
	} else if englishCount > frenchCount && englishCount >= 3 {
		return domain.LanguageEN
	}
	
	return domain.LanguageUnknown
}
