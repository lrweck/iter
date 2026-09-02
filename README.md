# iter

[![PkgGoDev](https://pkg.go.dev/badge/github.com/lrweck/iter)](https://pkg.go.dev/github.com/lrweck/iter)

Lazy, composable, type-safe pipelines for Go 1.27 iterators — the chill
sibling of `samber/lo`, built on range-over-func.

```go
iter.Of(1, 2, 3, 4, 5, 6).
    Filter(func(v int) bool { return v%2 == 0 }).
    Map(func(v int) int { return v * v }).
    Tap(func(v int) { log.Println("even square:", v) }).
    Collect()
// [4 16 36]
```

Nothing runs until the terminal method — and *you* are the driver. `break`,
`Take`, done. Upstream stops working the instant you stop asking. No pull
iterators, no `Next()`, no ceremony.

## Why you'll like it

- **Actually reads well.** Go 1.27's generic methods turn `of(x).filter(f).
  map(g).collect()` into a fluent chain instead of paren soup.
- **Lazy or get out.** Every op is a `range-over-func`. Work happens only
  when — and for as long as — you ask for elements. No eager surprises, no
  double work.
- **Errors are just data.** `MapErr` / `FlatMapErr` emit `(result, error)`
  pairs. Log them, drop them, or collect them. The stream never stops.
- **Zero dependencies.** `iter`, `slices`, `maps`, `cmp`, `errors`. That's it.
  That's the whole party. No vendoring drama, no supply-chain anxiety.

## Classic Go vs iter

Same logic. Two flavors. You pick. `iter` trades a little perf for code that
reads the way you think — no cranking through index math or nesting.

**Map + Filter + Take.** First 10 squares of the evens up to 10,000:

```go
// Classic Go
nums := make([]int, 0, 10)
for _, v := range input {
    if v%2 == 0 {
        nums = append(nums, v*v)
        if len(nums) == 10 {
            break
        }
    }
}

// iter
nums := iter.Of(input...).
    Filter(func(v int) bool { return v%2 == 0 }).
    Map(func(v int) int { return v * v }).
    Take(10).
    Collect()
```

**Sum.** Accumulate with a plain `for` vs `iter.Sum`:

```go
// Classic Go
var total int
for _, v := range nums {
    total += v
}

// iter
total := iter.Sum(iter.Of(nums...))
```

**Map with errors.** Turn strings into ints, skipping the ones that fail:

```go
// Classic Go
parsed := make([]int, 0, len(raw))
for _, s := range raw {
    if n, err := strconv.Atoi(s); err == nil {
        parsed = append(parsed, n)
    }
}

// iter
parsed := iter.Of(raw...).
    MapErr(strconv.Atoi).
    IgnoreErr().
    Collect()
```

**Dedup.** Keep the first occurrence of each key:

```go
// Classic Go
seen := make(map[int]struct{})
out := make([]int, 0, len(users))
for _, u := range users {
    if _, ok := seen[u.id]; ok {
        continue
    }
    seen[u.id] = struct{}{}
    out = append(out, u.id)
}

// iter
out := iter.Of(users...).UniqBy(func(u user) int { return u.id }).Collect()
```

Reading is more direct, but it comes at a cost (measured below): every
terminal carries a fixed allocation overhead from `range-over-func` and the
closures. Know your hot path.

## Install

```
go get github.com/lrweck/iter
```

Needs Go 1.27. Nothing else. Seriously.

## Pipeline methods

| Method | What it does |
| --- | --- |
| `Map(f)` | Apply `f` to every element |
| `Filter(f)` | Keep elements where `f` is true |
| `SkipErr(check)` | Keep elements where `check` returns nil, skip the rest |
| `UniqBy(key)` | Keep the first element of each distinct key (any `T`) |
| `FlatMap(f)` | Flatten the sequences produced by `f` |
| `Take(n)` / `Drop(n)` | Keep / drop the first `n` elements |
| `TakeWhile(p)` / `DropWhile(p)` | Keep / drop elements while `p` is true |
| `Tap(f)` | Peek at every element, pass it through (Elixir's `tap/2`) |
| `Enumerate()` | Attach an index → `Seq2[int, T]` |
| `Zip(o)` | Pair up corresponding elements of two sequences |

`Seq2` (the output of `Enumerate` / `Zip`) shares the pipeline: `Filter`,
`Take`, `Drop`, `TakeWhile`, `DropWhile`, `Tap`, plus `Keys`, `Values`,
`Collect`, and `ToMap`.

When the element type needs a constraint a method can't apply, you get a
package function — same vibe, shorter stack:

```go
iter.Uniq(iter.Of(1, 2, 2, 1))                        // []int{1, 2}
iter.Of(users).UniqBy(func(u user) int { return u.id })
```

The tail of the pipeline:

```go
vals := iter.Of(3, 1, 4, 1).Tap(func(v int) { log.Printf("%d", v) }).Collect() // []int
first, ok := iter.Of(9, 8).First()                          // 9, true
last, ok := iter.Of(9, 8).Last()                            // 8, true
total := iter.Sum(iter.Of(1, 2, 3))                         // 6
best, ok := iter.Max(iter.Of(3, 1, 4))                      // 4, true
pop, ok := iter.Of(users).MaxBy(func(u user) int { return u.age })
pairs := iter.ToMap(iter.Of("a", "b").Enumerate())          // map[int]string
buckets := iter.Of(vals).GroupBy(func(v int) string { return ... })
pick := iter.Of(vals).KeyBy(func(v int) string { return ... })
```

Generators:

```go
iter.Range(2, 5, 1).Collect()                                                // []int{2, 3, 4}
iter.Range(5, 1, -1).Collect()                                               // []int{5, 4, 3, 2}
iter.Repeat(3, "x").Collect()                                                // []string{"x", "x", "x"}
iter.RepeatBy(3, func(i int) int { return i * i }).Collect()                 // []int{0, 1, 4}
iter.From(slices.Values([]string{"a", "b"}))                               // wrap any stdlib iterator
iter.Concat(iter.Of(1), iter.Of(2, 3)).Collect()                         // []int{1, 2, 3}
iter.Contains(iter.Of(3, 1, 4), 4)                                       // true
iter.ToMap(iter.FromMap(m))                                              // map round-trip
iter.FromKeys(m).Collect(), iter.FromValues(m).Collect()                 // the map's keys / values
for v := range iter.Of(1, 2, 3).Seq() { ... }                            // or range directly
```

Predicate terminals bail at the first decisive element:

```go
iter.Of(1, 3, 4).Some(isEven)      // true, 4 never even gets looked at
iter.Of(2, 4).Every(isEven)        // true
iter.Of(1, 3).None(isEven)         // true
iter.Of(1, 3, 4).Find(isEven)      // 4, true
iter.Of(codes).CountBy(isEven)     // int
```

## Bad input? Whatever, keep going

Fallible ops never kill the stream. Every element pairs with an outcome as a
`Result[V]`, and you own what failures mean — no `if err != nil` after every
single line:

```go
parsed := iter.Of("1", "junk", "2").MapErr(strconv.Atoi)

// log failures as they flow through
parsed.Tap(func(v int, err error) {
    if err != nil {
        log.Printf("bad line: %v", err)
    }
}).CollectErr()

// filter the failures out, keep the wins
good := parsed.IgnoreErr().Collect() // []int{1, 2}

// ... or grab values and every failure at once
values, err := parsed.CollectErr() // []int{1, 2}, errors.Join of all failures
if err != nil {
    log.Printf("%v", err)
}
```

Errors behave the Go way: `errors.Is` to compare, `%w` to wrap. Because of
course they do — we're not animals.

## Why it's shaped like that

Go won't let a generic method constrain its own receiver's type. Three things
fall off the method set because of it:

- `Sum` / `Max` / `Min` / `Uniq` need numeric, ordered, or comparable
  element types, so they're package functions. Their keyed counterparts —
  `SumBy` / `MaxBy` / `MinBy` / `UniqBy` / `KeyBy` / `GroupBy` — shove the
  constraint onto the key and work as methods for any `T`.
- `Chunk` returns `Seq[[]T]`, and the compiler chases `T → []T → ...` through
  the method set until it gives up.
- Error recovery pins the pair slot to `error`, which generic `Seq2` methods
  can't say — so there's a dedicated `Result[V]` type whose methods (`Take`,
  `Filter`, `IgnoreErr`, `Errors`, `CollectErr`, `Tap`) speak error fluently.

Zero values are safe: a `Seq`, `Seq2`, or `Result` declared but never
initialized behaves as an empty sequence.

## Benchmarks

Measured on an Intel Core i7-13700H, `go test -bench` with `-benchmem` and
`-count=3`. Compare a declarative pipeline against the equivalent traditional
loop.

### Pipeline: Map + Filter + Take (10,000 items, take the first 10 evens and square them)

| Approach | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Traditional loop | ~134 | 0 | 0 |
| `iter` | ~3.9k | 2304 | 18 |

**Lazy short-circuiting works** — `Take(10)` stops early. The cost comes from
the `range-over-func` closures, not from scanning the whole source.

### Map (1,000 items, `v*2`)

| Approach | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Traditional loop | ~3.8k | 8192 | 1 |
| `iter.Map(...).Collect()` | ~19k | 25336 | 17 |

### Filter (1,000 items, evens)

| Approach | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Traditional loop | ~4.7k | 8192 | 1 |
| `iter.Filter(...).Collect()` | ~12k | 8312 | 15 |

### Sum (1,000 items)

| Approach | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Traditional loop | ~0.6k | 0 | 0 |
| `iter.Sum(...)` | ~1.8k | 48 | 2 |

### Map with errors (1,000 strings → int, all valid)

| Approach | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Traditional loop | ~4.0k | 8192 | 1 |
| `iter.MapErr(...).IgnoreErr().Collect()` | ~19k | 25400 | 19 |

### Terminals (fixed overhead per call)

| Terminal | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `First()` | ~78 | 80 | 3 |
| `Count()` | ~1.7k | 48 | 2 |
| `Some(isEven)` (early stop) | ~2.5k | 72 | 3 |

### The cost in one sentence

Every method call pays a fixed overhead (~48B/2 allocs per terminal, plus the
closures of each pipeline layer). For small lists or one-off operations this
is whatever, who cares. For hot loops over millions of items, the traditional
loop ends up 5–30× faster. `iter` is about **readability and composition** —
reach for a plain loop on the hot path when you need raw speed, and let
`iter` carry the lines you'd rather not hand-crank.

## License

[MIT](LICENSE)