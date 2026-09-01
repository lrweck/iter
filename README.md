# iter

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
  when — and for as long as — you ask for elements.
- **Errors are just data.** `MapErr` / `FlatMapErr` emit `(result, error)`
  pairs. Log them, drop them, or collect them. The stream never stops.
- **Zero dependencies.** `iter`, `slices`, `maps`, `cmp`, `errors`. That's it.
  That's the whole party.

## Install

```
go get github.com/lrweck/iter
```

Needs Go 1.27.

## Pipeline methods

| Method | What it does |
| --- | --- |
| `Map(f)` | Apply `f` to every element |
| `Filter(f)` | Keep elements where `f` is true |
| `SkipErr(check)` | Keep elements where `check` returns nil, skip the rest |
| `UniqByFunc(key)` | Keep the first element of each equal key (any `T`) |
| `FlatMap(f)` | Flatten the sequences produced by `f` |
| `Take(n)` / `Drop(n)` | Keep / drop the first `n` elements |
| `TakeWhile(p)` / `DropWhile(p)` | Keep / drop elements while `p` is true |
| `Tap(f)` | Peek at every element, pass it through (Elixir's `tap/2`) |
| `Enumerate()` | Attach an index → `Seq2[int, T]` |
| `Zip(o)` | Pair up corresponding elements of two sequences |

When the element type needs a constraint a method can't apply, you get a
package function — same vibe, shorter stack:

```go
iter.Uniq(iter.Of(1, 2, 2, 1))                        // []int{1, 2}
iter.Of(users).UniqByFunc(func(u user) int { return u.id })
```

The tail of the pipeline:

```go
vals := iter.Of(3, 1, 4, 1).Tap(func(v int) { log.Printf("%d", v) }).Collect() // []int
first, ok := iter.Of(9, 8).First()                          // 9, true
last, ok := iter.Of(9, 8).Last()                            // 8, true
total := iter.Sum(iter.Of(1, 2, 3))                         // 6
best, ok := iter.Max(iter.Of(3, 1, 4))                      // 4, true
pop, ok := iter.Of(users).MaxByFunc(func(u user) int { return u.age })
pairs := iter.ToMap(iter.Of("a", "b").Enumerate())          // map[int]string
buckets := iter.Of(vals).GroupByFunc(func(v int) string { return ... })
pick := iter.Of(vals).KeyByFunc(func(v int) string { return ... })
```

Generators:

```go
iter.Range(2, 5).Collect()                                                 // []int{2, 3, 4}
iter.Repeat(3, "x").Collect()                                              // []string{"x", "x", "x"}
iter.RepeatBy(3, func(i int) int { return i * i }).Collect()               // []int{0, 1, 4}
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
`Result[V]`, and you own what failures mean:

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

Errors behave the Go way: `errors.Is` to compare, `%w` to wrap.

## Why it's shaped like that

Go won't let a generic method constrain its own receiver's type. Three things
fall off the method set because of it:

- `Sum` / `Max` / `Min` / `Uniq` need numeric, ordered, or comparable
  element types, so they're package functions. Their keyed counterparts —
  `SumByFunc` / `MaxByFunc` / `MinByFunc` / `UniqByFunc` / `KeyByFunc` /
  `GroupByFunc` — shove the constraint onto the key and work as methods for
  any `T`.
- `Chunk` returns `Seq[[]T]`, and the compiler chases `T → []T → ...` through
  the method set until it gives up.
- Error recovery pins the pair slot to `error`, which generic `Seq2` methods
  can't say — so there's a dedicated `Result[V]` type whose methods (`Take`,
  `Filter`, `IgnoreErr`, `Errors`, `CollectErr`, `Tap`) speak error fluently.

## License

[MIT](LICENSE)