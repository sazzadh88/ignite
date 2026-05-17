# Collection Package

A Laravel-inspired, generic collection package for Go. Provides a fluent, chainable API for working with slices and maps.

## Features

- **Zero external dependencies** - uses only Go standard library
- **Generic types** - works with any type using Go generics
- **Immutable operations** - all methods return new collections
- **Fluent API** - chainable method calls
- **Type-safe** - compile-time type checking
- **Comprehensive** - 40+ collection methods

## Installation

```bash
go get github.com/sazzad/goframe/support/collection
```

## Quick Start

```go
import "github.com/sazzad/goframe/support/collection"

// Create a collection
numbers := collection.Make([]int{1, 2, 3, 4, 5})

// Chain operations
result := numbers.
    Filter(func(x int) bool { return x%2 == 0 }).
    Map(func(x int) int { return x * 2 }).
    All()
// result: [4, 8]
```

## Collection Methods

### Creating Collections

- `Make[T](items []T) *Collection[T]` - Create from slice
- `Range(start, end int) *Collection[int]` - Create range of integers

### Inspection

- `All() []T` - Get all items as slice
- `Count() int` - Count items
- `IsEmpty() bool` - Check if empty
- `IsNotEmpty() bool` - Check if not empty
- `First(fn ...func(T) bool) (T, bool)` - Get first item
- `Last(fn ...func(T) bool) (T, bool)` - Get last item

### Filtering

- `Filter(fn func(T) bool) *Collection[T]` - Keep matching items
- `Reject(fn func(T) bool) *Collection[T]` - Remove matching items
- `Unique(fn ...func(T) any) *Collection[T]` - Remove duplicates
- `Contains(fn func(T) bool) bool` - Check if any item matches
- `Every(fn func(T) bool) bool` - Check if all items match
- `Some(fn func(T) bool) bool` - Check if at least one item matches

### Transformation

- `Map(fn func(T) T) *Collection[T]` - Transform each item
- `Each(fn func(T, int))` - Iterate over items
- `Reduce(fn func(carry any, item T) any, initial any) any` - Reduce to single value
- `Flatten() *Collection[T]` - Flatten one level

### Ordering

- `Sort(fn func(a, b T) bool) *Collection[T]` - Sort items
- `Reverse() *Collection[T]` - Reverse order
- `Shuffle() *Collection[T]` - Random order

### Slicing

- `Slice(offset, length int) *Collection[T]` - Get subset
- `Take(n int) *Collection[T]` - Take first n items
- `Skip(n int) *Collection[T]` - Skip first n items
- `Chunk(size int) [][]T` - Split into chunks
- `Nth(step int, offset ...int) *Collection[T]` - Every nth item
- `ForPage(page, perPage int) *Collection[T]` - Pagination

### Combining

- `Push(items ...T) *Collection[T]` - Append items
- `Pop() (T, *Collection[T])` - Remove and return last
- `Prepend(items ...T) *Collection[T]` - Add to beginning
- `Concat(other *Collection[T]) *Collection[T]` - Merge collections
- `Diff(other *Collection[T], eq func(T, T) bool) *Collection[T]` - Difference
- `Intersect(other *Collection[T], eq func(T, T) bool) *Collection[T]` - Intersection

### Advanced

- `Partition(fn func(T) bool) (*Collection[T], *Collection[T])` - Split by condition
- `Tap(fn func(*Collection[T])) *Collection[T]` - Execute callback
- `Pipe(fn func(*Collection[T]) any) any` - Pass to callback
- `When(condition bool, fn func(*Collection[T]) *Collection[T]) *Collection[T]` - Conditional execution
- `Unless(condition bool, fn func(*Collection[T]) *Collection[T]) *Collection[T]` - Inverse conditional
- `Random(n ...int) *Collection[T]` - Random items
- `ToJSON() ([]byte, error)` - Serialize to JSON

## Numeric Operations

For numeric types, additional operations are available:

```go
numbers := collection.Make([]int{1, 2, 3, 4, 5})

sum := collection.Sum(numbers)        // 15
avg := collection.Avg(numbers)        // 3.0
min := collection.Min(numbers)        // 1
max := collection.Max(numbers)        // 5
median := collection.Median(numbers)  // 3.0
```

Supported numeric types: `int`, `int64`, `float64`, `float32`

## Map Collections

Work with maps using `MapCollection`:

```go
m := collection.MakeMap(map[string]int{
    "apple":  5,
    "banana": 3,
    "orange": 7,
})

// Filter by value
filtered := m.Filter(func(k string, v int) bool { return v > 4 })

// Transform values
doubled := m.Map(func(k string, v int) int { return v * 2 })

// Extract keys/values
keys := m.Keys()      // Collection of keys
values := m.Values()  // Collection of values

// Check existence
exists := m.Has("apple")

// Subset operations
subset := m.Only([]string{"apple", "orange"})
without := m.Except([]string{"banana"})

// Merge
m2 := collection.MakeMap(map[string]int{"grape": 4})
merged := m.Merge(m2)
```

## Examples

### Filter and Transform

```go
users := collection.Make([]User{
    {Name: "Alice", Age: 30},
    {Name: "Bob", Age: 25},
    {Name: "Charlie", Age: 35},
})

adults := users.
    Filter(func(u User) bool { return u.Age >= 30 }).
    Map(func(u User) User {
        u.Name = "Dr. " + u.Name
        return u
    })
```

### Pagination

```go
items := collection.Range(1, 101) // 1-100

page1 := items.ForPage(1, 10) // [1, 2, ..., 10]
page2 := items.ForPage(2, 10) // [11, 12, ..., 20]
```

### Partitioning

```go
numbers := collection.Make([]int{1, 2, 3, 4, 5, 6})

evens, odds := numbers.Partition(func(x int) bool { 
    return x%2 == 0 
})
// evens: [2, 4, 6]
// odds: [1, 3, 5]
```

### Conditional Operations

```go
numbers := collection.Make([]int{1, 2, 3})

result := numbers.
    When(len(numbers.All()) > 2, func(c *Collection[int]) *Collection[int] {
        return c.Push(4)
    }).
    Unless(numbers.IsEmpty(), func(c *Collection[int]) *Collection[int] {
        return c.Take(2)
    })
```

### Reduce

```go
words := collection.Make([]string{"hello", "world", "!"})

sentence := words.Reduce(func(carry any, word string) any {
    return carry.(string) + " " + word
}, "")
// "hello world !"
```

### Set Operations

```go
set1 := collection.Make([]int{1, 2, 3, 4})
set2 := collection.Make([]int{3, 4, 5, 6})

eq := func(a, b int) bool { return a == b }

diff := set1.Diff(set2, eq)         // [1, 2]
intersect := set1.Intersect(set2, eq) // [3, 4]
```

## Design Principles

1. **Immutability** - All operations return new collections, original remains unchanged
2. **Zero Allocations** - Where possible, slice backing arrays are reused
3. **Type Safety** - Generic types provide compile-time safety
4. **No Magic** - Explicit operations, no hidden behavior
5. **Laravel-Inspired** - Familiar API for Laravel developers

## Performance Considerations

- Collections copy slices on creation to ensure immutability
- Methods like `Filter`, `Map` create new slices
- For high-performance scenarios, direct slice operations may be preferred
- Use `Each` instead of `Map` when side effects are needed

## License

Part of the GoFrame framework.
