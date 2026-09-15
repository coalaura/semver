# semver

A small version parser supporting SemVer syntax and partial versions.

## Supported formats

```text
1
1.2
1.2.3
v1.2.3
1.2.3-alpha.1
1.2.3+build.2026
1.2.3-rc.1+build.2026
```

Leading and trailing whitespace is trimmed. Numeric components are limited to `uint16` and cannot contain leading zeros.

Partial versions, the `v` prefix and Unicode suffix identifiers are extensions to SemVer 2.0.0.

## Usage

```go
version, err := ParseSemVer("v1.2.3-rc.1", true)
if err != nil {
	log.Fatal(err)
}

fmt.Println(version.Major)      // 1
fmt.Println(version.Minor)      // 2
fmt.Println(version.Patch)      // 3
fmt.Println(version.Suffix)     // -rc.1
fmt.Println(version.IsValid())  // true
fmt.Println(version.String())   // v1.2.3-rc.1
```

`Valid`, `IsValid` and `IsInvalid` distinguish a successfully parsed zero version from an invalid result. `Invalid` is the canonical invalid value and can be used instead of `SemVer{}`. `ParseSemVer` and `NewEmptySemVer` return valid values on success. Set `Valid: true` when constructing a valid `SemVer` with a struct literal.

Set `allowSuffix` to `false` to reject prerelease and build metadata:

```go
_, err := ParseSemVer("1.2.3-alpha", false)
if errors.Is(err, ErrDisallowedSuffix) {
	// Suffixes are not allowed.
}
```

## Comparison

```go
stable, _ := ParseSemVer("1.0.0", true)
candidate, _ := ParseSemVer("1.0.0-rc.1", true)

stable.HigherThan(candidate) // true
stable.Compare(candidate)    // 1
```

`Compare` follows SemVer precedence:

- Missing components compare as zero.
- Prerelease versions have lower precedence than releases.
- Build metadata and the `v` prefix do not affect precedence.

`Equal` compares validity, parsed components, component presence and the complete suffix. It ignores the optional `v` prefix.

## Performance

Successful parsing is designed to perform no heap allocations. `String` allocates the returned string.

Run the tests and benchmarks with:

```sh
go test ./...
go test -race ./...
go test -bench=. -benchmem ./...
```
