package automation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

const Fork = "vijayet1/az-pim-cli"
const Upstream = "netr0m/az-pim-cli"
const PinPath = ".github/upstream-release.json"

var versionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
var shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

type Version struct{ Major, Minor, Patch uint64 }
type StableRelease struct{ Tag, URL string }
type UpstreamPin struct {
	SchemaVersion int    `json:"schema_version"`
	Repository    string `json:"repository"`
	Tag           string `json:"tag"`
	Commit        string `json:"commit"`
	ReleaseURL    string `json:"release_url"`
}

func ParseForkVersion(tag string) (Version, error) {
	m := versionPattern.FindStringSubmatch(tag)
	if m == nil {
		return Version{}, fmt.Errorf("expected stable vMAJOR.MINOR.PATCH: %q", tag)
	}
	var n [3]uint64
	for i := range n {
		var err error
		n[i], err = strconv.ParseUint(m[i+1], 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version component: %w", err)
		}
	}
	return Version{n[0], n[1], n[2]}, nil
}
func CompareVersions(a, b Version) int {
	aa, bb := [3]uint64{a.Major, a.Minor, a.Patch}, [3]uint64{b.Major, b.Minor, b.Patch}
	for i := range aa {
		if aa[i] < bb[i] {
			return -1
		}
		if aa[i] > bb[i] {
			return 1
		}
	}
	return 0
}
func ParseStableRelease(data []byte) (StableRelease, error) {
	var r struct {
		Tag   string `json:"tag_name"`
		URL   string `json:"html_url"`
		Draft *bool  `json:"draft"`
		Pre   *bool  `json:"prerelease"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return StableRelease{}, err
	}
	if _, err := ParseForkVersion(r.Tag); err != nil {
		return StableRelease{}, err
	}
	if r.Draft == nil || r.Pre == nil || *r.Draft || *r.Pre || r.URL != releaseURL(r.Tag) {
		return StableRelease{}, fmt.Errorf("response is not a published stable upstream release")
	}
	return StableRelease{r.Tag, r.URL}, nil
}
func releaseURL(tag string) string { return "https://github.com/" + Upstream + "/releases/tag/" + tag }
func ValidatePin(p UpstreamPin) error {
	if _, err := ParseForkVersion(p.Tag); err != nil {
		return err
	}
	if p.SchemaVersion != 1 || p.Repository != Upstream || !shaPattern.MatchString(p.Commit) || p.ReleaseURL != releaseURL(p.Tag) {
		return fmt.Errorf("invalid upstream source pin")
	}
	return nil
}
func DecodePin(data []byte) (UpstreamPin, error) {
	var p UpstreamPin
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return p, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return p, fmt.Errorf("unexpected trailing pin data")
	}
	return p, ValidatePin(p)
}
func ClassifyRelease(tagCommit, selectedCommit string, published, complete bool) (string, error) {
	if !shaPattern.MatchString(selectedCommit) {
		return "", fmt.Errorf("selected commit must be a full lowercase SHA")
	}
	if tagCommit != "" && tagCommit != selectedCommit {
		return "", fmt.Errorf("immutable tag points to another commit")
	}
	if published {
		if tagCommit == "" || !complete {
			return "", fmt.Errorf("published release is incomplete or missing its tag")
		}
		return "already-published", nil
	}
	if tagCommit != "" {
		return "retry-draft", nil
	}
	return "new", nil
}
