package context

import (
	"path/filepath"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func BenchmarkBuildPrompt(b *testing.B) {
	dir := b.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, _ := memory.NewSessionDB(dbPath)
	defer db.Close()

	engine := NewContextEngine(db, types.Config{
		Memory: types.MemoryConfig{
			TokenBudget: 8000,
			MaxL0Items:  20,
			MaxL1Items:  50,
		},
	})
	engine.SetTaskGoal("Build a web application")

	for i := 0; i < 50; i++ {
		engine.AddL0("tool output line " + string(rune('a'+i%26)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.BuildPrompt("some task")
	}
}

func BenchmarkCompress(b *testing.B) {
	dir := b.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, _ := memory.NewSessionDB(dbPath)
	defer db.Close()

	engine := NewContextEngine(db, types.Config{
		Memory: types.MemoryConfig{
			TokenBudget:  8000,
			MaxL0Items:   20,
			MaxL1Items:   50,
			AutoCompress: true,
		},
	})

	for i := 0; i < 100; i++ {
		engine.AddL0("log entry number " + string(rune('a'+i%26)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Compress()
	}
}
