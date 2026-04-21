package graph

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Graph struct {
	db       *sql.DB
	dbPath   string
	mu       sync.RWMutex
	basePath string
}

type Entity struct {
	ID        int64
	Name      string
	Type      string
	Tier      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Link struct {
	ID         int64
	SourceID   int64
	TargetID   int64
	LinkType   string
	CreatedAt  time.Time
	Confidence float64
}

type Page struct {
	ID        int64
	Title     string
	Content   string
	URL       string
	Type      string
	Tags      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	EntityTypePerson    = "person"
	EntityTypeCompany   = "company"
	EntityTypeConcept   = "concept"
	EntityTypeEvent     = "event"
	EntityTypeProject   = "project"

	LinkTypeWorksAt     = "works_at"
	LinkTypeFounded     = "founded"
	LinkTypeInvestedIn  = "invested_in"
	LinkTypeAdvises     = "advises"
	LinkTypeAttended    = "attended"
	LinkTypeMentioned   = "mentioned"
	LinkTypeRelatedTo   = "related_to"
)

func New(basePath string) (*Graph, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "graph")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("graph mkdir: %w", err)
	}

	dbPath := filepath.Join(basePath, "graph.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("graph open: %w", err)
	}

	g := &Graph{
		db:       db,
		dbPath:   dbPath,
		basePath: basePath,
	}

	if err := g.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return g, nil
}

func (g *Graph) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS entities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'person',
		tier INTEGER DEFAULT 3,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, type)
	);

	CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		link_type TEXT NOT NULL,
		confidence REAL DEFAULT 1.0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_id, target_id, link_type),
		FOREIGN KEY(source_id) REFERENCES entities(id) ON DELETE CASCADE,
		FOREIGN KEY(target_id) REFERENCES entities(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS pages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		url TEXT,
		type TEXT DEFAULT 'note',
		tags TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS page_entities (
		page_id INTEGER NOT NULL,
		entity_id INTEGER NOT NULL,
		PRIMARY KEY(page_id, entity_id),
		FOREIGN KEY(page_id) REFERENCES pages(id) ON DELETE CASCADE,
		FOREIGN KEY(entity_id) REFERENCES entities(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name);
	CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);
	CREATE INDEX IF NOT EXISTS idx_links_source ON links(source_id);
	CREATE INDEX IF NOT EXISTS idx_links_target ON links(target_id);
	CREATE INDEX IF NOT EXISTS idx_links_type ON links(link_type);
	CREATE INDEX IF NOT EXISTS idx_pages_type ON pages(type);
	`

	_, err := g.db.Exec(schema)
	return err
}

func (g *Graph) AddEntity(name, entityType string) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("empty entity name")
	}

	if entityType == "" {
		entityType = g.inferEntityType(name)
	}

	result, err := g.db.Exec(`
		INSERT INTO entities (name, type, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(name, type) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
	`, name, entityType)

	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	return id, nil
}

func (g *Graph) inferEntityType(name string) string {
	nameLower := strings.ToLower(name)

	companySuffixes := []string{"inc", "llc", "ltd", "corp", "co", "ai", "labs", "tech", "io"}
	for _, suffix := range companySuffixes {
		if strings.HasSuffix(nameLower, suffix) {
			return EntityTypeCompany
		}
	}

	personIndicators := []string{"@", "twitter.com", "github.com"}
	for _, ind := range personIndicators {
		if strings.Contains(nameLower, ind) {
			return EntityTypePerson
		}
	}

	return EntityTypeConcept
}

func (g *Graph) GetOrCreateEntity(name, entityType string) (*Entity, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	name = strings.TrimSpace(name)
	if entityType == "" {
		entityType = g.inferEntityType(name)
	}

	var entity Entity
	err := g.db.QueryRow(`
		SELECT id, name, type, tier, created_at, updated_at
		FROM entities WHERE name = ? AND type = ?
	`, name, entityType).Scan(&entity.ID, &entity.Name, &entity.Type, &entity.Tier, &entity.CreatedAt, &entity.UpdatedAt)

	if err == sql.ErrNoRows {
		result, err := g.db.Exec(`
			INSERT INTO entities (name, type)
			VALUES (?, ?)
		`, name, entityType)
		if err != nil {
			return nil, err
		}

		id, _ := result.LastInsertId()
		return &Entity{
			ID:        id,
			Name:      name,
			Type:      entityType,
			Tier:      3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	return &entity, err
}

func (g *Graph) AddLink(sourceID, targetID int64, linkType string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, err := g.db.Exec(`
		INSERT INTO links (source_id, target_id, link_type)
		VALUES (?, ?, ?)
		ON CONFLICT(source_id, target_id, link_type) DO NOTHING
	`, sourceID, targetID, linkType)

	return err
}

func (g *Graph) CreatePage(title, content, pageType, tags string) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	result, err := g.db.Exec(`
		INSERT INTO pages (title, content, type, tags)
		VALUES (?, ?, ?, ?)
	`, title, content, pageType, tags)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (g *Graph) UpdatePage(id int64, title, content, tags string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, err := g.db.Exec(`
		UPDATE pages SET title = ?, content = ?, tags = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, content, tags, id)

	return err
}

func (g *Graph) LinkPageToEntity(pageID, entityID int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, err := g.db.Exec(`
		INSERT OR IGNORE INTO page_entities (page_id, entity_id)
		VALUES (?, ?)
	`, pageID, entityID)

	return err
}

func (g *Graph) ExtractAndLink(content string, pageID int64) error {
	entities := g.extractEntities(content)

	for _, entityName := range entities {
		entity, err := g.GetOrCreateEntity(entityName, "")
		if err != nil {
			continue
		}

		g.LinkPageToEntity(pageID, entity.ID)

		linkType := g.inferLinkType(content, entityName)
		if linkType != "" {
			g.AddLink(pageID, entity.ID, linkType)
		}
	}

	return nil
}

