package encryption_test

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/sazzad/ignite/encryption"
)

func Example_basicUsage() {
	// Generate a secure key
	key, err := encryption.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Create encrypter
	encrypter, err := encryption.NewEncrypter(key)
	if err != nil {
		log.Fatal(err)
	}

	// Encrypt a string
	plaintext := "Hello, World!"
	encrypted, err := encrypter.EncryptString(plaintext)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Encrypted (base64)")

	// Decrypt
	decrypted, err := encrypter.DecryptString(encrypted)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decrypted: %s\n", decrypted)

	// Output:
	// Encrypted (base64)
	// Decrypted: Hello, World!
}

func Example_encryptStructs() {
	key, _ := encryption.GenerateKey()
	encrypter, _ := encryption.NewEncrypter(key)

	// Encrypt complex data
	data := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
		"admin": true,
	}

	encrypted, err := encrypter.Encrypt(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Data encrypted")

	// Decrypt
	decrypted, err := encrypter.Decrypt(encrypted)
	if err != nil {
		log.Fatal(err)
	}

	// Type assertion
	result := decrypted.(map[string]any)
	fmt.Printf("Name: %s\n", result["name"])
	fmt.Printf("Email: %s\n", result["email"])

	// Output:
	// Data encrypted
	// Name: John Doe
	// Email: john@example.com
}

func Example_generateKey() {
	// Generate a cryptographically secure 32-byte key
	key, err := encryption.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Key is suitable for AES-256
	fmt.Printf("Key length: %d bytes\n", len(key))

	// Encode as base64 for storage (e.g., in APP_KEY environment variable)
	encoded := base64.StdEncoding.EncodeToString(key)
	fmt.Println("Key encoded for storage (base64)")

	// Decode when needed
	decoded, _ := base64.StdEncoding.DecodeString(encoded)
	_ = decoded

	// Output:
	// Key length: 32 bytes
	// Key encoded for storage (base64)
}

func Example_differentCiphertexts() {
	key, _ := encryption.GenerateKey()
	encrypter, _ := encryption.NewEncrypter(key)

	plaintext := "same message"

	// Encrypt twice
	encrypted1, _ := encrypter.EncryptString(plaintext)
	encrypted2, _ := encrypter.EncryptString(plaintext)

	// Ciphertexts are different due to random nonces
	if encrypted1 != encrypted2 {
		fmt.Println("Ciphertexts are different")
	}

	// But both decrypt to the same plaintext
	decrypted1, _ := encrypter.DecryptString(encrypted1)
	decrypted2, _ := encrypter.DecryptString(encrypted2)

	if decrypted1 == decrypted2 && decrypted1 == plaintext {
		fmt.Println("Both decrypt to original message")
	}

	// Output:
	// Ciphertexts are different
	// Both decrypt to original message
}
