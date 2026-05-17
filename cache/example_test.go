package cache_test

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/sazzad/ignite/cache"
)

func ExampleCache_basic() {
	// Use the default memory store
	cache.Cache.Store("").Put("user:1", "John Doe", 5*time.Minute)

	name := cache.Cache.Store("").GetString("user:1")
	fmt.Println(name)
	// Output: John Doe
}

func ExampleRepository_Remember() {
	repo := cache.Cache.Store("")

	// Expensive computation
	result := repo.Remember("expensive", 5*time.Minute, func() any {
		// This only runs once
		return "computed result"
	})

	fmt.Println(result)
	// Output: computed result
}

func ExampleRepository_Increment() {
	repo := cache.Cache.Store("")

	repo.Put("views", 100, 1*time.Hour)
	newValue, _ := repo.Increment("views")
	fmt.Println(newValue)

	newValue, _ = repo.Increment("views", 5)
	fmt.Println(newValue)
	// Output:
	// 101
	// 106
}

func ExampleRepository_Lock() {
	repo := cache.Cache.Store("")
	lock := repo.Lock("critical-section", 10*time.Second)

	lock.Get(func() {
		// Only one process can execute this at a time
		fmt.Println("Executing critical section")
	})
	// Output: Executing critical section
}

func ExampleTaggedCache() {
	repo := cache.Cache.Store("")
	tagged := repo.Tags([]string{"users", "posts"})

	tagged.Put("user:1", "John", 1*time.Hour)
	tagged.Put("user:2", "Jane", 1*time.Hour)

	// Flush all cached items with these tags
	tagged.Flush()

	fmt.Println(tagged.Has("user:1"))
	// Output: false
}

func ExampleFileStore() {
	dir := filepath.Join("/tmp", "my-cache")
	fileStore, _ := cache.NewFileStore(dir)

	cache.Cache.Register("file", fileStore)
	cache.Cache.SetDefault("file")

	repo := cache.Cache.Store("")
	repo.Put("persistent-key", "persistent-value", 24*time.Hour)

	fmt.Println(repo.GetString("persistent-key"))
	// Output: persistent-value
}
