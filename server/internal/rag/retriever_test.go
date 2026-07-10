package rag

import "testing"

func TestAugmentSystemPrompt_matchesTags(t *testing.T) {
	r := &Retriever{
		byCharacter: map[string][]Passage{
			"luxun": {
				{Title: "国民性", Text: "看客的材料", Tags: []string{"国民性"}},
				{Title: "无关", Text: "其他话题", Tags: []string{"无关"}},
			},
		},
	}
	out := r.AugmentSystemPrompt("luxun", "国民性如何", "热榜议题", "base")
	if !contains(out, "国民性") || !contains(out, "看客") {
		t.Fatalf("expected matched passage in prompt: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
