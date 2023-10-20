package util

import (
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"testing"
)

func TestParseXMLtoMsg(t *testing.T) {
	msg1 := &message.MixMessage{}
	str1 := "<xml><ToUserName><![CDATA[gh_03c6a26dad0c]]></ToUserName><FromUserName><![CDATA[oc2Ys6dyYO3CnfnGZG8H5CGzgIBc]]></FromUserName><CreateTime>1686563323</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[基于java的xmlbeans，解析下面的xml内容<message version=\"1.0\"><header action=\"UPDATE\" command=\"CP_INFO_SYNC_REQ\" component-id=\"ICMS\" sequence=\"968785bd565a43249e27542693ad90be\" timestamp=\"20230609125504677\"/><body><memo/><status><![CDATA[5]]></status><cp_name><![CDATA[华谊传媒]]></cp_name><code><![CDATA[CP0032]]></code><id><![CDATA[CP0032]]></id></body></message>]]></Content><MsgId>24145781557537264</MsgId></xml>"
	err1 := ParseXmlToMsg([]byte(str1), msg1)
	if err1 != nil || msg1.Content != "基于java的xmlbeans，解析下面的xml内容<message version=\"1.0\"><header action=\"UPDATE\" command=\"CP_INFO_SYNC_REQ\" component-id=\"ICMS\" sequence=\"968785bd565a43249e27542693ad90be\" timestamp=\"20230609125504677\"/><body><memo/><status><![CDATA[5]]></status><cp_name><![CDATA[华谊传媒]]></cp_name><code><![CDATA[CP0032]]></code><id><![CDATA[CP0032]]></id></body></message>" {
		t.Error("failed to parse nested XML")
	}

	msg2 := &message.MixMessage{}
	str2 := "<xml><ToUserName><![CDATA[gh_03c6a26dad0c]]></ToUserName><FromUserName><![CDATA[oc2Ys6dyYO3CnfnGZG8H5CGzgIBc]]></FromUserName><CreateTime>1686563323</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[Heilongjiang Natural Science Founda\u0002tion (ZD2022E001)帮我查一下这个基金的批准时间，课题名称和经费]]></Content><MsgId>24145781557537264</MsgId></xml>"
	err2 := ParseXmlToMsg([]byte(str2), msg2)
	if err2 != nil || msg2.Content != "Heilongjiang Natural Science Founda\u0002tion (ZD2022E001)帮我查一下这个基金的批准时间，课题名称和经费" {
		t.Error("failed to parse XML with special chars")
	}

	msg3 := &message.MixMessage{}
	str3 := "<xml><ToUserName><![CDATA[gh_f82bcb116a8d]]></ToUserName>\n<FromUserName><![CDATA[owg0669-hFIqe8-gAxmxMid53-AA]]></FromUserName>\n<CreateTime>1697881508</CreateTime>\n<MsgType><![CDATA[event]]></MsgType>\n<Event><![CDATA[unsubscribe]]></Event>\n<EventKey><![CDATA[]]></EventKey>\n</xml>"
	err3 := ParseXmlToMsg([]byte(str3), msg3)
	if err3 != nil || msg3.Content != "" {
		t.Error("failed to parse XML without content")
	}
}
