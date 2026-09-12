package semver

import (
	"errors"
	"strings"
	"testing"
)

const benchmarkParseSemVerInput = "v1.2.3-rc.1+build.2026"

type semVerTestCase struct {
	name          string
	input         string
	allowSuffix   bool
	expected      SemVer
	expectedError error
}

type semVerStringTestCase struct {
	name     string
	version  SemVer
	expected string
}

type semVerCompareTestCase struct {
	name     string
	a        string
	b        string
	expected int
}

type semVerEqualTestCase struct {
	name     string
	version  SemVer
	expected bool
}

func TestParseSemVer(t *testing.T) {
	t.Parallel()

	tests := []semVerTestCase{
		{
			name:        "standard full version",
			input:       "1.2.3",
			allowSuffix: false,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
		},
		{
			name:        "zero values",
			input:       "0.0.0",
			allowSuffix: false,
			expected:    SemVer{Major: 0, Minor: 0, Patch: 0, HasMinor: true, HasPatch: true},
		},
		{
			name:        "outer spaces",
			input:       "  1.2.3  ",
			allowSuffix: false,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
		},
		{
			name:        "unicode outer spaces",
			input:       "\u00a0v1.2.3\u00a0",
			allowSuffix: false,
			expected:    SemVer{Prefix: 'v', Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
		},
		{
			name:        "major only",
			input:       "1",
			allowSuffix: false,
			expected:    SemVer{Major: 1},
		},
		{
			name:        "major and minor",
			input:       "1.2",
			allowSuffix: false,
			expected:    SemVer{Major: 1, Minor: 2, HasMinor: true},
		},
		{
			name:        "prefix on full version",
			input:       "v1.2.3",
			allowSuffix: false,
			expected:    SemVer{Prefix: 'v', Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
		},
		{
			name:        "prefix on partial version",
			input:       "v2",
			allowSuffix: false,
			expected:    SemVer{Prefix: 'v', Major: 2},
		},
		{
			name:        "maximum major",
			input:       "65535",
			allowSuffix: false,
			expected:    SemVer{Major: 65535},
		},
		{
			name:        "maximum components",
			input:       "65535.65535.65535",
			allowSuffix: false,
			expected:    SemVer{Major: 65535, Minor: 65535, Patch: 65535, HasMinor: true, HasPatch: true},
		},
		{
			name:        "prerelease",
			input:       "1.2.3-alpha",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-alpha"},
		},
		{
			name:        "prerelease identifiers",
			input:       "1.2.3-alpha.1",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-alpha.1"},
		},
		{
			name:        "build metadata",
			input:       "1.2.3+build.2026",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "+build.2026"},
		},
		{
			name:        "prerelease and build metadata",
			input:       "1.2.3-rc.1+sha.1234",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-rc.1+sha.1234"},
		},
		{
			name:        "suffix on major",
			input:       "1-beta",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Suffix: "-beta"},
		},
		{
			name:        "suffix on minor",
			input:       "1.2-beta",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, HasMinor: true, Suffix: "-beta"},
		},
		{
			name:        "numeric zero prerelease",
			input:       "1.2.3-0",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-0"},
		},
		{
			name:        "alphanumeric prerelease beginning with zero",
			input:       "1.2.3-01a",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-01a"},
		},
		{
			name:        "build identifier with leading zeros",
			input:       "1.2.3+001",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "+001"},
		},
		{
			name:        "hyphen identifier",
			input:       "1.2.3--",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "--"},
		},
		{
			name:        "unicode suffix",
			input:       "1.2.3-beta.à",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-beta.à"},
		},
		{
			name:        "unicode combining mark",
			input:       "1.2.3-a\u0300",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-a\u0300"},
		},

		{
			name:          "empty input",
			input:         "",
			allowSuffix:   true,
			expectedError: ErrEmptyVersion,
		},
		{
			name:          "whitespace only",
			input:         " \t\r\n ",
			allowSuffix:   true,
			expectedError: ErrEmptyVersion,
		},
		{
			name:          "only prefix",
			input:         "v",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "double prefix",
			input:         "vv1.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "uppercase prefix",
			input:         "V1.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "space after prefix",
			input:         "v 1.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "negative major",
			input:         "-1.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "positive sign",
			input:         "+1.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "leading dot",
			input:         ".1.2",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "trailing dot after major",
			input:         "1.",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "trailing dot after minor",
			input:         "1.2.",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "dot after patch",
			input:         "1.2.3.",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "empty minor",
			input:         "1..2",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "empty patch",
			input:         "1.2..3",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "four components",
			input:         "1.2.3.4",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "non-numeric major",
			input:         "one.2.3",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "non-numeric minor",
			input:         "1.two.3",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "non-numeric patch",
			input:         "1.2.three",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "leading zero major",
			input:         "01.2.3",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "leading zero minor",
			input:         "1.02.3",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "leading zero patch",
			input:         "1.2.03",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "major overflow",
			input:         "65536.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "minor overflow",
			input:         "1.65536.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "patch overflow",
			input:         "1.0.65536",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "very large component",
			input:         "999999999999999999999999999999999999.0.0",
			allowSuffix:   true,
			expectedError: ErrInvalidVersion,
		},
		{
			name:          "prerelease disallowed",
			input:         "1.2.3-alpha",
			allowSuffix:   false,
			expectedError: ErrDisallowedSuffix,
		},
		{
			name:          "build metadata disallowed",
			input:         "1.2.3+build",
			allowSuffix:   false,
			expectedError: ErrDisallowedSuffix,
		},
		{
			name:          "empty prerelease",
			input:         "1.2.3-",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty build metadata",
			input:         "1.2.3+",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty first prerelease identifier",
			input:         "1.2.3-.alpha",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty middle prerelease identifier",
			input:         "1.2.3-alpha..1",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty final prerelease identifier",
			input:         "1.2.3-alpha.",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty first build identifier",
			input:         "1.2.3-alpha+.build",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty middle build identifier",
			input:         "1.2.3+build..1",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "empty final build identifier",
			input:         "1.2.3+build.",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "multiple build separators",
			input:         "1.2.3-alpha+build+other",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "numeric prerelease leading zero",
			input:         "1.2.3-01",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "numeric prerelease component leading zero",
			input:         "1.2.3-alpha.01",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "ASCII space in suffix",
			input:         "1.2.3-alpha beta",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "unicode space in suffix",
			input:         "1.2.3-alpha\u00a0beta",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "underscore in suffix",
			input:         "1.2.3-alpha_beta",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "punctuation in suffix",
			input:         "1.2.3-alpha!",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "slash in suffix",
			input:         "1.2.3-alpha/beta",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "NUL in suffix",
			input:         "1.2.3-alpha\x00beta",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
		{
			name:          "invalid UTF-8 suffix",
			input:         "1.2.3-\xff",
			allowSuffix:   true,
			expectedError: ErrInvalidSuffix,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			version, err := ParseSemVer(test.input, test.allowSuffix)

			if test.expectedError != nil {
				if !errors.Is(err, test.expectedError) {
					t.Fatalf("expected error %v, got %v", test.expectedError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if version != test.expected {
				t.Fatalf("expected %+v, got %+v", test.expected, version)
			}
		})
	}
}

func TestSemVerString(t *testing.T) {
	t.Parallel()

	tests := []semVerStringTestCase{
		{
			name:     "zero value",
			version:  SemVer{},
			expected: "0",
		},
		{
			name:     "empty version",
			version:  NewEmptySemVer(),
			expected: "0.0.0",
		},
		{
			name:     "major",
			version:  SemVer{Major: 12},
			expected: "12",
		},
		{
			name:     "major and minor",
			version:  SemVer{Major: 12, Minor: 34, HasMinor: true},
			expected: "12.34",
		},
		{
			name:     "full version",
			version:  SemVer{Major: 12, Minor: 34, Patch: 56, HasMinor: true, HasPatch: true},
			expected: "12.34.56",
		},
		{
			name:     "prefix",
			version:  SemVer{Prefix: 'v', Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
			expected: "v1.2.3",
		},
		{
			name: "suffix",
			version: SemVer{
				Major:    1,
				Minor:    2,
				Patch:    3,
				HasMinor: true,
				HasPatch: true,
				Suffix:   "-rc.1+build.2",
			},
			expected: "1.2.3-rc.1+build.2",
		},
		{
			name: "hidden components",
			version: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			expected: "1",
		},
		{
			name: "maximum components",
			version: SemVer{
				Major:    65535,
				Minor:    65535,
				Patch:    65535,
				HasMinor: true,
				HasPatch: true,
			},
			expected: "65535.65535.65535",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := test.version.String()
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSemVerStringRoundTrip(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"0",
		"1.2",
		"v1.2.3",
		"1-beta",
		"1.2-rc.1",
		"1.2.3-alpha.1+build.2",
		"65535.65535.65535",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			expected, err := ParseSemVer(input, true)
			if err != nil {
				t.Fatalf("parse input: %v", err)
			}

			actual, err := ParseSemVer(expected.String(), true)
			if err != nil {
				t.Fatalf("parse rendered version: %v", err)
			}

			if actual != expected {
				t.Fatalf("expected %+v, got %+v", expected, actual)
			}
		})
	}
}

func TestSemVerCompare(t *testing.T) {
	t.Parallel()

	ordered := []string{
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
		"1.0.0",
	}

	versions := make([]SemVer, len(ordered))

	for i, input := range ordered {
		version, err := ParseSemVer(input, true)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}

		versions[i] = version
	}

	for i, a := range versions {
		for j, b := range versions {
			expected := 0

			switch {
			case i < j:
				expected = -1
			case i > j:
				expected = 1
			}

			comparison := a.Compare(b)
			if comparison != expected {
				t.Fatalf("%q.Compare(%q): expected %d, got %d", ordered[i], ordered[j], expected, comparison)
			}

			higher := a.HigherThan(b)
			if higher != (expected > 0) {
				t.Fatalf("%q.HigherThan(%q): expected %t, got %t", ordered[i], ordered[j], expected > 0, higher)
			}
		}
	}
}

func TestSemVerCompareSpecialCases(t *testing.T) {
	t.Parallel()

	tests := []semVerCompareTestCase{
		{
			name:     "missing components equal zero components",
			a:        "1",
			b:        "1.0.0",
			expected: 0,
		},
		{
			name:     "missing patch equals zero patch",
			a:        "1.2",
			b:        "1.2.0",
			expected: 0,
		},
		{
			name:     "minor decides precedence",
			a:        "1.3",
			b:        "1.2.65535",
			expected: 1,
		},
		{
			name:     "build metadata ignored",
			a:        "1.2.3+build.1",
			b:        "1.2.3+build.2",
			expected: 0,
		},
		{
			name:     "prefix ignored",
			a:        "v1.2.3",
			b:        "1.2.3",
			expected: 0,
		},
		{
			name:     "release higher than prerelease",
			a:        "1",
			b:        "1.0.0-alpha",
			expected: 1,
		},
		{
			name:     "numeric identifier lower than text",
			a:        "1.0.0-1",
			b:        "1.0.0-alpha",
			expected: -1,
		},
		{
			name:     "numeric identifiers compared numerically",
			a:        "1.0.0-99",
			b:        "1.0.0-100",
			expected: -1,
		},
		{
			name:     "long numeric identifiers do not overflow",
			a:        "1.0.0-999999999999999999999999999999",
			b:        "1.0.0-1000000000000000000000000000000",
			expected: -1,
		},
		{
			name:     "additional identifier is higher",
			a:        "1.0.0-alpha",
			b:        "1.0.0-alpha.1",
			expected: -1,
		},
		{
			name:     "Unicode identifiers compared lexically",
			a:        "1.0.0-à",
			b:        "1.0.0-é",
			expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			a, err := ParseSemVer(test.a, true)
			if err != nil {
				t.Fatalf("parse first version: %v", err)
			}

			b, err := ParseSemVer(test.b, true)
			if err != nil {
				t.Fatalf("parse second version: %v", err)
			}

			actual := a.Compare(b)
			if actual != test.expected {
				t.Fatalf("expected %d, got %d", test.expected, actual)
			}
		})
	}
}

