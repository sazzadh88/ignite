package collection_test

import (
	"fmt"

	"github.com/sazzadh88/ignite/support/collection"
)

func ExampleCollection_basic() {
	// Create a collection
	c := collection.Make([]int{1, 2, 3, 4, 5})

	// Filter even numbers
	evens := c.Filter(func(x int) bool { return x%2 == 0 })

	// Map to double values
	doubled := evens.Map(func(x int) int { return x * 2 })

	fmt.Println(doubled.All())
	// Output: [4 8]
}

func ExampleCollection_chain() {
	// Chaining operations
	result := collection.Make([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		Filter(func(x int) bool { return x%2 == 0 }).
		Map(func(x int) int { return x * 2 }).
		Take(3)

	fmt.Println(result.All())
	// Output: [4 8 12]
}

func ExampleSum() {
	c := collection.Make([]int{1, 2, 3, 4, 5})
	sum := collection.Sum(c)
	fmt.Println(sum)
	// Output: 15
}

func ExampleAvg() {
	c := collection.Make([]int{2, 4, 6, 8})
	avg := collection.Avg(c)
	fmt.Println(avg)
	// Output: 5
}

func ExampleMapCollection() {
	// Create a map collection
	m := collection.MakeMap(map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 7,
	})

	// Filter items with value > 4
	filtered := m.Filter(func(k string, v int) bool { return v > 4 })

	fmt.Println(filtered.Count())
	// Output: 2
}

func ExampleCollection_partition() {
	c := collection.Make([]int{1, 2, 3, 4, 5, 6})

	// Partition into evens and odds
	evens, odds := c.Partition(func(x int) bool { return x%2 == 0 })

	fmt.Println("Evens:", evens.All())
	fmt.Println("Odds:", odds.All())
	// Output:
	// Evens: [2 4 6]
	// Odds: [1 3 5]
}

func ExampleCollection_reduce() {
	c := collection.Make([]int{1, 2, 3, 4, 5})

	// Calculate sum using reduce
	sum := c.Reduce(func(carry any, item int) any {
		return carry.(int) + item
	}, 0)

	fmt.Println(sum)
	// Output: 15
}

func ExampleCollection_chunk() {
	c := collection.Make([]int{1, 2, 3, 4, 5, 6, 7})

	chunks := c.Chunk(3)

	for i, chunk := range chunks {
		fmt.Printf("Chunk %d: %v\n", i+1, chunk)
	}
	// Output:
	// Chunk 1: [1 2 3]
	// Chunk 2: [4 5 6]
	// Chunk 3: [7]
}

func ExampleCollection_forPage() {
	c := collection.Make([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	page1 := c.ForPage(1, 3)
	page2 := c.ForPage(2, 3)

	fmt.Println("Page 1:", page1.All())
	fmt.Println("Page 2:", page2.All())
	// Output:
	// Page 1: [1 2 3]
	// Page 2: [4 5 6]
}

func ExampleRange() {
	// Create a range of numbers
	c := collection.Range(1, 6)
	fmt.Println(c.All())
	// Output: [1 2 3 4 5]
}
