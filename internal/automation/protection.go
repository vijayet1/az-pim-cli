package automation

import (
	"context"
	"encoding/json"
	"fmt"
)

// CheckReleaseEnvironment refuses writes if the approval boundary is missing.
func CheckReleaseEnvironment(ctx context.Context, r Runner, dir string) error {
	c := commands{ctx, r, dir}
	base := "repos/" + Fork + "/environments/fork-release"
	b, e := c.api(base)
	if e != nil {
		return e
	}
	var environment struct {
		CanAdminsBypass        *bool `json:"can_admins_bypass"`
		DeploymentBranchPolicy struct {
			Custom    bool `json:"custom_branch_policies"`
			Protected bool `json:"protected_branches"`
		} `json:"deployment_branch_policy"`
		Rules []struct {
			Type      string            `json:"type"`
			Reviewers []json.RawMessage `json:"reviewers"`
		} `json:"protection_rules"`
	}
	if e = json.Unmarshal(b, &environment); e != nil {
		return e
	}
	reviewer := false
	for _, rule := range environment.Rules {
		if rule.Type == "required_reviewers" && len(rule.Reviewers) > 0 {
			reviewer = true
		}
	}
	if !reviewer || environment.CanAdminsBypass == nil || *environment.CanAdminsBypass || !environment.DeploymentBranchPolicy.Custom || environment.DeploymentBranchPolicy.Protected {
		return fmt.Errorf("fork-release must require reviewers, disable admin bypass, and use main-only custom branch policy")
	}
	b, e = c.api(base + "/deployment-branch-policies")
	if e != nil {
		return e
	}
	var policies struct {
		Items []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"branch_policies"`
	}
	if e = json.Unmarshal(b, &policies); e != nil {
		return e
	}
	if len(policies.Items) != 1 || policies.Items[0].Name != "main" || policies.Items[0].Type != "branch" {
		return fmt.Errorf("fork-release must allow only the main branch")
	}
	return nil
}