func (g *Graph) extractEntities(content string) []string {
	var entities []string

	patterns := []struct {
		regex *regexp.Regexp
		group int
	}{
		{regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`), 1},
		{regexp.MustCompile(`@(\w+)`), 1},
		{regexp.MustCompile(`\*\*([^*]+)\*\*`), 1},
		{regexp.MustCompile(`([A-Z][a-z]+ [A-Z][a-z]+)`), 1},
		{regexp.MustCompile(`([A-Z][a-z]+ (?:Inc|LLC|Ltd|Corp|Co|AI|Labs))`), 1},
	}

	for _, p := range patterns {
		matches := p.regex.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) > p.group {
				name := strings.TrimSpace(m[p.group])
				if len(name) > 2 && len(name) < 100 {
					entities = append(entities, name)
				}
			}
		}
	}

	seen := make(map[string]bool)
	var unique []string
	for _, e := range entities {
		if !seen[e] {
			seen[e] = true
			unique = append(unique, e)
		}
	}

	return unique
}

func (g *Graph) inferLinkType(content, entityName string) string {
	contentLower := strings.ToLower(content)
	nameLower := strings.ToLower(entityName)

	patterns := map[string][]string{
		LinkTypeFounded:     {"founded", "co-founded", "created", "started"},
		LinkTypeInvestedIn:  {"invested in", "investor", "backed", "funded"},
		LinkTypeAdvises:     {"advises", "advisor", "mentor", "board"},
		LinkTypeWorksAt:     {"works at", "working at", "employee", "engineer", "developer", "ceo", "cto", "founder", "president"},
		LinkTypeAttended:    {"attended", "meeting with", "talked to", "chat with"},
	}

	for linkType, keywords := range patterns {
		for _, kw := range keywords {
			if strings.Contains(contentLower, kw) {
				return linkType
			}
		}
	}

	return LinkTypeMentioned
}

func (g *Graph) Query(startEntity string, linkType string, depth int) ([]Entity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if depth > 3 {
		depth = 3
	}

	var startID int64
	err := g.db.QueryRow(`SELECT id FROM entities WHERE name = ?`, startEntity).Scan(&startID)
	if err != nil {
		return nil, err
	}

	query := `
		WITH RECURSIVE graph_traverse AS (
			SELECT target_id as entity_id, 1 as depth
			FROM links WHERE source_id = ? AND (? = '' OR link_type = ?)
			UNION ALL
			SELECT l.target_id, gt.depth + 1
			FROM links l
			JOIN graph_traverse gt ON l.source_id = gt.entity_id
			WHERE gt.depth < ?
		)
		SELECT DISTINCT e.id, e.name, e.type, e.tier, e.created_at, e.updated_at
		FROM entities e
		JOIN graph_traverse gt ON e.id = gt.entity_id
		ORDER BY gt.depth
		LIMIT 50
	`

	rows, err := g.db.Query(query, startID, linkType, linkType, depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.Tier, &e.CreatedAt, &e.UpdatedAt); err != nil {
			continue
		}
		results = append(results, e)
	}

	return results, nil
}

func (g *Graph) GetEntity(name string) (*Entity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var entity Entity
	err := g.db.QueryRow(`
		SELECT id, name, type, tier, created_at, updated_at
		FROM entities WHERE name = ?
	`, name).Scan(&entity.ID, &entity.Name, &entity.Type, &entity.Tier, &entity.CreatedAt, &entity.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &entity, err
}

func (g *Graph) SearchByEntity(entityName string, limit int) ([]Page, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT p.id, p.title, p.content, p.url, p.type, p.tags, p.created_at, p.updated_at
		FROM pages p
		JOIN page_entities pe ON p.id = pe.page_id
		JOIN entities e ON pe.entity_id = e.id
		WHERE e.name = ?
		ORDER BY p.updated_at DESC
		LIMIT ?
	`

	rows, err := g.db.Query(query, entityName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Page
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.URL, &p.Type, &p.Tags, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		results = append(results, p)
	}

	return results, nil
}

func (g *Graph) GetStats() map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := make(map[string]int)

	g.db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&stats["entities"])
	g.db.QueryRow("SELECT COUNT(*) FROM links").Scan(&stats["links"])
	g.db.QueryRow("SELECT COUNT(*) FROM pages").Scan(&stats["pages"])

	var typeCounts map[string]int
	rows, _ := g.db.Query("SELECT type, COUNT(*) FROM entities GROUP BY type")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var t string
			var c int
			rows.Scan(&t, &c)
			stats["entities_"+t] = c
		}
	}

	return stats
}

func (g *Graph) BacklinkBoost(query string, limit int) ([]Page, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	query = `
		SELECT p.id, p.title, p.content, p.url, p.type, p.tags, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM page_entities pe2 JOIN entities e2 ON pe2.entity_id = e2.id
			 WHERE pe2.page_id = p.id) as link_count
		FROM pages p
		WHERE p.content LIKE ? OR p.title LIKE ?
		ORDER BY link_count DESC, p.updated_at DESC
		LIMIT ?
	`

	searchTerm := "%" + query + "%"
	rows, err := g.db.Query(query, searchTerm, searchTerm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Page
	for rows.Next() {
		var p Page
		var linkCount int
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.URL, &p.Type, &p.Tags, &p.CreatedAt, &p.UpdatedAt, &linkCount); err != nil {
			continue
		}
		results = append(results, p)
	}

	return results, nil
}

func (g *Graph) Close() error {
	return g.db.Close()
}