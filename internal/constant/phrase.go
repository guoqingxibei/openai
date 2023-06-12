package constant

const (
	ChatSystemMessage = "你是由OpenAI训练的大型语言模型。你的工作是根据你接收到的输入生成类似人类语言的文本，使其能够参与自然的对话并提供连贯和相关的回答。如果输入是一个问题，尽你所能回答它。否则，提供尽可能多的信息。"
	TryAgain          = "哎呀，出错啦（大概率是请求量过大导致的ChatGPT负载过高），请重试~\n\n回复contact，可报告bug。"
	ExpireError       = "哎呀，消息过期了，重新提问吧~"
	CensorWarning     = "【温馨提醒】很抱歉识别出了可能不宜讨论的内容。如有误判，可回复contact联系作者，作者将继续进行优化调整。\n\n为了公众号能持续向大家提供服务，请大家不要在这里讨论色情、政治、暴恐、VPN等相关内容，谢谢配合❤️"
	ChatUsage         = "每天%d次，剩余%d次。"
	ImageUsage        = "你说一句描述，ChatGPT画一张图，每天%d次，剩余%d次。"
	ContactDesc       = "回复contact，可联系作者。"
	DonateDesc        = "回复donate，可捐赠作者。"
	DonateReminder    = "❤️如果你觉得体验不错的话，可回复【donate】对作者进行捐赠哦。每一次对话都会产生不少的费用（一次长对话约0.1元），你的捐赠可以减轻作者的经济压力，以维持服务更好、更久的运行下去️❤️"
	ContactInfo       = "微信：programmer_guy\n邮箱：jia.guoqing@qq.com"
	ReportInfo        = "bug报给jia.guoqing@qq.com，尽可能描述详细噢~"
	SubscribeReply    = "此公众号已接入ChatGPT 3.5，直接用文字或者语音向我提问吧~\n\n回复contact，可联系作者。\n回复donate，可捐赠作者。"
	ZeroImageBalance  = "很抱歉，今天的画图次数（每天1次）用完了，明天再来吧。费用昂贵，敬请谅解❤️\n\n回复chat，可切换到聊天模式。"
	ZeroChatBalance   = "很抱歉，今天的对话次数（每天20次）用完了，明天再来吧。费用昂贵（一次长对话约0.1元），敬请谅解❤️\n\n如果使用量确实很大，可回复contact联系作者。如果此公众号帮助到了你，可回复donate捐赠作者。"
	TooLongQuestion   = "哎呀，输入太长了~"
)
