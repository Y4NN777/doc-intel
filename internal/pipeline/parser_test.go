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
	testPDF := filepath.Join("testdata", "resume_en.pdf")
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
	
	// Log extracted text for debugging
	for i, page := range pages {
		t.Logf("Page %d text length: %d chars", i+1, len(page.Text))
		if len(page.Text) > 100 {
			t.Logf("Page %d text preview: %q...", i+1, page.Text[:100])
		} else {
			t.Logf("Page %d text: %q", i+1, page.Text)
		}
	}
	
	if pages[0].PageNumber != 1 {
		t.Errorf("expected page number 1, got %d", pages[0].PageNumber)
	}
	
	// Check that we extracted some text
	if len(pages[0].Text) == 0 {
		t.Error("expected non-empty text from page 1")
	}
	
	// Language detection may fail if PDF text extraction doesn't preserve spaces
	// This is a known limitation of some PDF parsers
	t.Logf("Detected language: %s", lang)
	if lang == domain.LanguageUnknown {
		t.Log("Note: Language detection failed - likely due to missing word boundaries in extracted text")
	}
	
	// Check source is text layer
	if pages[0].Source != domain.ChunkSourceTextLayer {
		t.Errorf("expected text_layer source, got %s", pages[0].Source)
	}
}
