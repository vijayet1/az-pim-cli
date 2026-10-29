package automation

import (
	"encoding/json"
	"strings"
	"testing"
)

const testSHA = "0123456789abcdef0123456789abcdef01234567"

func fixturePin() UpstreamPin {
	return UpstreamPin{1, Upstream, "v1.18.0", testSHA, "https://github.com/netr0m/az-pim-cli/releases/tag/v1.18.0"}
}
func TestParseStableRelease(t *testing.T) {
	good := `{"tag_name":"v1.18.0","html_url":"https://github.com/netr0m/az-pim-cli/releases/tag/v1.18.0","draft":false,"prerelease":false}`
	for _, tc := range []struct {
		name, input string
		bad         bool
	}{
		{"stable", good, false}, {"draft", strings.Replace(good, `"draft":false`, `"draft":true`, 1), true},
		{"prerelease", strings.Replace(good, `"prerelease":false`, `"prerelease":true`, 1), true},
		{"missing flags", `{"tag_name":"v1.18.0"}`, true}, {"invalid JSON", "{", true},
		{"leading zero", strings.ReplaceAll(good, "v1.18.0", "v01.18.0"), true},
		{"different repository", strings.ReplaceAll(good, "netr0m", "someone"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, e := ParseStableRelease([]byte(tc.input))
			if (e != nil) != tc.bad {
				t.Fatalf("release=%+v error=%v", r, e)
			}
			if !tc.bad && r.Tag != "v1.18.0" {
				t.Fatal(r)
			}
		})
	}
}
func TestValidatePin(t *testing.T) {
	for _, field := range []string{"valid", "repository", "schema", "tag", "commit", "url"} {
		t.Run(field, func(t *testing.T) {
			p := fixturePin()
			switch field {
			case "repository":
				p.Repository = "other/repo"
			case "schema":
				p.SchemaVersion = 2
			case "tag":
				p.Tag = "v1.0.0-rc.1"
			case "commit":
				p.Commit = "1234"
			case "url":
				p.ReleaseURL = "https://github.com/other/repo/releases/tag/v1.18.0"
			}
			if e := ValidatePin(p); (e != nil) != (field != "valid") {
				t.Fatalf("%+v: %v", p, e)
			}
		})
	}
	b, _ := json.Marshal(fixturePin())
	p, e := DecodePin(b)
	if e != nil || p.Commit != testSHA {
		t.Fatalf("%+v %v", p, e)
	}
	if _, e := DecodePin(append(b, []byte(` {}`)...)); e == nil {
		t.Fatal("accepted trailing JSON")
	}
}
func TestParseForkVersion(t *testing.T) {
	for _, tag := range []string{"v0.0.0", "v1.17.0", "v18446744073709551615.0.0"} {
		if _, e := ParseForkVersion(tag); e != nil {
			t.Fatal(tag, e)
		}
	}
	for _, tag := range []string{"1.0.0", "v01.0.0", "v1.0", "v1.0.0-fork.1", "v1.0.0+build", "v18446744073709551616.0.0", "v1.0.0\n"} {
		if _, e := ParseForkVersion(tag); e == nil {
			t.Fatal("accepted", tag)
		}
	}
}
func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{{"v1.10.0", "v1.9.9", 1}, {"v2.0.0", "v1.99.99", 1}, {"v1.0.0", "v1.0.0", 0}, {"v1.0.0", "v1.0.1", -1}} {
		a, _ := ParseForkVersion(tc.a)
		b, _ := ParseForkVersion(tc.b)
		if got := CompareVersions(a, b); got != tc.want {
			t.Fatalf("%s %s got %d", tc.a, tc.b, got)
		}
	}
}
func TestClassifyRelease(t *testing.T) {
	for _, tc := range []struct {
		tag                 string
		published, complete bool
		want                string
	}{{"", false, false, "new"}, {testSHA, false, false, "retry-draft"}, {testSHA, true, true, "already-published"}, {testSHA, true, false, ""}, {strings.Repeat("a", 40), false, false, ""}} {
		got, e := ClassifyRelease(tc.tag, testSHA, tc.published, tc.complete)
		if tc.want == "" {
			if e == nil {
				t.Fatal("accepted unsafe state")
			}
		} else if e != nil || got != tc.want {
			t.Fatalf("%s %v", got, e)
		}
	}
}
