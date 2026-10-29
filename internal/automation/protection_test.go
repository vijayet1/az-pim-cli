package automation

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type protectionRunner struct{ environment, policies string }

func (r protectionRunner) Run(_ context.Context, _ string, _ []string, name string, args ...string) ([]byte, error) {
	if name != "gh" {
		return nil, fmt.Errorf("unexpected command")
	}
	for _, a := range args {
		if a == "repos/"+Fork+"/environments/fork-release" {
			return []byte(r.environment), nil
		}
		if a == "repos/"+Fork+"/environments/fork-release/deployment-branch-policies" {
			return []byte(r.policies), nil
		}
	}
	return nil, fmt.Errorf("unexpected request")
}
func TestProtectedReleaseEnvironment(t *testing.T) {
	good := `{"can_admins_bypass":false,"deployment_branch_policy":{"custom_branch_policies":true,"protected_branches":false},"protection_rules":[{"type":"required_reviewers","reviewers":[{"type":"User","reviewer":{"login":"maintainer"}}]}]}`
	policy := `{"branch_policies":[{"name":"main","type":"branch"}]}`
	for _, tc := range []struct {
		name, env, policy string
		bad               bool
	}{{"protected", good, policy, false}, {"empty", "{}", policy, true}, {"no reviewer", strings.Replace(good, `"required_reviewers"`, `"wait_timer"`, 1), policy, true}, {"admin bypass", strings.Replace(good, `"can_admins_bypass":false`, `"can_admins_bypass":true`, 1), policy, true}, {"all branches", good, `{"branch_policies":[{"name":"*","type":"branch"}]}`, true}, {"tag branch", good, `{"branch_policies":[{"name":"main","type":"tag"}]}`, true}} {
		t.Run(tc.name, func(t *testing.T) {
			e := CheckReleaseEnvironment(context.Background(), protectionRunner{tc.env, tc.policy}, t.TempDir())
			if (e != nil) != tc.bad {
				t.Fatalf("error=%v", e)
			}
		})
	}
}
