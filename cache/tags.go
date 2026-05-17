package cache

import (
	"fmt"
	"strings"
	"time"
)

// TaggedCache provides tagged cache functionality.
type TaggedCache struct {
	repo *Repository
	tags []string
}

// NewTaggedCache creates a new TaggedCache.
func NewTaggedCache(repo *Repository, tags []string) *TaggedCache {
	return &TaggedCache{
		repo: repo,
		tags: tags,
	}
}

// Get retrieves a value from the tagged cache.
func (t *TaggedCache) Get(key string) any {
	return t.repo.Get(t.taggedKey(key))
}

// GetString retrieves a string value from the tagged cache.
func (t *TaggedCache) GetString(key string, defaultVal ...string) string {
	return t.repo.GetString(t.taggedKey(key), defaultVal...)
}

// GetInt retrieves an int value from the tagged cache.
func (t *TaggedCache) GetInt(key string, defaultVal ...int) int {
	return t.repo.GetInt(t.taggedKey(key), defaultVal...)
}

// Put stores a value in the tagged cache with a TTL.
func (t *TaggedCache) Put(key string, value any, ttl time.Duration) error {
	// Store the value
	taggedKey := t.taggedKey(key)
	if err := t.repo.Put(taggedKey, value, ttl); err != nil {
		return err
	}

	// Track the key in each tag set
	for _, tag := range t.tags {
		setKey := t.tagSetKey(tag)
		keys := t.getTaggedKeys(setKey)
		keys[taggedKey] = true
		t.repo.Forever(setKey, keys)
	}

	return nil
}

// Forever stores a value in the tagged cache permanently.
func (t *TaggedCache) Forever(key string, value any) error {
	return t.Put(key, value, 0)
}

// Has checks if a key exists in the tagged cache.
func (t *TaggedCache) Has(key string) bool {
	return t.repo.Has(t.taggedKey(key))
}

// Forget removes a value from the tagged cache.
func (t *TaggedCache) Forget(key string) bool {
	return t.repo.Forget(t.taggedKey(key))
}

// Flush removes all values associated with the tags.
func (t *TaggedCache) Flush() error {
	// Get all keys for each tag
	keySet := make(map[string]bool)
	for _, tag := range t.tags {
		setKey := t.tagSetKey(tag)
		keys := t.getTaggedKeys(setKey)
		for k := range keys {
			keySet[k] = true
		}
		// Remove the tag set itself
		t.repo.Forget(setKey)
	}

	// Remove all tagged keys
	for key := range keySet {
		t.repo.Forget(key)
	}

	return nil
}

// Remember retrieves a value or stores the result of the callback if missing.
func (t *TaggedCache) Remember(key string, ttl time.Duration, fn func() any) any {
	taggedKey := t.taggedKey(key)
	if val, ok := t.repo.store.Get(taggedKey); ok {
		return val
	}

	val := fn()
	t.Put(key, val, ttl)
	return val
}

// RememberForever retrieves a value or stores the result of the callback permanently.
func (t *TaggedCache) RememberForever(key string, fn func() any) any {
	taggedKey := t.taggedKey(key)
	if val, ok := t.repo.store.Get(taggedKey); ok {
		return val
	}

	val := fn()
	t.Forever(key, val)
	return val
}

// taggedKey generates a key prefixed with the tag namespace.
func (t *TaggedCache) taggedKey(key string) string {
	tagNamespace := strings.Join(t.tags, "|")
	return fmt.Sprintf("tagged:%s:%s", tagNamespace, key)
}

// tagSetKey generates the key for storing the set of tagged keys.
func (t *TaggedCache) tagSetKey(tag string) string {
	return fmt.Sprintf("tagset:%s", tag)
}

// getTaggedKeys retrieves the set of keys for a tag.
func (t *TaggedCache) getTaggedKeys(setKey string) map[string]bool {
	val := t.repo.Get(setKey)
	if keys, ok := val.(map[string]bool); ok {
		return keys
	}
	return make(map[string]bool)
}
