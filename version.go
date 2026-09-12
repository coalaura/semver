package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/coalaura/byteconv"
)

var (
	ErrEmptyVersion     = errors.New("empty version")
	ErrInvalidVersion   = errors.New("invalid version")
	ErrInvalidSuffix    = errors.New("invalid version suffix")
	ErrDisallowedSuffix = errors.New("disallowed version suffix")
)

type SemVer struct {
	Major uint16
	Minor uint16
	Patch uint16

	HasMinor bool
	HasPatch bool

	Prefix byte
	Suffix string
}

func NewEmptySemVer() SemVer {
	return SemVer{
		Major: 0,
		Minor: 0,
		Patch: 0,

		HasMinor: true,
		HasPatch: true,
	}
}

func ParseSemVer(input string, allowSuffix bool) (SemVer, error) {
	var version SemVer

	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return version, ErrEmptyVersion
	}

	if input[0] == 'v' {
		version.Prefix = 'v'

		input = input[1:]
		if len(input) == 0 {
			return version, ErrInvalidVersion
		}
	}

	var (
		isMajor  = true
		isMinor  bool
		isPatch  bool
		isSuffix bool
		buf      bytes.Buffer
	)

	buf.Grow(4)

	push := func() error {
		if buf.Len() == 0 {
			return ErrInvalidVersion
		}

		b := buf.Bytes()

		if buf.Len() > 1 && b[0] == '0' {
			return ErrInvalidVersion
		}

		i, err := byteconv.ParseUint(b, 10, 64)
		if err != nil {
			return err
		}

		if i > math.MaxUint16 {
			return ErrInvalidVersion
		}

		if isMajor {
			version.Major = uint16(i)

			isMajor = false
			isMinor = true
		} else if isMinor {
			version.Minor = uint16(i)
			version.HasMinor = true

			isMinor = false
			isPatch = true
		} else if isPatch {
			version.Patch = uint16(i)
			version.HasPatch = true

			isPatch = false
		} else {
			return ErrInvalidVersion
		}

		buf.Reset()

		return nil
	}

	for i, ch := range input {
		if ch >= '0' && ch <= '9' {
			buf.WriteRune(ch)
		} else if !isSuffix {
			if ch == '.' {
				err := push()
				if err != nil {
					return version, err
				}

				if i+1 == len(input) {
					return version, ErrInvalidVersion
				}
			} else {
				if ch == '-' || ch == '+' {
					if !allowSuffix {
						return version, ErrDisallowedSuffix
					}
				} else {
					return version, ErrInvalidVersion
				}

				err := push()
				if err != nil {
					return version, err
				}

				if i+1 == len(input) {
					return version, ErrInvalidVersion
				}

				isSuffix = true

				buf.WriteRune(ch)
			}
		} else if isSuffix {
			if unicode.IsSpace(ch) {
				return version, ErrInvalidSuffix
			}

			buf.WriteRune(ch)
		} else {
			return version, ErrInvalidVersion
		}
	}

	if buf.Len() == 0 {
		if isMajor {
			return version, ErrInvalidVersion
		}
	} else {
		if !isSuffix {
			err := push()
			if err != nil {
				return version, err
			}
		} else {
			version.Suffix = buf.String()
		}
	}

	return version, nil
}

func (s SemVer) String() string {
	var sb strings.Builder

	if s.Prefix != 0 {
		sb.WriteByte(s.Prefix)
	}

	fmt.Fprint(&sb, s.Major)

	if s.HasMinor {
		sb.WriteByte('.')
		fmt.Fprint(&sb, s.Minor)

		if s.HasPatch {
			sb.WriteByte('.')
			fmt.Fprint(&sb, s.Patch)
		}
	}

	if s.Suffix != "" {
		sb.WriteString(s.Suffix)
	}

	return sb.String()
}

func (s SemVer) HigherThan(b SemVer) bool {
	if s.Major != b.Major {
		return s.Major > b.Major
	}

	if s.Minor != b.Minor {
		return s.Minor > b.Minor
	}

	if s.Patch != b.Patch {
		return s.Patch > b.Patch
	}

	return false
}

func (s SemVer) Equal(b SemVer) bool {
	if s.Major != b.Major || s.Suffix != b.Suffix {
		return false
	}

	if s.HasMinor != b.HasMinor || (s.HasMinor && s.Minor != b.Minor) {
		return false
	}

	if s.HasPatch != b.HasPatch || (s.HasPatch && s.Patch != b.Patch) {
		return false
	}

	return true
}
