package util

import "testing"

func TestIsEnglishSentence(t *testing.T) {
	sentence1 := "Hi, how are you?"
	if !IsEnglishSentence(sentence1) {
		t.Error("Wrong assessment for English sentence")
	}

	sentence2 := "你好，中国"
	if IsEnglishSentence(sentence2) {
		t.Error("Wrong assessment for Chinese sentence")
	}

	sentence3 := "你好，Jack"
	if IsEnglishSentence(sentence3) {
		t.Error("Wrong assessment for mixed sentence")
	}

	sentence4 := "Hi, how are you？"
	if !IsEnglishSentence(sentence4) {
		t.Error("Wrong assessment for English sentence with Chinese punctuation")
	}
}

func TestTruncateString(t *testing.T) {
	str1 := "Hi, how are you? Is everything okay?"
	len1 := len([]rune(str1))
	if TruncateString(str1, len1) != str1 {
		t.Error()
	}
	if TruncateString(str1, len1+1) != str1 {
		t.Error()
	}
	if TruncateString(str1, 4) != "Hi...y?" {
		t.Error()
	}

	str2 := "你好，今天碰到什么开心的事情了吗"
	len2 := len([]rune(str2))
	if TruncateString(str2, len2) != str2 {
		t.Error()
	}
	if TruncateString(str2, len2+1) != str2 {
		t.Error()
	}
	if TruncateString(str2, 4) != "你好...了吗" {
		t.Error()
	}
}
