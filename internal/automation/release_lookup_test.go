package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const draftReleaseResponse = `{"id":42,"tag_name":"v1.17.0","draft":true,"prerelease":false,"assets":[{"id":1,"name":"fixture","size":1}]}`

type releaseLookupRunner struct {
	published    string
	publishedErr error
	pages        map[int]string
	listErr      error
}

func (r releaseLookupRunner) Run(_ context.Context, _ string, _ []string, name string, args ...string) ([]byte, error) {
	if name != "gh" || len(args) != 4 || args[0] != "api" || args[1] != "--hostname" || args[2] != "github.com" {
		return nil, fmt.Errorf("unexpected command")
	}
	if args[3] == "repos/"+Fork+"/releases/tags/v1.17.0" {
		return []byte(r.published), r.publishedErr
	}
	prefix := "repos/" + Fork + "/releases?per_page=100&page="
	if strings.HasPrefix(args[3], prefix) {
		if r.listErr != nil {
			return nil, r.listErr
		}
		page, err := strconv.Atoi(strings.TrimPrefix(args[3], prefix))
		if err != nil {
			return nil, err
		}
		if body, ok := r.pages[page]; ok {
			return []byte(body), nil
		}
	}
	return nil, fmt.Errorf("unexpected API request %s", args[3])
}

func TestReleaseLookupFindsDraftAfterPublishedEndpoint404(t *testing.T) {
	runner := releaseLookupRunner{
		publishedErr: &CommandError{HTTPStatus: 404},
		pages:        map[int]string{1: "[" + draftReleaseResponse + "]"},
	}
	rel, err := (commands{context.Background(), runner, t.TempDir()}).release("v1.17.0")
	if err != nil || rel == nil {
		t.Fatalf("draft not discovered through releases list: release=%+v error=%v", rel, err)
	}
	if rel.ID != 42 || rel.Draft == nil || !*rel.Draft || len(rel.Assets) != 1 || rel.Assets[0].Name != "fixture" {
		t.Fatalf("wrong draft or incomplete asset metadata: %+v", rel)
	}
}

func TestReleaseLookupPublishedFastPath(t *testing.T) {
	runner := releaseLookupRunner{
		published: strings.Replace(draftReleaseResponse, `"draft":true`, `"draft":false`, 1),
		listErr:   fmt.Errorf("published releases must not require listing"),
	}
	rel, err := (commands{context.Background(), runner, t.TempDir()}).release("v1.17.0")
	if err != nil || rel == nil || *rel.Draft || rel.ID != 42 {
		t.Fatalf("published lookup failed: release=%+v error=%v", rel, err)
	}
}

func TestReleaseLookupFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		runner releaseLookupRunner
	}{
		{"published API denied", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 403}}},
		{"draft list denied", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, listErr: &CommandError{HTTPStatus: 403}}},
		{"duplicate drafts", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, pages: map[int]string{1: "[" + draftReleaseResponse + "," + strings.Replace(draftReleaseResponse, `"id":42`, `"id":43`, 1) + "]"}}},
		{"invalid list", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, pages: map[int]string{1: `{}`}}},
		{"null list", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, pages: map[int]string{1: `null`}}},
		{"prerelease draft", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, pages: map[int]string{1: "[" + strings.Replace(draftReleaseResponse, `"prerelease":false`, `"prerelease":true`, 1) + "]"}}},
		{"missing draft identity", releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, pages: map[int]string{1: `[{"tag_name":"v1.17.0","draft":true,"prerelease":false}]`}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if rel, err := (commands{context.Background(), tc.runner, t.TempDir()}).release("v1.17.0"); err == nil {
				t.Fatalf("accepted unsafe lookup result: %+v", rel)
			}
		})
	}
}

