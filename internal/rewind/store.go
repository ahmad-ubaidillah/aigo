package rewind

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// RewindEntry represents a stored content snapshot with metadata.
type RewindEntry struct {
	Hash           string
	FullHash       string
	Content        string
	ContentType    string
	OriginalSize   int
	CompressedSize int
	Timestamp      time.Time
	SessionID      string
}

// RewindStore is an in-memory archive of content by SHA-256 hash.
type RewindStore struct {
	entries map[string]*RewindEntry
	mu      sync.RWMutex
}

// NewRewindStore creates a new, empty rewind store.
func NewRewindStore() *RewindStore {
	return &RewindStore{entries: make(map[string]*RewindEntry)}
}

// Store saves the given content and returns an 8-char short hash.
// It computes the SHA-256 of the content and stores the entry in-memory.
func (s *RewindStore) Store(content, contentType, sessionID string) string {
	hash := sha256.Sum256([]byte(content))
	fullHash := hex.EncodeToString(hash[:])
	shortHash := fullHash[:8]

	entry := &RewindEntry{
		Hash:           shortHash,
		FullHash:       fullHash,
		Content:        content,
		ContentType:    contentType,
		OriginalSize:   len(content),
		CompressedSize: 0,
		Timestamp:      time.Now(),
		SessionID:      sessionID,
	}

	s.mu.Lock()
	s.entries[shortHash] = entry
	s.mu.Unlock()
	return shortHash
}

// Retrieve gets an entry by its 8-char short hash.
func (s *RewindStore) Retrieve(shortHash string) (*RewindEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.entries[shortHash]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("rewind entry not found for hash %s", shortHash)
}

// List returns all entries, optionally filtered by sessionID, sorted by time
// with newest first.
func (s *RewindStore) List(sessionID string) []RewindEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []RewindEntry
	for _, e := range s.entries {
		if sessionID == "" || e.SessionID == sessionID {
			res = append(res, *e)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.After(res[j].Timestamp)
	})
	return res
}

// Count returns the total number of stored entries.
func (s *RewindStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// Clear removes all stored entries from the in-memory archive.
func (s *RewindStore) Clear() {
	s.mu.Lock()
	s.entries = make(map[string]*RewindEntry)
	s.mu.Unlock()
}

// Show returns the full content for the given short hash.
func (s *RewindStore) Show(shortHash string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.entries[shortHash]; ok {
		return e.Content, nil
	}
	return "", fmt.Errorf("rewind entry not found for hash %s", shortHash)
}
