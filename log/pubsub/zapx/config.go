package zapx

import (
	"net/url"
	//"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sink zapcore.WriteSyncer
var encoder zapcore.Encoder
var level zapcore.Level

func init() {
	zap.RegisterSink("redis", InitRedisSink)
	zap.RegisterSink("kafka", InitKafkaSink)
	eConfig := zap.NewProductionEncoderConfig()
	eConfig.LineEnding = cachedLogLineEnding
	eConfig.MessageKey = cachedLogMessageKey
	eConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	eConfig.SkipLineEnding=true
	//eConfig := zap.NewProductionConfig()
	sink, _, _ = zap.Open(cachedLogSinkURL)
	encoder = zapcore.NewJSONEncoder(eConfig)
}

const (
	redisPubSubType      = "channel"
	redisDefaultType     = "list"
	redisClusterNodesKey = "nodes"
	redisDefaultPwd      = ""
	redisDefaultKey      = "just_a_test_key"
	kafkaDefaultTopic    = "just_a_test_topic"
	kafkaAsyncKey        = "isAsync"

	cachedLogLineEnding = ","
	cachedLogMessageKey = ""
	cachedLogSinkURL    = "stderr"
	globalKeyPrefix     = "__"
)

var (
	stdErrSink, _, _ = zap.Open("stderr")
)

// SinkURL sink url
type SinkURL struct {
	url.URL
}

// CachedLogConfig cached log config
type CachedLogConfig struct {
	sinkURLs         []SinkURL
	level            zapcore.Level
	econfig          zapcore.EncoderConfig
	isEConfigChanged bool
}

// Level set level
func (cl CachedLogConfig) Level(level zapcore.Level) CachedLogConfig {
	cl.level = level
	return cl
}

// AddSinks add sinks
func (cl CachedLogConfig) AddSinks(url ...SinkURL) CachedLogConfig {
	cl.sinkURLs = append(cl.sinkURLs, url...)
	return cl
}

// EncoderConfig set encoder config
func (cl CachedLogConfig) EncoderConfig(econfig zapcore.EncoderConfig) CachedLogConfig {
	cl.econfig = econfig
	cl.isEConfigChanged = true
	return cl
}

// Build create logger
func (cl CachedLogConfig) Build(env int) (*CachedLogger, error) {
	ws := stdErrSink
	core := getCachedCore()
	core.LevelEnabler = level
	urls := make([]string, 0, len(cl.sinkURLs)+1)
	urls = append(urls, cachedLogSinkURL)
	if len(cl.sinkURLs) > 0 {
		urls = urls[:0]
		for _, url := range cl.sinkURLs {
			urls = append(urls, url.String())
		}
		var err error
		ws, _, err = zap.Open(urls...)
		if err != nil {
			return nil, err
		}
	}
	core.out = ws
	if cl.isEConfigChanged {
		core.enc = zapcore.NewJSONEncoder(cl.econfig)
	}
	switch env {
	case 0://prd
		logger := zap.New(core,zap.AddCaller(),zap.AddCallerSkip(1))
		return &CachedLogger{Logger: logger}, nil
	default://local | stg
		logger := zap.New(core,zap.AddCaller(),zap.AddCallerSkip(1))
		return &CachedLogger{Logger: logger}, nil
	}
}

// NewCachedLoggerConfig create logger config
func NewCachedLoggerConfig() CachedLogConfig {
	return CachedLogConfig{}
}