func TestReleaseLookupPaginatesAndHandlesAbsence(t *testing.T) {
	unrelated := json.RawMessage(`{"id":1,"tag_name":"v1.0.0","draft":false,"prerelease":false}`)
	first := make([]json.RawMessage, 100)
	for i := range first {
		first[i] = unrelated
	}
	body, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, found := range []bool{false, true} {
		last := "[]"
		if found {
			last = "[" + draftReleaseResponse + "]"
		}
		runner := releaseLookupRunner{publishedErr: &CommandError{HTTPStatus: 404}, pages: map[int]string{1: string(body), 2: last}}
		rel, err := (commands{context.Background(), runner, t.TempDir()}).release("v1.17.0")
		if err != nil || (rel != nil) != found {
			t.Fatalf("found=%v release=%+v error=%v", found, rel, err)
		}
	}
}

func TestVerifyReleaseReportsMissingDraft(t *testing.T) {
	f := newPublishFixture(t)
	err := VerifyRelease(context.Background(), f, f.checkout, ReleaseOptions{Repository: Fork, Tag: f.tag, Commit: f.selected, Ref: "refs/heads/main"}, "missing-artifacts.json")
	if err == nil || !strings.Contains(err.Error(), "draft release v1.17.0 not found") {
		t.Fatalf("missing draft diagnostic: %v", err)
	}
	if len(f.assetBytes) != 0 {
		t.Fatal("wrote assets without a draft")
	}
}

// Opt-in and read-only: reproduce the contract against an actual unpublished draft.
func TestGitHubDraftLookup(t *testing.T) {
	tag := os.Getenv("FORK_TEST_DRAFT_TAG")
	if tag == "" {
		t.Skip("set FORK_TEST_DRAFT_TAG to an existing draft tag for a read-only GitHub contract check")
	}
	if _, err := ParseForkVersion(tag); err != nil {
		t.Fatal(err)
	}
	c := commands{context.Background(), ExecRunner{}, t.TempDir()}
	if _, err := c.api("repos/" + Fork + "/releases/tags/" + tag); !is404(err) {
		t.Fatalf("expected published-only endpoint to exclude this draft, got %v", err)
	}
	rel, err := c.release(tag)
	if err != nil || rel == nil || rel.Draft == nil || !*rel.Draft || rel.Tag != tag || len(rel.Assets) == 0 {
		t.Fatalf("existing draft not found: release=%+v error=%v", rel, err)
	}
	tags, err := c.remoteTags()
	if err != nil {
		t.Fatal(err)
	}
	commit := tags[tag]
	if !shaPattern.MatchString(commit) {
		t.Fatal("draft tag does not resolve to a commit")
	}
	pinData, err := c.gh("api", "--hostname", "github.com", "repos/"+Fork+"/contents/"+PinPath+"?ref="+commit, "-H", "Accept: application/vnd.github.raw+json")
	if err != nil {
		t.Fatal(err)
	}
	pin, err := DecodePin(pinData)
	if err != nil {
		t.Fatal(err)
	}
	data, err := c.api("repos/" + Fork + "/releases/" + strconv.FormatInt(rel.ID, 10))
	if err != nil {
		t.Fatal(err)
	}
	var metadata struct {
		Assets []struct {
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err = json.Unmarshal(data, &metadata); err != nil {
		t.Fatal(err)
	}
	manifest := ReleaseManifest{SchemaVersion: 1, Repository: Fork, Tag: tag, Commit: commit, Upstream: pin}
	for _, asset := range metadata.Assets {
		if !strings.HasPrefix(asset.Digest, "sha256:") {
			t.Fatal("GitHub asset digest missing")
		}
		manifest.Assets = append(manifest.Assets, AssetDigest{Name: asset.Name, Size: asset.Size, SHA256: strings.TrimPrefix(asset.Digest, "sha256:")})
	}
	if err = ValidateManifest(manifest, tag, commit, pin); err != nil {
		t.Fatal(err)
	}
	if err = c.verifyRemote(rel, manifest, false); err != nil {
		t.Fatal(err)
	}
	t.Logf("Verified draft %s: release ID %d, %d assets and checksums; no writes performed", rel.Tag, rel.ID, len(rel.Assets))
}
