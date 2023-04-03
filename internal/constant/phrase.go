package constant

const (
	ChatSystemMessage = "You are a large language model trained by OpenAI." +
		" Your job is to generate human-like text based on the input you received, allowing it to engage in natural-sounding conversations and provide responses that are coherent and relevant to the topic at hand." +
		" If the input is a question, try your best to answer it. Otherwise, provide as much information as you can."
	TryAgain         = "哎呀，出错啦，请重试。"
	CensorWarning    = "不要在这里讨论政治哦，谢谢配合[爱心]"
	ChatUsage        = "你问 ChatGPT 答，不限次数。"
	ImageUsage       = "你说一句描述，ChatGPT 画一张图，每天仅限 %d 次（价格昂贵，敬请谅解）。"
	ContactDesc      = "回复 contact，可联系作者。"
	DonateReminder   = "如果你觉得体验不错的话，可回复 donate 对作者进行捐赠哦。每一次对话都会产生不少的费用，你的捐赠可以减轻作者的经济压力，以维持服务更好、更久的运行下去。"
	ContactInfo      = "微信：programmer_guy\n邮箱：guoqingxibei@gmail.com"
	SubscribeReply   = "此公众号已接入 ChatGPT。\n\n回复 help，可查看详细用法。\n回复 donate，可捐赠作者。"
	ZeroImageBalance = "哎呀，今天的 image 次数用完啦，只能 24 小时后再来了。回复 chat，可切换到不限次数的 chat 模式哦。"
)
