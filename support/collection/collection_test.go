package collection

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMakeAndAll(t *testing.T) {
	items := []int{1, 2, 3}
	c := Make(items)
	if !reflect.DeepEqual(c.All(), items) {
		t.Errorf("expected %v, got %v", items, c.All())
	}
}

func TestCount(t *testing.T) {
	c := Make([]int{1, 2, 3})
	if c.Count() != 3 {
		t.Errorf("expected 3, got %d", c.Count())
	}
}

func TestIsEmpty(t *testing.T) {
	c := Make([]int{})
	if !c.IsEmpty() {
		t.Error("expected empty collection")
	}
	if c.IsNotEmpty() {
		t.Error("expected not IsNotEmpty")
	}
	c2 := Make([]int{1})
	if c2.IsEmpty() {
		t.Error("expected non-empty collection")
	}
	if !c2.IsNotEmpty() {
		t.Error("expected IsNotEmpty")
	}
}

func TestRange(t *testing.T) {
	c := Range(1, 5)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(c.All(), expected) {
		t.Errorf("expected %v, got %v", expected, c.All())
	}
}

func TestFirst(t *testing.T) {
	c := Make([]int{1, 2, 3})
	first, ok := c.First()
	if !ok || first != 1 {
		t.Errorf("expected 1, got %d", first)
	}

	first, ok = c.First(func(x int) bool { return x > 1 })
	if !ok || first != 2 {
		t.Errorf("expected 2, got %d", first)
	}

	empty := Make([]int{})
	_, ok = empty.First()
	if ok {
		t.Error("expected no item in empty collection")
	}
}

func TestLast(t *testing.T) {
	c := Make([]int{1, 2, 3})
	last, ok := c.Last()
	if !ok || last != 3 {
		t.Errorf("expected 3, got %d", last)
	}

	last, ok = c.Last(func(x int) bool { return x < 3 })
	if !ok || last != 2 {
		t.Errorf("expected 2, got %d", last)
	}
}

func TestFilter(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	filtered := c.Filter(func(x int) bool { return x%2 == 0 })
	expected := []int{2, 4}
	if !reflect.DeepEqual(filtered.All(), expected) {
		t.Errorf("expected %v, got %v", expected, filtered.All())
	}
}

func TestReject(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	rejected := c.Reject(func(x int) bool { return x%2 == 0 })
	expected := []int{1, 3, 5}
	if !reflect.DeepEqual(rejected.All(), expected) {
		t.Errorf("expected %v, got %v", expected, rejected.All())
	}
}

func TestMap(t *testing.T) {
	c := Make([]int{1, 2, 3})
	mapped := c.Map(func(x int) int { return x * 2 })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(mapped.All(), expected) {
		t.Errorf("expected %v, got %v", expected, mapped.All())
	}
}

func TestEach(t *testing.T) {
	c := Make([]int{1, 2, 3})
	sum := 0
	c.Each(func(x int, i int) {
		sum += x
	})
	if sum != 6 {
		t.Errorf("expected 6, got %d", sum)
	}
}

func TestContains(t *testing.T) {
	c := Make([]int{1, 2, 3})
	if !c.Contains(func(x int) bool { return x == 2 }) {
		t.Error("expected to contain 2")
	}
	if c.Contains(func(x int) bool { return x == 5 }) {
		t.Error("expected not to contain 5")
	}
}

func TestEvery(t *testing.T) {
	c := Make([]int{2, 4, 6})
	if !c.Every(func(x int) bool { return x%2 == 0 }) {
		t.Error("expected all items to be even")
	}
	c2 := Make([]int{2, 3, 4})
	if c2.Every(func(x int) bool { return x%2 == 0 }) {
		t.Error("expected not all items to be even")
	}
}

func TestSome(t *testing.T) {
	c := Make([]int{1, 2, 3})
	if !c.Some(func(x int) bool { return x == 2 }) {
		t.Error("expected some item to be 2")
	}
}

func TestReduce(t *testing.T) {
	c := Make([]int{1, 2, 3, 4})
	result := c.Reduce(func(carry any, item int) any {
		return carry.(int) + item
	}, 0)
	if result != 10 {
		t.Errorf("expected 10, got %d", result)
	}
}

func TestChunk(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	chunks := c.Chunk(2)
	expected := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("expected %v, got %v", expected, chunks)
	}
}

func TestSlice(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	sliced := c.Slice(1, 3)
	expected := []int{2, 3, 4}
	if !reflect.DeepEqual(sliced.All(), expected) {
		t.Errorf("expected %v, got %v", expected, sliced.All())
	}
}

