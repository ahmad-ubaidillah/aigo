package orchestration

import (
	"fmt"
	"sync"
	"time"
)

// ConflictType represents the type of resource conflict.
type ConflictType int

const (
	ConflictTypeFileAccess ConflictType = iota
	ConflictTypeResourceLock
	ConflictTypeWriteOperation
	ConflictTypeExclusiveAccess
)

// ResolutionStrategy defines how conflicts are resolved.
type ResolutionStrategy int

const (
	ResolutionFIFO     ResolutionStrategy = iota // First-in-first-out
	ResolutionPriority                            // Higher priority wins
	ResolutionYoungest                            // Most recent request wins
	ResolutionOldest                              // Oldest request wins
)

// Conflict represents a detected resource conflict.
type Conflict struct {
	ID          string
	Type        ConflictType
	Resource    string
	Requesters  []string // Agent IDs competing for the resource
	Resolved    bool
	Resolution  string // ID of the winning requester
	DetectedAt  time.Time
	ResolvedAt  time.Time
	Priority    int
}

// ResourceLock represents a lock on a shared resource.
type ResourceLock struct {
	Resource    string
	Holder      string // Agent ID holding the lock
	AcquiredAt  time.Time
	ExpiresAt   time.Time
	LockType    ConflictType
	WaitQueue   []string // Agent IDs waiting for this lock
}

// ConflictResolver detects and resolves conflicts for shared resources.
type ConflictResolver struct {
	mu            sync.RWMutex
	locks         map[string]*ResourceLock // resource -> lock
	conflicts     map[string]*Conflict     // conflict ID -> conflict
	pending       map[string][]string      // resource -> waiting agent IDs
	strategy      ResolutionStrategy
	maxRetries    int
	lockTimeout   time.Duration
	conflictCount int64
}

// NewConflictResolver creates a new ConflictResolver.
func NewConflictResolver(strategy ResolutionStrategy) *ConflictResolver {
	return &ConflictResolver{
		locks:       make(map[string]*ResourceLock),
		conflicts:   make(map[string]*Conflict),
		pending:     make(map[string][]string),
		strategy:    strategy,
		maxRetries:  3,
		lockTimeout: 5 * time.Minute,
	}
}

// DetectConflicts checks for conflicts on the given resources.
func (cr *ConflictResolver) DetectConflicts(resources []string) []Conflict {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var conflicts []Conflict
	for _, resource := range resources {
		if lock, exists := cr.locks[resource]; exists {
			conflicts = append(conflicts, Conflict{
				ID:         generateConflictID(),
				Type:       lock.LockType,
				Resource:   resource,
				Requesters: append([]string{lock.Holder}, lock.WaitQueue...),
				Resolved:   false,
				DetectedAt: time.Now(),
			})
		}
	}
	return conflicts
}

// TryLock attempts to acquire a lock on a resource.
// Returns true if successful, false if the resource is already locked.
func (cr *ConflictResolver) TryLock(resource, agentID string, lockType ConflictType) bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Check if resource is already locked
	if lock, exists := cr.locks[resource]; exists {
		// Check if lock has expired
		if !lock.ExpiresAt.IsZero() && time.Now().After(lock.ExpiresAt) {
			// Lock expired, release it
			delete(cr.locks, resource)
		} else {
			// Resource is locked, add to wait queue
			lock.WaitQueue = append(lock.WaitQueue, agentID)
			return false
		}
	}

	// Acquire lock
	cr.locks[resource] = &ResourceLock{
		Resource:   resource,
		Holder:     agentID,
		AcquiredAt: time.Now(),
		ExpiresAt:  time.Now().Add(cr.lockTimeout),
		LockType:   lockType,
		WaitQueue:  make([]string, 0),
	}
	return true
}

