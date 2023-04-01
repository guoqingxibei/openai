package constant

const (
	ChatSystemMessage = "You are a large language model trained by OpenAI." +
		" Your job is to generate human-like text based on the input you received, allowing it to engage in natural-sounding conversations and provide responses that are coherent and relevant to the topic at hand." +
		" If the input is a question, try your best to answer it. Otherwise, provide as much information as you can."
	TryAgain         = "哎呀，出错啦，请重试。"
	CensorWarning    = "请不要在这里讨论政治，谢谢配合。"
	ChatUsage        = "你问 ChatGPT 答，不限次数。"
	ImageUsage       = "你说一句描述，ChatGPT 画一张图，每天仅限 %d 次。"
	UsageTail        = "回复 help，可查看详细用法。\n回复 donate，可捐赠作者。"
	DonateDesc       = "回复 donate，可捐赠作者。"
	ContactDesc      = "回复 contact，可联系作者。"
	ContactInfo      = "微信：programmer_guy\n邮箱：guoqingxibei@gmail.com"
	SubscribeReply   = "此公众号已接入 ChatGPT，回复 help，查看详细用法。"
	ZeroImageBalance = "你今天的 image 次数已用完，24 小时后再来吧。回复 chat，可切换到不限次数的 chat 模式。"
)