func TestTake(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	taken := c.Take(3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(taken.All(), expected) {
		t.Errorf("expected %v, got %v", expected, taken.All())
	}

	// Negative take
	taken = c.Take(-2)
	expected = []int{4, 5}
	if !reflect.DeepEqual(taken.All(), expected) {
		t.Errorf("expected %v, got %v", expected, taken.All())
	}
}

func TestSkip(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	skipped := c.Skip(2)
	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(skipped.All(), expected) {
		t.Errorf("expected %v, got %v", expected, skipped.All())
	}
}

func TestPushAndPop(t *testing.T) {
	c := Make([]int{1, 2, 3})
	pushed := c.Push(4, 5)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(pushed.All(), expected) {
		t.Errorf("expected %v, got %v", expected, pushed.All())
	}

	popped, remaining := pushed.Pop()
	if popped != 5 {
		t.Errorf("expected 5, got %d", popped)
	}
	expected = []int{1, 2, 3, 4}
	if !reflect.DeepEqual(remaining.All(), expected) {
		t.Errorf("expected %v, got %v", expected, remaining.All())
	}
}

func TestPrepend(t *testing.T) {
	c := Make([]int{3, 4, 5})
	prepended := c.Prepend(1, 2)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(prepended.All(), expected) {
		t.Errorf("expected %v, got %v", expected, prepended.All())
	}
}

func TestConcat(t *testing.T) {
	c1 := Make([]int{1, 2, 3})
	c2 := Make([]int{4, 5})
	concatenated := c1.Concat(c2)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(concatenated.All(), expected) {
		t.Errorf("expected %v, got %v", expected, concatenated.All())
	}
}

func TestDiff(t *testing.T) {
	c1 := Make([]int{1, 2, 3, 4})
	c2 := Make([]int{3, 4, 5})
	diff := c1.Diff(c2, func(a, b int) bool { return a == b })
	expected := []int{1, 2}
	if !reflect.DeepEqual(diff.All(), expected) {
		t.Errorf("expected %v, got %v", expected, diff.All())
	}
}

func TestIntersect(t *testing.T) {
	c1 := Make([]int{1, 2, 3, 4})
	c2 := Make([]int{3, 4, 5})
	intersect := c1.Intersect(c2, func(a, b int) bool { return a == b })
	expected := []int{3, 4}
	if !reflect.DeepEqual(intersect.All(), expected) {
		t.Errorf("expected %v, got %v", expected, intersect.All())
	}
}

func TestPartition(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	evens, odds := c.Partition(func(x int) bool { return x%2 == 0 })
	expectedEvens := []int{2, 4}
	expectedOdds := []int{1, 3, 5}
	if !reflect.DeepEqual(evens.All(), expectedEvens) {
		t.Errorf("expected %v, got %v", expectedEvens, evens.All())
	}
	if !reflect.DeepEqual(odds.All(), expectedOdds) {
		t.Errorf("expected %v, got %v", expectedOdds, odds.All())
	}
}

func TestSort(t *testing.T) {
	c := Make([]int{3, 1, 4, 1, 5, 9})
	sorted := c.Sort(func(a, b int) bool { return a < b })
	expected := []int{1, 1, 3, 4, 5, 9}
	if !reflect.DeepEqual(sorted.All(), expected) {
		t.Errorf("expected %v, got %v", expected, sorted.All())
	}
}

func TestReverse(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	reversed := c.Reverse()
	expected := []int{5, 4, 3, 2, 1}
	if !reflect.DeepEqual(reversed.All(), expected) {
		t.Errorf("expected %v, got %v", expected, reversed.All())
	}
}

func TestShuffle(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	shuffled := c.Shuffle()
	if len(shuffled.All()) != 5 {
		t.Error("shuffle changed collection size")
	}
}

func TestUnique(t *testing.T) {
	c := Make([]int{1, 2, 2, 3, 3, 3, 4})
	unique := c.Unique()
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(unique.All(), expected) {
		t.Errorf("expected %v, got %v", expected, unique.All())
	}

	// With key function
	type Person struct {
		Name string
		Age  int
	}
	people := Make([]Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Alice", Age: 35},
	})
	uniquePeople := people.Unique(func(p Person) any { return p.Name })
	if uniquePeople.Count() != 2 {
		t.Errorf("expected 2 unique people, got %d", uniquePeople.Count())
	}
}

func TestWhen(t *testing.T) {
	c := Make([]int{1, 2, 3})
	result := c.When(true, func(col *Collection[int]) *Collection[int] {
		return col.Push(4)
	})
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result.All(), expected) {
		t.Errorf("expected %v, got %v", expected, result.All())
	}

	result2 := c.When(false, func(col *Collection[int]) *Collection[int] {
		return col.Push(4)
	})
	expected2 := []int{1, 2, 3}
	if !reflect.DeepEqual(result2.All(), expected2) {
		t.Errorf("expected %v, got %v", expected2, result2.All())
	}
}

func TestUnless(t *testing.T) {
	c := Make([]int{1, 2, 3})
	result := c.Unless(false, func(col *Collection[int]) *Collection[int] {
		return col.Push(4)
	})
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(result.All(), expected) {
		t.Errorf("expected %v, got %v", expected, result.All())
	}
}

