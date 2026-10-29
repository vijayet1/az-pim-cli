package automation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCandidate(t *testing.T) {
	for _, scenario := range []string{"bootstrap", "valid", "missing pin", "invalid ancestry", "retargeted", "downgrade", "invalid base"} {
		t.Run(scenario, func(t *testing.T) {
			f := newSyncFixture(t)
			c := filepath.Join(f.root, "checkout")
			testGit(t, f.root, "clone", f.fork, c)
			if scenario != "bootstrap" {
				pin := UpstreamPin{1, Upstream, "v1.18.0", f.sha, releaseURL("v1.18.0")}
				if scenario != "invalid ancestry" {
					testGit(t, c, "fetch", "--no-tags", f.up, f.sha)
					testGit(t, c, "merge", "--no-ff", "FETCH_HEAD", "-m", "merge release")
				}
				if scenario == "downgrade" {
					testGit(t, f.up, "tag", "v1.19.0", f.sha)
					pin.Tag = "v1.19.0"
					pin.ReleaseURL = releaseURL(pin.Tag)
				}
				os.MkdirAll(filepath.Join(c, ".github"), 0755)
				b, _ := json.Marshal(pin)
				os.WriteFile(filepath.Join(c, PinPath), b, 0644)
				testGit(t, c, "add", ".")
				testGit(t, c, "commit", "-m", "record pin")
				if scenario == "missing pin" || scenario == "downgrade" {
					testGit(t, c, "update-ref", "refs/remotes/origin/main", "HEAD")
				}
				if scenario == "missing pin" {
					testGit(t, c, "rm", PinPath)
					testGit(t, c, "commit", "-m", "remove pin")
				}
				if scenario == "downgrade" {
					pin.Tag = "v1.18.0"
					pin.ReleaseURL = releaseURL(pin.Tag)
					b, _ := json.Marshal(pin)
					os.WriteFile(filepath.Join(c, PinPath), b, 0644)
					testGit(t, c, "add", ".")
					testGit(t, c, "commit", "-m", "downgrade pin")
				}
				if scenario == "retargeted" {
					testGit(t, f.up, "tag", "-f", "v1.18.0", "HEAD")
				}
			}
			base := "refs/remotes/origin/main"
			if scenario == "invalid base" {
				base = "HEAD"
			}
			e := ValidateCandidate(context.Background(), f, c, base)
			wantErr := scenario != "bootstrap" && scenario != "valid"
			if (e != nil) != wantErr {
				t.Fatalf("error=%v", e)
			}
		})
	}
}
