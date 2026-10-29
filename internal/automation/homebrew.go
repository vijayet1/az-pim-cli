package automation

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const HomebrewTap = "vijayet1/homebrew-tap"
const HomebrewFormulaPath = "Formula/az-pim-cli.rb"

type HomebrewUpdate struct {
	Tag         string `json:"tag"`
	Commit      string `json:"commit"`
	Arm64SHA256 string `json:"arm64_sha256"`
	AMD64SHA256 string `json:"amd64_sha256"`
}
type HomebrewResult struct {
	State     string `json:"state"`
	Tag       string `json:"tag"`
	TapCommit string `json:"tap_commit,omitempty"`
}

var homebrewVersion = regexp.MustCompile(`(?m)^  version "([^"\r\n]+)"$`)
var homebrewArchives = regexp.MustCompile(`(?m)^      url "([^"\r\n]+)"\n      sha256 "([0-9a-f]{64})"$`)
var homebrewDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

func homebrewAssetURL(tag, arch string) string {
	return "https://github.com/" + Fork + "/releases/download/" + tag + "/az-pim-cli-" + strings.TrimPrefix(tag, "v") + "-macos-" + arch + ".tar.gz"
}
func HomebrewPlanFromManifest(m ReleaseManifest) (HomebrewUpdate, error) {
	p := HomebrewUpdate{Tag: m.Tag, Commit: m.Commit}
	if err := ValidateManifest(m, m.Tag, m.Commit, m.Upstream); err != nil {
		return p, err
	}
	prefix := "az-pim-cli-" + strings.TrimPrefix(m.Tag, "v") + "-macos-"
	for _, a := range m.Assets {
		switch a.Name {
		case prefix + "arm64.tar.gz":
			p.Arm64SHA256 = a.SHA256
		case prefix + "x86_64.tar.gz":
			p.AMD64SHA256 = a.SHA256
		}
	}
	return p, nil
}

// UpdateHomebrewFormula changes metadata only; unknown formula layouts fail closed.
func UpdateHomebrewFormula(current []byte, p HomebrewUpdate) ([]byte, error) {
	next, err := ParseForkVersion(p.Tag)
	if err != nil {
		return nil, err
	}
	if !shaPattern.MatchString(p.Commit) || !homebrewDigest.MatchString(p.Arm64SHA256) || !homebrewDigest.MatchString(p.AMD64SHA256) {
		return nil, fmt.Errorf("invalid Homebrew source or checksum")
	}
	text := string(current)
	versions := homebrewVersion.FindAllStringSubmatch(text, -1)
	if len(versions) != 1 {
		return nil, fmt.Errorf("formula must contain exactly one stable version")
	}
	oldTag := "v" + versions[0][1]
	old, err := ParseForkVersion(oldTag)
	if err != nil {
		return nil, err
	}
	order := CompareVersions(next, old)
	if order < 0 {
		return nil, fmt.Errorf("refusing Homebrew downgrade from %s to %s", oldTag, p.Tag)
	}
	pairs := homebrewArchives.FindAllStringSubmatch(text, -1)
	if len(pairs) != 2 {
		return nil, fmt.Errorf("formula must contain exactly two expected macOS archive/checksum pairs")
	}
	seen := map[string]bool{}
	for _, pair := range pairs {
		arch, digest := "", ""
		switch pair[1] {
		case homebrewAssetURL(oldTag, "arm64"):
			arch = "arm64"
			digest = p.Arm64SHA256
		case homebrewAssetURL(oldTag, "x86_64"):
			arch = "x86_64"
			digest = p.AMD64SHA256
		default:
			return nil, fmt.Errorf("unexpected Homebrew archive URL")
		}
		if seen[arch] {
			return nil, fmt.Errorf("duplicate Homebrew architecture")
		}
		seen[arch] = true
		replacement := "      url \"" + homebrewAssetURL(p.Tag, arch) + "\"\n      sha256 \"" + digest + "\""
		text = strings.Replace(text, pair[0], replacement, 1)
	}
	text = homebrewVersion.ReplaceAllString(text, "  version \""+strings.TrimPrefix(p.Tag, "v")+"\"")
	if order == 0 && text != string(current) {
		return nil, fmt.Errorf("same-version Homebrew metadata differs; operator inspection required")
	}
	return []byte(text), nil
}