func TestForPage(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	page1 := c.ForPage(1, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(page1.All(), expected) {
		t.Errorf("expected %v, got %v", expected, page1.All())
	}

	page2 := c.ForPage(2, 3)
	expected = []int{4, 5, 6}
	if !reflect.DeepEqual(page2.All(), expected) {
		t.Errorf("expected %v, got %v", expected, page2.All())
	}
}

func TestNth(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5, 6, 7, 8})
	nth := c.Nth(2)
	expected := []int{1, 3, 5, 7}
	if !reflect.DeepEqual(nth.All(), expected) {
		t.Errorf("expected %v, got %v", expected, nth.All())
	}

	nth = c.Nth(3, 1)
	expected = []int{2, 5, 8}
	if !reflect.DeepEqual(nth.All(), expected) {
		t.Errorf("expected %v, got %v", expected, nth.All())
	}
}

func TestRandom(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	random := c.Random(3)
	if random.Count() != 3 {
		t.Errorf("expected 3 items, got %d", random.Count())
	}
}

func TestTap(t *testing.T) {
	c := Make([]int{1, 2, 3})
	called := false
	result := c.Tap(func(col *Collection[int]) {
		called = true
	})
	if !called {
		t.Error("tap callback not called")
	}
	if !reflect.DeepEqual(result.All(), c.All()) {
		t.Error("tap should return same collection")
	}
}

func TestPipe(t *testing.T) {
	c := Make([]int{1, 2, 3})
	result := c.Pipe(func(col *Collection[int]) any {
		return col.Count() * 2
	})
	if result != 6 {
		t.Errorf("expected 6, got %v", result)
	}
}

func TestToJSON(t *testing.T) {
	c := Make([]int{1, 2, 3})
	jsonBytes, err := c.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	var result []int
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSum(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	sum := Sum(c)
	if sum != 15 {
		t.Errorf("expected 15, got %d", sum)
	}
}

func TestAvg(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	avg := Avg(c)
	if avg != 3.0 {
		t.Errorf("expected 3.0, got %f", avg)
	}
}

func TestMin(t *testing.T) {
	c := Make([]int{5, 2, 8, 1, 9})
	min := Min(c)
	if min != 1 {
		t.Errorf("expected 1, got %d", min)
	}
}

func TestMax(t *testing.T) {
	c := Make([]int{5, 2, 8, 1, 9})
	max := Max(c)
	if max != 9 {
		t.Errorf("expected 9, got %d", max)
	}
}

func TestMedian(t *testing.T) {
	c := Make([]int{1, 2, 3, 4, 5})
	median := Median(c)
	if median != 3.0 {
		t.Errorf("expected 3.0, got %f", median)
	}

	c2 := Make([]int{1, 2, 3, 4})
	median2 := Median(c2)
	if median2 != 2.5 {
		t.Errorf("expected 2.5, got %f", median2)
	}
}

func TestMapCollection(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	if m.Count() != 3 {
		t.Errorf("expected 3, got %d", m.Count())
	}

	if !m.Has("a") {
		t.Error("expected to have key 'a'")
	}

	val, ok := m.Get("b")
	if !ok || val != 2 {
		t.Errorf("expected 2, got %d", val)
	}

	m2 := m.Put("d", 4)
	if m2.Count() != 4 {
		t.Errorf("expected 4, got %d", m2.Count())
	}
}

func TestMapCollectionOnly(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	only := m.Only([]string{"a", "c"})
	if only.Count() != 2 {
		t.Errorf("expected 2, got %d", only.Count())
	}
	if !only.Has("a") || !only.Has("c") || only.Has("b") {
		t.Error("only failed")
	}
}

func TestMapCollectionExcept(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	except := m.Except([]string{"b"})
	if except.Count() != 2 {
		t.Errorf("expected 2, got %d", except.Count())
	}
	if !except.Has("a") || !except.Has("c") || except.Has("b") {
		t.Error("except failed")
	}
}

func TestMapCollectionFilter(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	filtered := m.Filter(func(k string, v int) bool { return v > 1 })
	if filtered.Count() != 2 {
		t.Errorf("expected 2, got %d", filtered.Count())
	}
}

func TestMapCollectionMap(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	mapped := m.Map(func(k string, v int) int { return v * 2 })
	val, _ := mapped.Get("a")
	if val != 2 {
		t.Errorf("expected 2, got %d", val)
	}
}

func TestMapCollectionMerge(t *testing.T) {
	m1 := MakeMap(map[string]int{"a": 1, "b": 2})
	m2 := MakeMap(map[string]int{"c": 3, "d": 4})
	merged := m1.Merge(m2)
	if merged.Count() != 4 {
		t.Errorf("expected 4, got %d", merged.Count())
	}
}

func TestMapCollectionKeys(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	keys := m.Keys()
	if keys.Count() != 3 {
		t.Errorf("expected 3 keys, got %d", keys.Count())
	}
}

func TestMapCollectionValues(t *testing.T) {
	m := MakeMap(map[string]int{"a": 1, "b": 2, "c": 3})
	values := m.Values()
	if values.Count() != 3 {
		t.Errorf("expected 3 values, got %d", values.Count())
	}
}

func TestMapCollectionToMap(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}
	m := MakeMap(original)
	converted := m.ToMap()
	if !reflect.DeepEqual(converted, original) {
		t.Errorf("expected %v, got %v", original, converted)
	}
}
