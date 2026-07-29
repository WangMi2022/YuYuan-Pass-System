package provider

import "testing"

func TestInvoiceTypeName(t *testing.T) {
	if got := invoiceTypeName("10"); got != "增值税电子普通发票" {
		t.Fatalf("unexpected invoice type: %s", got)
	}
	if got := invoiceTypeName("unknown"); got != "电子发票" {
		t.Fatalf("unexpected fallback type: %s", got)
	}
}

func TestMergeMissingPreservesPrimaryResult(t *testing.T) {
	target := Result{Provider: "http", InvoiceNumber: "A", RawText: "ocr"}
	fallback := Result{InvoiceCode: "C", InvoiceNumber: "B", RawText: "qr"}
	mergeMissing(&target, fallback)
	if target.InvoiceCode != "C" || target.InvoiceNumber != "A" || target.RawText != "ocr\nqr" {
		t.Fatalf("unexpected merged result: %#v", target)
	}
}

func TestParseDecimalCentsUsesExactDecimalArithmetic(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{input: "0", want: 0},
		{input: "12.3", want: 1230},
		{input: "12.34", want: 1234},
		{input: " 999999.99 ", want: 99999999},
	}
	for _, test := range tests {
		got, err := parseDecimalCents(test.input)
		if err != nil || got != test.want {
			t.Fatalf("parse %q: got=%d want=%d err=%v", test.input, got, test.want, err)
		}
	}
}

func TestParseDecimalCentsRejectsInvalidOrOverflowingValues(t *testing.T) {
	inputs := []string{"", "-1", "1.234", "NaN", "Inf", "1.", ".5", "92233720368547758.08"}
	for _, input := range inputs {
		if _, err := parseDecimalCents(input); err == nil {
			t.Fatalf("expected %q to be rejected", input)
		}
	}
}
