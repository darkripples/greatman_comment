package eval

import "testing"

func TestInspectFlagsEraBoundaryAndMissingCitation(t *testing.T) {
	findings := inspect("era", "我知道 2020 年的所有事情", nil)
	if len(findings) != 1 || findings[0].Severity != "severe" {
		t.Fatalf("unexpected findings: %+v", findings)
	}
	findings = inspect("evidence", "这是我的看法", nil)
	if len(findings) != 1 || findings[0].Rule != "missing-citation" {
		t.Fatalf("unexpected findings: %+v", findings)
	}
}

func TestInspectAcceptsBoundaryStatement(t *testing.T) {
	if findings := inspect("era", "此事我不可确知，只能从旧日经验推想。", nil); len(findings) != 0 {
		t.Fatalf("unexpected findings: %+v", findings)
	}
}
