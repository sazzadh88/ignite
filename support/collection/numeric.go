package collection

import "sort"

// Numeric is a type constraint for numeric types.
type Numeric interface {
	~int | ~int64 | ~float64 | ~float32
}

// Sum returns the sum of all numeric items.
func Sum[T Numeric](c *Collection[T]) T {
	var sum T
	for _, item := range c.items {
		sum += item
	}
	return sum
}

// Avg returns the average of all numeric items.
func Avg[T Numeric](c *Collection[T]) float64 {
	if len(c.items) == 0 {
		return 0
	}
	sum := Sum(c)
	return float64(sum) / float64(len(c.items))
}

// Min returns the minimum numeric item.
func Min[T Numeric](c *Collection[T]) T {
	if len(c.items) == 0 {
		var zero T
		return zero
	}
	min := c.items[0]
	for _, item := range c.items[1:] {
		if item < min {
			min = item
		}
	}
	return min
}

// Max returns the maximum numeric item.
func Max[T Numeric](c *Collection[T]) T {
	if len(c.items) == 0 {
		var zero T
		return zero
	}
	max := c.items[0]
	for _, item := range c.items[1:] {
		if item > max {
			max = item
		}
	}
	return max
}

// Median returns the median of all numeric items.
func Median[T Numeric](c *Collection[T]) float64 {
	if len(c.items) == 0 {
		return 0
	}
	sorted := make([]T, len(c.items))
	copy(sorted, c.items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (float64(sorted[mid-1]) + float64(sorted[mid])) / 2
	}
	return float64(sorted[mid])
}
