package logic

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/constant"
	"openai/internal/model"
	"openai/internal/store"
	"openai/internal/util"
)

func Transfer(msg *message.MixMessage) (reply *message.Reply) {
	if !util.AccountIsUncle() {
		return util.BuildTextReply("此公众号不支持迁移，请移步公众号「程序员uncle」进行迁移。")
	}

	userName := string(msg.FromUserName)
	paidBalance, _ := store.GetPaidBalance(userName)
	replyText := "你的付费额度剩余0次，无需迁移。"
	if paidBalance > 0 {
		_ = store.SetPaidBalance(userName, 0)
		code := uuid.New().String()
		codeDetail := model.CodeDetail{
			Code:   code,
			Times:  paidBalance,
			Status: constant.Created,
		}
		codeDetailBytes, _ := json.Marshal(codeDetail)
		_ = store.SetCodeDetail(code, string(codeDetailBytes), true)
		replyText = fmt.Sprintf("你的付费额度剩余%d次，已在此公众号下清零。请复制下面的code发送给新公众号「程序员brother」，"+
			"即可完成迁移。感谢你的一路陪伴❤️\n\n%s", paidBalance, code)
	}
	return util.BuildTextReply(replyText)
}