func PlanHomebrew(ctx context.Context, r Runner, dir string, o ReleaseOptions) (HomebrewUpdate, error) {
	result, err := CheckRelease(ctx, r, dir, o)
	if err != nil {
		return HomebrewUpdate{}, err
	}
	if result.State != "already-published" || result.Manifest == nil {
		return HomebrewUpdate{}, fmt.Errorf("homebrew requires a verified published stable release")
	}
	return HomebrewPlanFromManifest(*result.Manifest)
}

type tapFile struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	SHA      string `json:"sha"`
	Content  string `json:"content"`
}

func decodeTapFile(data []byte) (tapFile, []byte, error) {
	var f tapFile
	if err := json.Unmarshal(data, &f); err != nil {
		return f, nil, err
	}
	if f.Type != "file" || f.Encoding != "base64" || !shaPattern.MatchString(f.SHA) {
		return f, nil, fmt.Errorf("unexpected Homebrew formula response")
	}
	b, err := base64.StdEncoding.DecodeString(f.Content)
	return f, b, err
}
func UpdateHomebrewTap(ctx context.Context, r Runner, dir string, o ReleaseOptions) (HomebrewResult, error) {
	result := HomebrewResult{Tag: o.Tag}
	token := strings.TrimSpace(os.Getenv("HOMEBREW_TAP_TOKEN"))
	if token == "" {
		return result, fmt.Errorf("environment secret HOMEBREW_TAP_TOKEN is required")
	}
	plan, err := PlanHomebrew(ctx, r, dir, o)
	if err != nil {
		return result, err
	}
	path := "repos/" + HomebrewTap + "/contents/" + HomebrewFormulaPath
	call := func(args ...string) ([]byte, error) {
		return r.Run(ctx, dir, []string{"GH_TOKEN=" + token}, "gh", append([]string{"api", "--hostname", "github.com", path}, args...)...)
	}
	for attempt := 0; attempt < 3; attempt++ {
		data, err := call("--method", "GET", "--raw-field", "ref=main")
		if err != nil {
			return result, err
		}
		file, current, err := decodeTapFile(data)
		if err != nil {
			return result, err
		}
		updated, err := UpdateHomebrewFormula(current, plan)
		if err != nil {
			return result, err
		}
		if bytes.Equal(current, updated) {
			result.State = "noop"
			return result, nil
		}
		message := "chore: update az-pim-cli to " + strings.TrimPrefix(o.Tag, "v") + "\n\nSource: " + "https://github.com/" + Fork + "/releases/tag/" + o.Tag + "\nCommit: " + o.Commit
		data, err = call("--method", "PUT", "--raw-field", "branch=main", "--raw-field", "sha="+file.SHA, "--raw-field", "message="+message, "--raw-field", "content="+base64.StdEncoding.EncodeToString(updated))
		if err != nil {
			var commandErr *CommandError
			if errors.As(err, &commandErr) && commandErr.HTTPStatus == 409 && attempt < 2 {
				continue
			}
			return result, err
		}
		var response struct {
			Commit struct {
				SHA string `json:"sha"`
			} `json:"commit"`
		}
		if err = json.Unmarshal(data, &response); err != nil {
			return result, err
		}
		if !shaPattern.MatchString(response.Commit.SHA) {
			return result, fmt.Errorf("tap update response lacks commit SHA; inspect target before retrying")
		}
		verifyData, err := call("--method", "GET", "--raw-field", "ref="+response.Commit.SHA)
		if err != nil {
			return result, err
		}
		_, verified, err := decodeTapFile(verifyData)
		if err != nil {
			return result, err
		}
		if !bytes.Equal(verified, updated) {
			return result, fmt.Errorf("tap commit does not contain expected formula")
		}
		result.State = "updated"
		result.TapCommit = response.Commit.SHA
		return result, nil
	}
	return result, fmt.Errorf("homebrew formula changed concurrently; retry after inspection")
}
