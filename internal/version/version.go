package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverRegexp = regexp.MustCompile(`^(v)?(\d+)\.(\d+)\.(\d+)(-(\S+)\.(\d))?$`)

// Version wraps semver library with features related with parsed from
// commit message change type.
type Version struct {
	Prefix string

	Major int
	Minor int
	Patch int

	PreRelease string
	Build      int
}

// New returns new instance of version parsed from given string.
func New(s string) (v *Version, err error) {
	matches := semverRegexp.FindStringSubmatch(s)
	if len(matches) == 0 {
		return nil, ErrInvalidVersion
	}

	// v1.2.3-alpha.4
	// matches[0] = "v1.2.3"
	// matches[1] = 'v'
	// matches[2] = 1
	// matches[3] = 2
	// matches[4] = 3
	// matches[5] = "-alpha.1"
	// matches[6] = "alpha"
	// matches[7] = 4

	v = &Version{
		Prefix: matches[1],
	}

	v.Major, err = strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("%w: parse major: %w", ErrInvalidVersion, err)
	}
	v.Minor, err = strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("%w: parse minor: %w", ErrInvalidVersion, err)
	}
	v.Patch, err = strconv.Atoi(matches[4])
	if err != nil {
		return nil, fmt.Errorf("%w: parse patch: %w", ErrInvalidVersion, err)
	}

	if len(matches[5]) != 0 {
		v.PreRelease = matches[6]
		v.Build, err = strconv.Atoi(matches[7])
		if err != nil {
			return nil, fmt.Errorf("%w: parse build: %w", ErrInvalidVersion, err)
		}
	}

	return
}

// Next returns next version depending on given change type. In case of unknown
// or not versioned change type returned version stays the same.
func (v Version) Next(t ChangeType, pre ...string) *Version {
	preRelease := ""
	if len(pre) != 0 && pre[0] != "" {
		preRelease = pre[0]
	}

	if preRelease == "" {
		v.Bump(t)
	} else {
		v.BumpPreRelease(t, preRelease)
	}

	return &v
}

// IsPreRelease returns true if version has pre-release part.
func (v *Version) IsPreRelease() bool {
	return v.PreRelease != ""
}

// BumpMajor bumps major version.
func (v *Version) BumpMajor() {
	v.Major++
	v.BumpMinor()
	v.Minor = 0
}

// BumpMinor bumps minor version.
func (v *Version) BumpMinor() {
	v.Minor++
	v.BumpPatch()
	v.Patch = 0
}

// BumpPatch bumps patch version.
func (v *Version) BumpPatch() {
	v.Patch++
	v.PreRelease = ""
	v.Build = 0
}

// BumpPreRelease bumps pre-release build version.
func (v *Version) BumpPreRelease(change ChangeType, pre string) {
	latestChange := v.LatestChangeType()
	if v.PreRelease == pre && change <= latestChange {
		v.Build++
	} else {
		v.Bump(change)

		v.PreRelease = pre
		v.Build = 1
	}
}

// Bump changes version based on given change type.
func (v *Version) Bump(change ChangeType) {
	if v.IsPreRelease() {
		latestChange := v.LatestChangeType()
		if change <= latestChange {
			v.PreRelease = ""
			v.Build = 0

			return
		}

		v.Revert()
	}

	switch change {
	case ChangeTypeMajor:
		v.BumpMajor()
	case ChangeTypeMinor:
		v.BumpMinor()
	case ChangeTypePatch:
		v.BumpPatch()
	}
}

// Revert changes version to previous one.
func (v *Version) Revert() {
	latest := v.LatestChangeType()

	switch latest {
	case ChangeTypeMajor:
		v.Major = max(0, v.Major-1)
	case ChangeTypeMinor:
		v.Minor = max(0, v.Minor-1)
	case ChangeTypePatch:
		v.Patch = max(0, v.Patch-1)
	}
}

// LatestChangeType returns latest change type based on current version.
func (v *Version) LatestChangeType() ChangeType {
	switch {
	case v.Patch != 0:
		return ChangeTypePatch
	case v.Minor != 0:
		return ChangeTypeMinor
	case v.Major != 0:
		return ChangeTypeMajor
	default:
		return ChangeTypeNone
	}
}

func (v *Version) IsGreater(other *Version) bool {
	isGreater := func(a, b int) bool {
		return a > b
	}

	switch {
	case v.Major != other.Major:
		return isGreater(v.Major, other.Major)
	case v.Minor != other.Minor:
		return isGreater(v.Minor, other.Minor)
	case v.Patch != other.Patch:
		return isGreater(v.Patch, other.Patch)
	case v.Build != other.Build && v.PreRelease == other.PreRelease:
		return isGreater(v.Build, other.Build)
	}
	return false
}

// String is a text formatted version name.
func (v *Version) String() string {
	s := fmt.Sprintf("%s%d.%d.%d", v.Prefix, v.Major, v.Minor, v.Patch)
	if len(v.PreRelease) != 0 {
		return fmt.Sprintf("%s-%s.%d", s, v.PreRelease, v.Build)
	}
	return s
}

// ChangeType is the type of change that commit has made in repository.
type ChangeType int

// Parse string into [ChangeType].
func (t *ChangeType) Parse(s string) {
	*t = commitTypes[strings.ToLower(strings.TrimSpace(s))]
}

const (
	// ChangeTypeNone - changes made in commit that doesn't match any
	// of the provided in configuration file version type.
	ChangeTypeNone ChangeType = iota
	// ChangeTypePatch - bump patch version.
	ChangeTypePatch
	// ChangeTypeMinor - bump minor version.
	ChangeTypeMinor
	// ChangeTypeMajor - bump major version.
	ChangeTypeMajor
)

var commitTypes = map[string]ChangeType{
	"none":  ChangeTypeNone,
	"major": ChangeTypeMajor,
	"minor": ChangeTypeMinor,
	"patch": ChangeTypePatch,
}

func (ct ChangeType) String() (s string) {
	for s, c := range commitTypes {
		if ct == c {
			return s
		}
	}
	return "none"
}

var ErrInvalidVersion = errors.New("invalid version")
