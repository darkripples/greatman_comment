package character

import "testing"

func TestSanitizeReply_removesBoundaryBlock(t *testing.T) {
	in := "正文内容。\n\n---\n\n**附：边界说明**\n依据：…"
	out := SanitizeReply(in)
	if out != "正文内容。" {
		t.Fatalf("got %q", out)
	}
}

func TestSanitizeReply_preservesNormalContent(t *testing.T) {
	in := "这是正常回复，不含边界说明段落。"
	if SanitizeReply(in) != in {
		t.Fatal("content changed unexpectedly")
	}
}
