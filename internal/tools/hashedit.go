package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type HashEdit struct {
	lineHashes map[int]string
}

func NewHashEdit() *HashEdit {
	return &HashEdit{lineHashes: make(map[int]string)}
}

func (h *HashEdit) Compute(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func (h *HashEdit) LineHash(line string) string {
	return h.Compute(line)[:8]
}

func (h *HashEdit) StoreLineHash(lineNum int, hash string) {
	h.lineHashes[lineNum] = hash
}

func (h *HashEdit) HasLineHash(lineNum int) bool {
	_, ok := h.lineHashes[lineNum]
	return ok
}

func (h *HashEdit) GetLineHash(lineNum int) string {
	return h.lineHashes[lineNum]
}

func (h *HashEdit) SurgicalEdit(content string, lineNum int, newLine string) (string, error) {
	lines := strings.Split(content, "\n")
	if lineNum < 1 || lineNum > len(lines) {
		return content, nil
	}

	lines[lineNum-1] = newLine
	return strings.Join(lines, "\n"), nil
}

func (h *HashEdit) Verify(content string) bool {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		expected := h.lineHashes[i+1]
		if expected != "" && h.LineHash(line) != expected {
			return false
		}
	}
	return true
}

func (h *HashEdit) Rollback(lineNum int) string {
	return h.lineHashes[lineNum]
}