module openai

go 1.22

require (
	github.com/bsm/redislock v0.9.4
	github.com/disintegration/imaging v1.6.2
	github.com/felixge/httpsnoop v1.0.4
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a
	github.com/go-pay/gopay v1.5.100
	github.com/go-pay/util v0.0.2
	github.com/gomarkdown/markdown v0.0.0-20240328165702-4d01890c35c0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.5.1
	github.com/robfig/cron v1.2.0
	github.com/sashabaranov/go-openai v1.20.2
	github.com/silenceper/wechat/v2 v2.1.6
	github.com/sirupsen/logrus v1.9.3
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
	github.com/tiktoken-go/tokenizer v0.1.0
	golang.org/x/sync v0.6.0
)

require (
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-pay/bm v0.0.1 // indirect
	github.com/go-pay/crypto v0.0.1 // indirect
	github.com/go-pay/errgroup v0.0.2 // indirect
	github.com/go-pay/xhttp v0.0.2 // indirect
	github.com/go-pay/xlog v0.0.2 // indirect
	github.com/go-pay/xtime v0.0.2 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	github.com/tidwall/gjson v1.17.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/image v0.15.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

replace github.com/silenceper/wechat/v2 => github.com/guoqingxibei/wechat/v2 v2.2.7