func TestSemVerEqual(t *testing.T) {
	t.Parallel()

	base := SemVer{
		Major:    1,
		Minor:    2,
		Patch:    3,
		HasMinor: true,
		HasPatch: true,
		Suffix:   "-rc.1+build.2",
	}

	tests := []semVerEqualTestCase{
		{
			name:     "same version",
			version:  base,
			expected: true,
		},
		{
			name: "prefix ignored",
			version: SemVer{
				Prefix:   'v',
				Major:    1,
				Minor:    2,
				Patch:    3,
				HasMinor: true,
				HasPatch: true,
				Suffix:   "-rc.1+build.2",
			},
			expected: true,
		},
		{
			name: "different major",
			version: SemVer{
				Major:    2,
				Minor:    2,
				Patch:    3,
				HasMinor: true,
				HasPatch: true,
				Suffix:   "-rc.1+build.2",
			},
			expected: false,
		},
		{
			name: "missing minor",
			version: SemVer{
				Major:  1,
				Suffix: "-rc.1+build.2",
			},
			expected: false,
		},
		{
			name: "missing patch",
			version: SemVer{
				Major:    1,
				Minor:    2,
				HasMinor: true,
				Suffix:   "-rc.1+build.2",
			},
			expected: false,
		},
		{
			name: "different suffix",
			version: SemVer{
				Major:    1,
				Minor:    2,
				Patch:    3,
				HasMinor: true,
				HasPatch: true,
				Suffix:   "-rc.2+build.2",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := base.Equal(test.version)
			if actual != test.expected {
				t.Fatalf("expected %t, got %t", test.expected, actual)
			}
		})
	}
}

