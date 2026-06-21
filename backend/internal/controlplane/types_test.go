package controlplane

import "testing"

func TestDeriveSlug(t *testing.T) {
	cases := map[string]string{
		"Saigon Tours Co.":  "saigon-tours-co",
		"  Acme   Travel  ": "acme-travel",
		"ABC-123":           "abc-123",
		"!!!weird@@@name":   "weird-name",
	}
	for in, want := range cases {
		if got := DeriveSlug(in); got != want {
			t.Errorf("DeriveSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidateSlug(t *testing.T) {
	valid := []string{"acme", "saigon-tours", "a1", "abc-123-xyz"}
	for _, s := range valid {
		if !validateSlug(s) {
			t.Errorf("validateSlug(%q) = false, want true", s)
		}
	}
	invalid := []string{"a", "", "-acme", "acme-", "Acme", "a_b", "a--b"}
	for _, s := range invalid {
		if validateSlug(s) {
			t.Errorf("validateSlug(%q) = true, want false", s)
		}
	}
}

func TestDBNameFromSlug(t *testing.T) {
	if got := dbNameFromSlug("opero_tenant_", "saigon-tours"); got != "opero_tenant_saigon_tours" {
		t.Errorf("dbNameFromSlug = %q", got)
	}
}
