.PHONY: test test-store test-domain test-vectorindex test-short test-faiss build build-faiss run clean

# Run all tests with FTS5 support (pure Go vector index)
test:
	go test -v -tags fts5 ./...

# Run all tests with short timeout
test-short:
	go test -v -tags fts5 -timeout 30s ./...

# Run only store tests
test-store:
	go test -v -tags fts5 -timeout 30s ./internal/store

# Run only domain tests
test-domain:
	go test -v ./internal/domain

# Run only vectorindex tests
test-vectorindex:
	go test -v ./internal/vectorindex

# Run all tests with FAISS CGO support (requires FAISS C library)
test-faiss:
	go test -v -tags "fts5 faiss" ./...

# Build with pure Go vector index (default)
build:
	go build -tags fts5 -o bin/doc-intel ./cmd/doc-intel

# Build with FAISS CGO support (requires FAISS C library installed)
build-faiss:
	go build -tags "fts5 faiss" -o bin/doc-intel ./cmd/doc-intel

# Run the application
run: build
	./bin/doc-intel

# Clean build artifacts and test databases
clean:
	rm -rf bin/
	find . -name "*.db" -type f -delete
	find . -name "*.db-shm" -type f -delete
	find . -name "*.db-wal" -type f -delete

