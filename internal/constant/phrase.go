package constant

const (
	ChatSystemMessage = "You are a large language model trained by OpenAI." +
		" Your job is to generate human-like text based on the input you received, allowing it to engage in natural-sounding conversations and provide responses that are coherent and relevant to the topic at hand." +
		" If the input is a question, try your best to answer it. Otherwise, provide as much information as you can." +
		" Knowledge cutoff: {knowledge_cutoff} Current date: {current_date}"
	TryAgain         = "哎呀，出错啦，重试一下嘛~"
	CensorWarning    = "【警告】我是公众号作者，检测到你的发言可能涉嫌违规。如果你继续违规使用，公众号将拒绝为你提供服务。"
	ChatUsage        = "你问 ChatGPT 答，不限次数。"
	ImageUsage       = "你说一句描述，ChatGPT 画一张图，每天仅限 %d 次（成本昂贵，敬请谅解）。"
	UsageTail        = "回复 help，可查看详细用法。\n回复 donate，可捐赠作者。"
	DonateDesc       = "回复 donate，可对作者进行捐赠。所有对话都会产生费用，你的捐赠可以减轻作者的经济压力，以维持服务更好、更久的运行下去。"
	SubscribeReply   = "我是【程序员uncle】，会定期分享一些关于 ChatGPT 的知识。有任何问题，可微信：programmer_guy。\n\n此公众号已经接入 ChatGPT，回复 help，查看详细用法。"
	ZeroImageBalance = "你今天的 image 次数已用完，24 小时后再来吧。回复 chat，可切换到不限次数的 chat 模式。"
)
