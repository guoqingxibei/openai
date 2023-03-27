package wechat

import (
	"encoding/xml"
	"time"
)

type Msg struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Event        string   `xml:"Event"`
	Content      string   `xml:"Content"`
	Image        Image    `xml:"Image"`
	MsgId        int64    `xml:"MsgId,omitempty"`
	EventKey     string   `xml:"EventKey"`
}

type Image struct {
	MediaId string `xml:"MediaId"`
}

func NewInMsg(data []byte) *Msg {
	var msg Msg
	if err := xml.Unmarshal(data, &msg); err != nil {
		return nil
	}
	return &msg
}

func (msg *Msg) BuildTextMsg(reply string) []byte {
	data := Msg{
		ToUserName:   msg.FromUserName,
		FromUserName: msg.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      reply,
	}
	bs, _ := xml.Marshal(&data)
	return bs
}

func (msg *Msg) BuildImageMsg(mediaId string) []byte {
	data := Msg{
		ToUserName:   msg.FromUserName,
		FromUserName: msg.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "image",
		Image: Image{
			MediaId: mediaId,
		},
	}
	bs, _ := xml.Marshal(&data)
	return bs
}
