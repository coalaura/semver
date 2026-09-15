package semver

import "strings"

type SemVer struct {
	Major uint16
	Minor uint16
	Patch uint16

	HasMinor bool
	HasPatch bool
	Valid    bool

	Prefix byte
	Suffix string
}

// IsValid reports whether s was produced by a successful parse or constructor.
func (s SemVer) IsValid() bool {
	return s.Valid
}

// IsInvalid reports whether s does not represent a successfully parsed or constructed version.
func (s SemVer) IsInvalid() bool {
	return !s.Valid
}

// HigherThan reports whether s has higher precedence than b.
func (s SemVer) HigherThan(b SemVer) bool {
	return s.Compare(b) > 0
}

// Equal compares validity and parsed identity, ignoring the optional v prefix.
func (s SemVer) Equal(b SemVer) bool {
	if s.Valid != b.Valid || s.Major != b.Major || s.Suffix != b.Suffix {
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

// MajorString returns the major component, with the v prefix when present and no minor or patch components.
func (s SemVer) MajorString() string {
	length := decimalLength(s.Major)

	if s.Prefix != 0 {
		length++
	}

	var sb strings.Builder

	sb.Grow(length)

	if s.Prefix != 0 {
		sb.WriteByte(s.Prefix)
	}

	writeUint16(&sb, s.Major)

	return sb.String()
}

// String returns the canonical string form of the version, including the v prefix and any suffix.
func (s SemVer) String() string {
	length := decimalLength(s.Major) + len(s.Suffix)

	if s.Prefix != 0 {
		length++
	}

	if s.HasMinor {
		length += 1 + decimalLength(s.Minor)

		if s.HasPatch {
			length += 1 + decimalLength(s.Patch)
		}
	}

	var sb strings.Builder

	sb.Grow(length)

	if s.Prefix != 0 {
		sb.WriteByte(s.Prefix)
	}

	writeUint16(&sb, s.Major)

	if s.HasMinor {
		sb.WriteByte('.')

		writeUint16(&sb, s.Minor)

		if s.HasPatch {
			sb.WriteByte('.')

			writeUint16(&sb, s.Patch)
		}
	}

	if s.Suffix != "" {
		sb.WriteString(s.Suffix)
	}

	return sb.String()
}

// Compare compares semantic-version precedence. Prefixes, missing zero components and build metadata do not affect precedence.
func (s SemVer) Compare(b SemVer) int {
	if s.Major != b.Major {
		if s.Major < b.Major {
			return -1
		}

		return 1
	}

	var sMinor uint16

	if s.HasMinor {
		sMinor = s.Minor
	}

	var bMinor uint16

	if b.HasMinor {
		bMinor = b.Minor
	}

	if sMinor != bMinor {
		if sMinor < bMinor {
			return -1
		}

		return 1
	}

	var sPatch uint16

	if s.HasMinor && s.HasPatch {
		sPatch = s.Patch
	}

	var bPatch uint16

	if b.HasMinor && b.HasPatch {
		bPatch = b.Patch
	}

	if sPatch != bPatch {
		if sPatch < bPatch {
			return -1
		}

		return 1
	}

	sPreRelease, sHasPreRelease := preRelease(s.Suffix)
	bPreRelease, bHasPreRelease := preRelease(b.Suffix)

	if !sHasPreRelease {
		if bHasPreRelease {
			return 1
		}

		return 0
	}

	if !bHasPreRelease {
		return -1
	}

	return comparePreRelease(sPreRelease, bPreRelease)
}

// SetMajorOnly truncates the version to its major component, clearing minor, patch and suffix.
func (s *SemVer) SetMajorOnly() {
	s.HasMinor = false
	s.Minor = 0

	s.HasPatch = false
	s.Patch = 0

	s.Suffix = ""
}

// SetMajorMinorOnly truncates the version to its major and minor components, clearing patch and suffix.
func (s *SemVer) SetMajorMinorOnly() {
	s.HasPatch = false
	s.Patch = 0

	s.Suffix = ""
}
