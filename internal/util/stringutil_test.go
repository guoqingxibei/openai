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
