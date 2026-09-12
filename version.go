package semver

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrEmptyVersion     = errors.New("empty version")
	ErrInvalidVersion   = errors.New("invalid version")
	ErrInvalidSuffix    = errors.New("invalid version suffix")
	ErrDisallowedSuffix = errors.New("disallowed version suffix")
)

// NewEmptySemVer returns a zeroed SemVer with minor and patch components marked as present.
func NewEmptySemVer() SemVer {
	return SemVer{
		Major: 0,
		Minor: 0,
		Patch: 0,

		HasMinor: true,
		HasPatch: true,
	}
}

// ParseSemVer parses input into a SemVer. An optional v prefix is accepted and a suffix is only allowed when allowSuffix is true.
func ParseSemVer(input string, allowSuffix bool) (SemVer, error) {
	var version SemVer

	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return version, ErrEmptyVersion
	}

	index := 0

	if input[index] == 'v' {
		version.Prefix = 'v'

		index++

		if index == len(input) {
			return version, ErrInvalidVersion
		}
	}

	for component := range 3 {
		if index == len(input) || !isASCIIDigit(input[index]) {
			return version, ErrInvalidVersion
		}

		if input[index] == '0' && index+1 < len(input) && isASCIIDigit(input[index+1]) {
			return version, ErrInvalidVersion
		}

		var value uint32

		for index < len(input) && isASCIIDigit(input[index]) {
			value = value*10 + uint32(input[index]-'0')
			if value > 65535 {
				return version, ErrInvalidVersion
			}

			index++
		}

		switch component {
		case 0:
			version.Major = uint16(value)
		case 1:
			version.Minor = uint16(value)
			version.HasMinor = true
		case 2:
			version.Patch = uint16(value)
			version.HasPatch = true
		}

		if index == len(input) {
			return version, nil
		}

		switch input[index] {
		case '.':
			if component == 2 {
				return version, ErrInvalidVersion
			}

			index++

			if index == len(input) {
				return version, ErrInvalidVersion
			}
		case '-', '+':
			if !allowSuffix {
				return version, ErrDisallowedSuffix
			}

			suffix := input[index:]

			err := validateSemVerSuffix(suffix)
			if err != nil {
				return version, err
			}

			version.Suffix = suffix

			return version, nil
		default:
			return version, ErrInvalidVersion
		}
	}

	return version, ErrInvalidVersion
}

func validateSemVerSuffix(suffix string) error {
	if len(suffix) < 2 || (suffix[0] != '-' && suffix[0] != '+') {
		return ErrInvalidSuffix
	}

	isBuild := suffix[0] == '+'
	identifierStart := 1
	isNumeric := true

	for index := 1; index < len(suffix); {
		ch := suffix[index]

		switch {
		case isASCIIDigit(ch):
			index++
		case isASCIIAlpha(ch) || ch == '-':
			isNumeric = false

			index++
		case ch == '.':
			if !validSuffixIdentifier(suffix, identifierStart, index, isNumeric, isBuild) {
				return ErrInvalidSuffix
			}

			identifierStart = index + 1
			isNumeric = true

			index++
		case ch == '+':
			if isBuild || !validSuffixIdentifier(suffix, identifierStart, index, isNumeric, false) {
				return ErrInvalidSuffix
			}

			isBuild = true
			identifierStart = index + 1
			isNumeric = true

			index++
		default:
			if ch < utf8.RuneSelf {
				return ErrInvalidSuffix
			}

			decoded, size := utf8.DecodeRuneInString(suffix[index:])
			if decoded == utf8.RuneError && size == 1 {
				return ErrInvalidSuffix
			}

			if !unicode.IsLetter(decoded) && !unicode.IsDigit(decoded) && !unicode.IsMark(decoded) {
				return ErrInvalidSuffix
			}

			isNumeric = false

			index += size
		}
	}

	if !validSuffixIdentifier(suffix, identifierStart, len(suffix), isNumeric, isBuild) {
		return ErrInvalidSuffix
	}

	return nil
}

func validSuffixIdentifier(input string, start, end int, isNumeric, isBuild bool) bool {
	if start == end {
		return false
	}

	if !isBuild && isNumeric && end-start > 1 && input[start] == '0' {
		return false
	}

	return true
}

func isASCIIDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isASCIIAlpha(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func decimalLength(value uint16) int {
	switch {
	case value >= 10000:
		return 5
	case value >= 1000:
		return 4
	case value >= 100:
		return 3
	case value >= 10:
		return 2
	}

	return 1
}

func writeUint16(sb *strings.Builder, value uint16) {
	divisor := uint16(10000)

	for divisor > 1 && value < divisor {
		divisor /= 10
	}

	for divisor != 0 {
		sb.WriteByte(byte(value/divisor) + '0')

		value %= divisor
		divisor /= 10
	}
}

func preRelease(suffix string) (string, bool) {
	if len(suffix) == 0 || suffix[0] != '-' {
		return "", false
	}

	index := strings.IndexByte(suffix, '+')
	if index >= 0 {
		return suffix[1:index], true
	}

	return suffix[1:], true
}

func comparePreRelease(a, b string) int {
	for {
		aIdentifier, aRemaining, aHasRemaining := nextIdentifier(a)
		bIdentifier, bRemaining, bHasRemaining := nextIdentifier(b)

		compared := compareIdentifier(aIdentifier, bIdentifier)
		if compared != 0 {
			return compared
		}

		switch {
		case !aHasRemaining && !bHasRemaining:
			return 0
		case !aHasRemaining:
			return -1
		case !bHasRemaining:
			return 1
		}

		a = aRemaining
		b = bRemaining
	}
}

func nextIdentifier(input string) (string, string, bool) {
	before, after, ok := strings.Cut(input, ".")
	if !ok {
		return input, "", false
	}

	return before, after, true
}

func compareIdentifier(a, b string) int {
	aNumeric := isNumericIdentifier(a)
	bNumeric := isNumericIdentifier(b)

	switch {
	case aNumeric && !bNumeric:
		return -1
	case !aNumeric && bNumeric:
		return 1
	case !aNumeric:
		return strings.Compare(a, b)
	}

	a = trimLeadingZeros(a)
	b = trimLeadingZeros(b)

	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}

		return 1
	}

	return strings.Compare(a, b)
}

func isNumericIdentifier(input string) bool {
	if len(input) == 0 {
		return false
	}

	for index := 0; index < len(input); index++ {
		if !isASCIIDigit(input[index]) {
			return false
		}
	}

	return true
}

func trimLeadingZeros(input string) string {
	var index int

	for index+1 < len(input) && input[index] == '0' {
		index++
	}

	return input[index:]
}
