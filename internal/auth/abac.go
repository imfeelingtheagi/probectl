package auth

// ABAC over RBAC (S31, F25). The two-level boundary already resolves the tenant
// then RBAC; ABAC is a THIRD check layered on top: tenant-scoped attribute
// policies that can DENY a permission an RBAC role grants, based on the subject's
// attributes (e.g. department, mfa) and the resource's attributes (e.g. org/team/
// project — delegated admin). The model is intentionally deny-override: RBAC is
// the baseline grant, and ABAC narrows it. An allow policy is a silent permit
// (RBAC already permitted), so ABAC never widens access beyond RBAC.

// PolicyEffect is a policy's decision.
type PolicyEffect string

const (
	PolicyAllow PolicyEffect = "allow"
	PolicyDeny  PolicyEffect = "deny"
)

// Policy is one tenant-scoped attribute policy. It applies to a permission (or "*"
// for any) and matches when EVERY listed subject attribute and EVERY listed
// resource attribute equals the request's value. Among matching policies the
// highest Priority decides; a deny wins ties (deny-override).
type Policy struct {
	ID         string            `json:"id,omitempty"`
	Name       string            `json:"name"`
	Effect     PolicyEffect      `json:"effect"`
	Permission string            `json:"permission"` // "*" = any
	Subject    map[string]string `json:"subject,omitempty"`
	Resource   map[string]string `json:"resource,omitempty"`
	Priority   int               `json:"priority"`
	Enabled    bool              `json:"enabled"`
}

// Evaluate returns the ABAC decision for a permission given the subject/resource
// attributes, or "" when no policy applies (ABAC is silent — RBAC governs).
func Evaluate(policies []Policy, permission string, subject, resource map[string]string) PolicyEffect {
	decided := PolicyEffect("")
	bestPriority := 0
	for i := range policies {
		p := policies[i]
		if !p.Enabled {
			continue
		}
		if p.Permission != "*" && p.Permission != permission {
			continue
		}
		if !attrsSubset(p.Subject, subject) || !attrsSubset(p.Resource, resource) {
			continue
		}
		switch {
		case decided == "" || p.Priority > bestPriority:
			decided, bestPriority = p.Effect, p.Priority
		case p.Priority == bestPriority && p.Effect == PolicyDeny:
			decided = PolicyDeny // deny-override on ties
		}
	}
	return decided
}

// Permit is the full S31 access decision for a permission: RBAC must grant it AND
// ABAC must not deny it. resource may be nil for routes that carry no resource
// attributes (then only subject-attribute policies apply).
func Permit(p *Principal, permission string, policies []Policy, resource map[string]string) bool {
	if !p.Has(permission) {
		return false // RBAC baseline
	}
	return Evaluate(policies, permission, attrsOf(p), resource) != PolicyDeny
}

// attrsSubset reports whether every key/value in required is present and equal in
// actual. An empty requirement matches anything.
func attrsSubset(required, actual map[string]string) bool {
	for k, v := range required {
		if actual[k] != v {
			return false
		}
	}
	return true
}

func attrsOf(p *Principal) map[string]string {
	if p == nil {
		return nil
	}
	return p.Attributes
}
