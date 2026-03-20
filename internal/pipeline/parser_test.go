package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Y4NN777/doc-intel/internal/domain"
)

func TestPDFParser_Extract_FileNotFound(t *testing.T) {
	parser := NewPDFParser()
	
	_, _, err := parser.Extract("/nonexistent/file.pdf")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestPDFParser_DetectLanguage_English(t *testing.T) {
	parser := NewPDFParser()
	
	pages := []PageText{
		{
			PageNumber: 1,
			Text:       "The quick brown fox jumps over the lazy dog. This is a test of the English language detection.",
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	lang := parser.detectLanguage(pages)
	if lang != domain.LanguageEN {
		t.Errorf("expected English, got %s", lang)
	}
}

func TestPDFParser_DetectLanguage_French(t *testing.T) {
	parser := NewPDFParser()
	
	pages := []PageText{
		{
			PageNumber: 1,
			Text:       "Le renard brun rapide saute par-dessus le chien paresseux. C'est un test de la détection de la langue française.",
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	lang := parser.detectLanguage(pages)
	if lang != domain.LanguageFR {
		t.Errorf("expected French, got %s", lang)
	}
}

func TestPDFParser_DetectLanguage_Unknown(t *testing.T) {
	parser := NewPDFParser()
	
	pages := []PageText{
		{
			PageNumber: 1,
			Text:       "12345 67890 !@#$%",
			Source:     domain.ChunkSourceTextLayer,
		},
	}
	
	lang := parser.detectLanguage(pages)
	if lang != domain.LanguageUnknown {
		t.Errorf("expected Unknown, got %s", lang)
	}
}

// Integration test - requires a real PDF file
func TestPDFParser_Extract_RealPDF(t *testing.T) {
	// Skip if no test PDF available
	testPDF := filepath.Join("testdata", "sample.pdf")
	if _, err := os.Stat(testPDF); os.IsNotExist(err) {
		t.Skip("Skipping integration test: no test PDF available")
	}
	
	parser := NewPDFParser()
	
	pages, lang, err := parser.Extract(testPDF)
	if err != nil {
		t.Fatalf("failed to extract PDF: %v", err)
	}
	
	if len(pages) == 0 {
		t.Error("expected at least one page")
	}
	
	if lang == domain.LanguageUnknown {
		t.Log("Warning: language detection returned unknown")
	}
	
	for _, page := range pages {
		if page.PageNumber < 1 {
			t.Errorf("invalid page number: %d", page.PageNumber)
		}
	}
}
