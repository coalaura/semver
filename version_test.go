package main

import "testing"

type semVerTestCase struct {
	name        string
	input       string
	allowSuffix bool
	expected    SemVer
	shouldErr   bool
}

func TestParseSemVer(t *testing.T) {
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
			name:        "standard full version with spaces",
			input:       "  1.2.3     ",
			allowSuffix: false,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
		},
		{
			name:        "major only",
			input:       "1",
			allowSuffix: false,
			expected:    SemVer{Major: 1, HasMinor: false, HasPatch: false},
		},
		{
			name:        "major and minor only",
			input:       "1.2",
			allowSuffix: false,
			expected:    SemVer{Major: 1, Minor: 2, HasMinor: true, HasPatch: false},
		},
		{
			name:        "lowercase v prefix",
			input:       "v1.2.3",
			allowSuffix: false,
			expected:    SemVer{Prefix: 'v', Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
		},
		{
			name:        "prefix on partial",
			input:       "v2",
			allowSuffix: false,
			expected:    SemVer{Prefix: 'v', Major: 2, HasMinor: false, HasPatch: false},
		},
		{
			name:        "pre-release suffix allowed",
			input:       "1.2.3-alpha.1",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-alpha.1"},
		},
		{
			name:        "build metadata suffix allowed",
			input:       "1.2.3+build.2026",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "+build.2026"},
		},
		{
			name:        "pre-release and build metadata",
			input:       "1.2.3-rc.1+sha.1234",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-rc.1+sha.1234"},
		},
		{
			name:        "suffix disallowed returns error",
			input:       "1.2.3-beta",
			allowSuffix: false,
			shouldErr:   true,
		},
		{
			name:        "suffix on minor-only version",
			input:       "1.2-beta",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, HasMinor: true, HasPatch: false, Suffix: "-beta"},
		},
		{
			name:        "suffix on major-only version",
			input:       "1-beta",
			allowSuffix: true,
			expected:    SemVer{Major: 1, HasMinor: false, HasPatch: false, Suffix: "-beta"},
		},
		{
			name:        "utf-8 multibyte suffix byte collision",
			input:       "1.2.3-beta.à",
			allowSuffix: true,
			expected:    SemVer{Major: 1, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true, Suffix: "-beta.à"},
		},
		{
			name:        "uint16 max boundaries",
			input:       "65535.65535.65535",
			allowSuffix: false,
			expected:    SemVer{Major: 65535, Minor: 65535, Patch: 65535, HasMinor: true, HasPatch: true},
		},
		{
			name:        "uint16 overflow major",
			input:       "65536.0.0",
			allowSuffix: false,
			shouldErr:   true,
		},
		{
			name:        "uint16 overflow patch",
			input:       "1.0.70000",
			allowSuffix: false,
			shouldErr:   true,
		},

		{name: "empty input", input: "", allowSuffix: true, shouldErr: true},
		{name: "only prefix", input: "v", allowSuffix: true, shouldErr: true},
		{name: "double prefix", input: "vv1.0.0", allowSuffix: true, shouldErr: true},
		{name: "negative numbers", input: "-1.0.0", allowSuffix: true, shouldErr: true},
		{name: "trailing dot", input: "1.2.", allowSuffix: false, shouldErr: true},
		{name: "leading dot", input: ".1.2", allowSuffix: false, shouldErr: true},
		{name: "empty segment", input: "1..2", allowSuffix: false, shouldErr: true},
		{name: "four segments", input: "1.2.3.4", allowSuffix: false, shouldErr: true},
		{name: "non-numeric segment", input: "1.two.3", allowSuffix: true, shouldErr: true},
		{name: "trailing hyphen with no pre-release", input: "1.2.3-", allowSuffix: true, shouldErr: true},
		{name: "trailing plus with no build metadata", input: "1.2.3+", allowSuffix: true, shouldErr: true},
		{name: "dot after patch without suffix", input: "1.2.3.", allowSuffix: false, shouldErr: true},
		{name: "uppercase V prefix", input: "V1.2.3", allowSuffix: false, shouldErr: true},
		{name: "leading zeros in numeric identifiers", input: "01.2.3", allowSuffix: false, shouldErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ver, err := ParseSemVer(test.input, test.allowSuffix)
			if err != nil {
				if !test.shouldErr {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			} else {
				if test.shouldErr {
					t.Fatal("expected error, got nil")
				}
			}

			if !ver.Equal(test.expected) || ver.Prefix != test.expected.Prefix {
				t.Fatalf("expected %q, got %q", test.expected.String(), ver.String())
			}
		})
	}
}
