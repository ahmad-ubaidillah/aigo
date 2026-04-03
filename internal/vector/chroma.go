package vector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ChromaClient communicates with a ChromaDB server via HTTP.
type ChromaClient struct {
	baseURL     string
	httpClient  *http.Client
	mu          sync.RWMutex
	collections map[string]*ChromaCollection
}

// ChromaCollection represents a ChromaDB collection.
type ChromaCollection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewChromaClient creates a new ChromaDB HTTP client.
func NewChromaClient(baseURL string) *ChromaClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return &ChromaClient{
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		collections: make(map[string]*ChromaCollection),
	}
}

// Connect tests the connection to ChromaDB.
func (c *ChromaClient) Connect() error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/heartbeat")
	if err != nil {
		return fmt.Errorf("connect to ChromaDB: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ChromaDB heartbeat failed: %s", resp.Status)
	}
	return nil
}

// CreateCollection creates a new collection in ChromaDB.
func (c *ChromaClient) CreateCollection(name string) error {
	body := map[string]any{"name": name}
	data, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/collections", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create collection failed: %s", resp.Status)
	}
	var col ChromaCollection
	if err := json.NewDecoder(resp.Body).Decode(&col); err != nil {
		return fmt.Errorf("decode collection: %w", err)
	}
	c.mu.Lock()
	c.collections[name] = &col
	c.mu.Unlock()
	return nil
}

// DeleteCollection deletes a collection from ChromaDB.
func (c *ChromaClient) DeleteCollection(name string) error {
	req, _ := http.NewRequest("DELETE", c.baseURL+"/api/v1/collections/"+name, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete collection failed: %s", resp.Status)
	}
	c.mu.Lock()
	delete(c.collections, name)
	c.mu.Unlock()
	return nil
}

// ListCollections returns all collection names.
func (c *ChromaClient) ListCollections() ([]string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/collections")
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer resp.Body.Close()
	var cols []ChromaCollection
	if err := json.NewDecoder(resp.Body).Decode(&cols); err != nil {
		return nil, fmt.Errorf("decode collections: %w", err)
	}
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names, nil
}

// Upsert adds or updates documents in a collection.
func (c *ChromaClient) Upsert(collection string, ids []string, embeddings [][]float64, metadatas []map[string]any, documents []string) error {
	body := map[string]any{
		"ids":        ids,
		"embeddings": embeddings,
		"metadatas":  metadatas,
		"documents":  documents,
	}
	data, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/collections/"+collection+"/upsert", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("upsert: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upsert failed: %s", resp.Status)
	}
	return nil
}

// Query searches a collection by embedding.
func (c *ChromaClient) Query(collection string, embedding []float64, nResults int) (*QueryResult, error) {
	body := map[string]any{
		"query_embeddings": [][]float64{embedding},
		"n_results":        nResults,
		"include":          []string{"documents", "metadatas", "distances"},
	}
	data, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/collections/"+collection+"/query", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer resp.Body.Close()
	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode query result: %w", err)
	}
	return &result, nil
}

// QueryResult holds the response from a ChromaDB query.
type QueryResult struct {
	IDs       [][]string         `json:"ids"`
	Distances [][]float64        `json:"distances"`
	Metadatas [][]map[string]any `json:"metadatas"`
	Documents [][]string         `json:"documents"`
}

// Get retrieves documents by ID.
func (c *ChromaClient) Get(collection string, ids []string) (*QueryResult, error) {
	body := map[string]any{"ids": ids, "include": []string{"documents", "metadatas"}}
	data, _ := json.Marshal(body)
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/collections/"+collection+"/get", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()
	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode get result: %w", err)
	}
	return &result, nil
}

func (c *ChromaClient) Delete(collection string, ids []string) error {
	body := map[string]any{"ids": ids}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("DELETE", c.baseURL+"/api/v1/collections/"+collection+"/delete", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s: %s", resp.Status, string(body))
	}
	return nil
}
