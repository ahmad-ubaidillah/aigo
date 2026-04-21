package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type Dedupe struct {
	mu       sync.RWMutex
	hashes   map[string]string
	contents map[string]bool
}

func NewDedupe() *Dedupe {
	return &Dedupe{
		hashes:   make(map[string]string),
		contents: make(map[string]bool),
	}
}

func (d *Dedupe) Hash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func (d *Dedupe) Add(content, hash string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.hashes[hash] = content
	d.contents[content] = true
}

func (d *Dedupe) IsDuplicate(content string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.contents[content]
}

func (d *Dedupe) RemoveDuplicate(content string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	hash := d.Hash(content)
	delete(d.hashes, hash)
	delete(d.contents, content)
}

func (d *Dedupe) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.hashes)
}