func TestParseSemVerAllocations(t *testing.T) {
	input := "v1.2.3-rc.1+build.2026"

	var (
		version SemVer
		err     error
	)

	allocations := testing.AllocsPerRun(1000, func() {
		version, err = ParseSemVer(input, true)
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if version.String() != input {
		t.Fatalf("expected %q, got %q", input, version.String())
	}

	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %v", allocations)
	}
}

func FuzzParseSemVer(f *testing.F) {
	seeds := []string{
		"",
		"1",
		"1.2",
		"v1.2.3",
		"1.2.3-alpha.1+build.2",
		"1.2.3-beta.à",
		"65535.65535.65535",
		"65536.0.0",
		"1.2.3-alpha..1",
		string([]byte{'1', '.', '2', '.', '3', '-', 0xff}),
	}

	for _, seed := range seeds {
		f.Add(seed, true)
		f.Add(seed, false)
	}

	f.Fuzz(func(t *testing.T, input string, allowSuffix bool) {
		version, err := ParseSemVer(input, allowSuffix)
		if err != nil {
			if !errors.Is(err, ErrEmptyVersion) && !errors.Is(err, ErrInvalidVersion) && !errors.Is(err, ErrInvalidSuffix) && !errors.Is(err, ErrDisallowedSuffix) {
				t.Fatalf("unexpected error: %v", err)
			}

			return
		}

		rendered := version.String()

		roundTrip, err := ParseSemVer(rendered, true)
		if err != nil {
			t.Fatalf("parse rendered version %q: %v", rendered, err)
		}

		if roundTrip != version {
			t.Fatalf("round trip changed version: before=%+v after=%+v", version, roundTrip)
		}
	})
}

func BenchmarkParseSemVer(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_, err := ParseSemVer(benchmarkParseSemVerInput, true)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSemVerString(b *testing.B) {
	version := SemVer{
		Prefix:   'v',
		Major:    1,
		Minor:    2,
		Patch:    3,
		HasMinor: true,
		HasPatch: true,
		Suffix:   "-rc.1+build.2026",
	}

	b.ReportAllocs()

	for b.Loop() {
		_ = version.String()
	}
}

func BenchmarkParseSemVerLongSuffix(b *testing.B) {
	input := "1.2.3-" + strings.Repeat("identifier.", 20) + "final"

	b.ReportAllocs()

	for b.Loop() {
		_, err := ParseSemVer(input, true)
		if err != nil {
			b.Fatal(err)
		}
	}
}
