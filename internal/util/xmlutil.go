package util

import (
	"encoding/xml"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"strings"
)

const (
	contentStartMark  = "<Content>"
	contentEndMark    = "</Content>"
	cdataOpenMarkLen  = 9 // len("<![CDATA["): 9
	cdataCloseMarkLen = 3 // len("]]>"): 3
)

func ParseXmlToMsg(rawXMLMsgBytes []byte, msg *message.MixMessage) (err error) {
	rawXMLMsg := string(rawXMLMsgBytes)
	contentStartMarkIndex := strings.Index(rawXMLMsg, contentStartMark)
	start := contentStartMarkIndex + len(contentStartMark)
	end := strings.LastIndex(rawXMLMsg, contentEndMark)
	contentMarkExist := contentStartMarkIndex != -1 && end != -1
	content := ""
	if contentMarkExist {
		content = rawXMLMsg[start:end]
	}

	xmlWithEmptyContent := strings.Replace(rawXMLMsg, content, "", 1)
	err = xml.Unmarshal([]byte(xmlWithEmptyContent), msg)
	if err != nil {
		return
	}

	if contentMarkExist {
		msg.Content = content[cdataOpenMarkLen : len(content)-cdataCloseMarkLen]
	}
	return
}
