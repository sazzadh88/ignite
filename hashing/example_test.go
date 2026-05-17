package hashing_test

import (
	"fmt"
	"log"

	"github.com/sazzad/ignite/hashing"
)

func Example_basicUsage() {
	// Hash a password using the global facade
	password := "secret123"
	hash, err := hashing.Hash.Make(password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hash format: $ignite$iterations$salt$hash")

	// Verify correct password
	if hashing.Hash.Check(password, hash) {
		fmt.Println("Password is correct")
	}

	// Verify incorrect password
	if !hashing.Hash.Check("wrongpassword", hash) {
		fmt.Println("Password is incorrect")
	}

	// Check if hash needs rehashing (e.g., old iteration count)
	if hashing.Hash.NeedsRehash(hash) {
		fmt.Println("Hash needs rehashing")
	} else {
		fmt.Println("Hash is current")
	}

	// Output:
	// Hash format: $ignite$iterations$salt$hash
	// Password is correct
	// Password is incorrect
	// Hash is current
}

func Example_customDriver() {
	// Create a custom hasher with higher iterations
	customHasher := hashing.NewSHA256Hasher(20000)

	password := "supersecure"
	hash, _ := customHasher.Make(password)

	// Verify
	if customHasher.Check(password, hash) {
		fmt.Println("Custom hasher works")
	}

	// Output:
	// Custom hasher works
}

func Example_manager() {
	// Create a new manager
	manager := hashing.NewManager()

	// Register a custom driver
	strongHasher := hashing.NewSHA256Hasher(50000)
	manager.Register("strong", strongHasher)

	// Use the custom driver
	password := "mypassword"
	hash, _ := manager.Driver("strong").Make(password)

	// Verify
	if manager.Driver("strong").Check(password, hash) {
		fmt.Println("Strong hasher verification passed")
	}

	// Set as default
	manager.SetDefault("strong")

	// Now default methods use the strong hasher
	newHash, _ := manager.Make("anotherpassword")
	_ = newHash

	// Output:
	// Strong hasher verification passed
}
