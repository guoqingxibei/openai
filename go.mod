module openai

go 1.24

require (
	github.com/bsm/redislock v0.9.4
	github.com/disintegration/imaging v1.6.2
	github.com/felixge/httpsnoop v1.0.4
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a
	github.com/go-pay/gopay v1.5.109
	github.com/go-pay/util v0.0.4
	github.com/gomarkdown/markdown v0.0.0-20250207164621-7a1f277a159e
	github.com/google/uuid v1.6.0
	github.com/pkoukk/tiktoken-go v0.1.7
	github.com/redis/go-redis/v9 v9.7.1
	github.com/robfig/cron v1.2.0
	github.com/sashabaranov/go-openai v1.38.0
	github.com/silenceper/wechat/v2 v2.1.7
	github.com/sirupsen/logrus v1.9.3
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
	golang.org/x/sync v0.11.0
)

require (
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-pay/crypto v0.0.1 // indirect
	github.com/go-pay/errgroup v0.0.3 // indirect
	github.com/go-pay/smap v0.0.2 // indirect
	github.com/go-pay/xlog v0.0.3 // indirect
	github.com/go-pay/xtime v0.0.2 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/image v0.24.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace github.com/silenceper/wechat/v2 => github.com/guoqingxibei/wechat/v2 v2.2.7

// TODO: remove this line util ReasoningContent is ready in the formmer
replace github.com/sashabaranov/go-openai => github.com/goodenough227/go-openai v0.0.0-20250207091202-52738616d5a4
