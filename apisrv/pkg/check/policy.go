package check

import (
	"context"

	"github.com/immune-gmbh/guard/apisrv/v2/pkg/check/policy"
	"github.com/immune-gmbh/guard/apisrv/v2/pkg/issuesv1"
)

type policyEndpointProtection struct{}

func (policyEndpointProtection) String() string {
	return "Endpoint protection policy"
}

func (policyEndpointProtection) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if !hasEndpointProtection(ctx, subj) && subj.Policy.EndpointProtection == policy.True {
		var iss issuesv1.PolicyEndpointProtection
		iss.Common.Id = issuesv1.PolicyEndpointProtectionId
		iss.Common.Aspect = issuesv1.PolicyEndpointProtectionAspect
		iss.Common.Incident = true
		return &iss
	}

	return nil
}

func (policyEndpointProtection) Update(ctx context.Context, overrides []string, subj *Subject) {
}

type policyIntelTSC struct{}

func (policyIntelTSC) String() string {
	return "Endpoint protection policy"
}

func (policyIntelTSC) Verify(ctx context.Context, subj *Subject) issuesv1.Issue {
	if !hasIntelTSC(ctx, subj) && subj.Policy.IntelTSC == policy.True {
		var iss issuesv1.PolicyIntelTsc
		iss.Common.Id = issuesv1.PolicyIntelTscId
		iss.Common.Aspect = issuesv1.PolicyIntelTscAspect
		iss.Common.Incident = true
		return &iss
	}

	return nil
}

func (policyIntelTSC) Update(ctx context.Context, overrides []string, subj *Subject) {
}
