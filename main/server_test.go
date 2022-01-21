package main

import (
	"testing"
	admission "k8s.io/api/admission/v1"
	k8meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ruleShouldNotApply(t *testing.T) {
	request_empty := admission.AdmissionRequest{}
	request_default_pod := admission.AdmissionRequest{
		Namespace: "default",
		Kind: k8meta.GroupVersionKind{Kind: "Pod"},
	}
	rule_empty := Rule{}
	rule_default_pod := Rule{Namespace: "kube-system", Kinds: []string{"Pod", "Node"}}

	if doesRuleApply(&rule_empty, &request_empty) {
		t.Errorf("rule_empty should not apply on request_empty")
	}
	if doesRuleApply(&rule_empty, &request_default_pod) {
		t.Errorf("rule_empty should not apply on request_default_pod")
	}
	if doesRuleApply(&rule_default_pod, &request_default_pod) {
		t.Errorf("rule_default_pod should not apply on request_default_pod")
	}
}

func Test_ruleShouldApply(t *testing.T) {
	request_default_pod := admission.AdmissionRequest{
		Namespace: "default",
		Kind: k8meta.GroupVersionKind{Kind: "Pod"},
	}
	rule_default_pod := Rule{Namespace: "default", Kinds: []string{"Pod", "Node"}}
	rule_wild_ns_pod := Rule{Namespace: "*", Kinds: []string{"Pod", "Node"}}
	rule_wild_ns_wild := Rule{Namespace: "*", Kinds: []string{"*"}}

	if !doesRuleApply(&rule_default_pod, &request_default_pod) {
		t.Errorf("rule_default_pod should apply on request_default_pod")
	}
	if !doesRuleApply(&rule_wild_ns_pod, &request_default_pod) {
		t.Errorf("rule_default_pod should apply on rule_wild_ns_pod")
	}
	if !doesRuleApply(&rule_wild_ns_wild, &request_default_pod) {
		t.Errorf("rule_default_pod should apply on rule_wild_ns_wild")
	}
}