// Lock acquires a lock, blocking until available or timeout.
func (cr *ConflictResolver) Lock(resource, agentID string, lockType ConflictType, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if cr.TryLock(resource, agentID, lockType) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("lock timeout for resource %s", resource)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Unlock releases a lock on a resource.
func (cr *ConflictResolver) Unlock(resource, agentID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	lock, exists := cr.locks[resource]
	if !exists {
		return fmt.Errorf("resource %s not locked", resource)
	}
	if lock.Holder != agentID {
		return fmt.Errorf("agent %s does not hold lock on %s", agentID, resource)
	}

	// If there are waiters, transfer lock to next in queue
	if len(lock.WaitQueue) > 0 {
		nextHolder := cr.resolveNextHolder(lock)
		lock.Holder = nextHolder
		lock.AcquiredAt = time.Now()
		lock.ExpiresAt = time.Now().Add(cr.lockTimeout)
		lock.WaitQueue = lock.WaitQueue[1:]
	} else {
		delete(cr.locks, resource)
	}

	return nil
}

// Resolve resolves a detected conflict using the configured strategy.
func (cr *ConflictResolver) Resolve(conflictID string) (*Conflict, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	conflict, exists := cr.conflicts[conflictID]
	if !exists {
		return nil, fmt.Errorf("conflict %s not found", conflictID)
	}

	if conflict.Resolved {
		return conflict, nil
	}

	// Resolve based on strategy
	switch cr.strategy {
	case ResolutionPriority:
		// Find requester with highest priority (already sorted in requesters)
		if len(conflict.Requesters) > 0 {
			conflict.Resolution = conflict.Requesters[0]
		}
	case ResolutionYoungest:
		// Most recent request wins
		if len(conflict.Requesters) > 0 {
			conflict.Resolution = conflict.Requesters[len(conflict.Requesters)-1]
		}
	case ResolutionOldest:
		// Oldest request wins
		if len(conflict.Requesters) > 0 {
			conflict.Resolution = conflict.Requesters[0]
		}
	default: // FIFO
		if len(conflict.Requesters) > 0 {
			conflict.Resolution = conflict.Requesters[0]
		}
	}

	conflict.Resolved = true
	conflict.ResolvedAt = time.Now()
	cr.conflictCount++

	return conflict, nil
}

// resolveNextHolder determines the next agent to receive a lock based on strategy.
func (cr *ConflictResolver) resolveNextHolder(lock *ResourceLock) string {
	if len(lock.WaitQueue) == 0 {
		return ""
	}

	switch cr.strategy {
	case ResolutionPriority, ResolutionOldest, ResolutionFIFO:
		return lock.WaitQueue[0]
	case ResolutionYoungest:
		return lock.WaitQueue[len(lock.WaitQueue)-1]
	default:
		return lock.WaitQueue[0]
	}
}

// IsLocked checks if a resource is currently locked.
func (cr *ConflictResolver) IsLocked(resource string) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	if lock, exists := cr.locks[resource]; exists {
		// Check if lock has expired
		if !lock.ExpiresAt.IsZero() && time.Now().After(lock.ExpiresAt) {
			return false
		}
		return true
	}
	return false
}

// GetLockHolder returns the agent ID holding the lock on a resource.
func (cr *ConflictResolver) GetLockHolder(resource string) (string, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	lock, exists := cr.locks[resource]
	if !exists {
		return "", fmt.Errorf("resource %s not locked", resource)
	}
	if !lock.ExpiresAt.IsZero() && time.Now().After(lock.ExpiresAt) {
		return "", fmt.Errorf("lock on %s has expired", resource)
	}
	return lock.Holder, nil
}

// GetLockedResources returns all resources currently locked by an agent.
func (cr *ConflictResolver) GetLockedResources(agentID string) []string {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var resources []string
	for resource, lock := range cr.locks {
		if lock.Holder == agentID {
			resources = append(resources, resource)
		}
	}
	return resources
}

// ReleaseAll releases all locks held by an agent.
func (cr *ConflictResolver) ReleaseAll(agentID string) int {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	count := 0
	for resource, lock := range cr.locks {
		if lock.Holder == agentID {
			// Transfer to next waiter if any
			if len(lock.WaitQueue) > 0 {
				nextHolder := cr.resolveNextHolder(lock)
				lock.Holder = nextHolder
				lock.AcquiredAt = time.Now()
				lock.ExpiresAt = time.Now().Add(cr.lockTimeout)
				lock.WaitQueue = lock.WaitQueue[1:]
			} else {
				delete(cr.locks, resource)
			}
			count++
		}
	}
	return count
}

// SetLockTimeout sets the default timeout for locks.
func (cr *ConflictResolver) SetLockTimeout(d time.Duration) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.lockTimeout = d
}

// ConflictCount returns the total number of conflicts resolved.
func (cr *ConflictResolver) ConflictCount() int64 {
	return cr.conflictCount
}

// Clear removes all locks and conflicts.
func (cr *ConflictResolver) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.locks = make(map[string]*ResourceLock)
	cr.conflicts = make(map[string]*Conflict)
	cr.pending = make(map[string][]string)
}

// generateConflictID creates a unique conflict identifier.
func generateConflictID() string {
	return fmt.Sprintf("conflict-%d", time.Now().UnixNano())
}
