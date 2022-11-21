package metric

import (
	"sync"
	"os"
	"net/url"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	//kafka "github.com/segmentio/kafka-go"
	"github.com/goonma/sdk/log/pubsub/zapx"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	dlog "log"
	"fmt"
	"encoding/json"
)


type Logger struct {
	id string
	host string
	service     string
	logEngineer *zap.Logger
	logCacheEngineer *zapx.CachedLogger
	mode        int
	dest        int
	env         string 
}
var log Logger
var wg sync.WaitGroup
//var logger *zap.Logger

/*
Log mode:
	- 0: production
	- 1: staging
	- 2: local
	- 3: dev
Log destination:
	-0: stdout
	-1: file
	-2: pub/sub
args: pub/sub kafka
	args[0]:brokers list, ex: 10.148.0.177:9092,10.148.15.194:9092,10.148.15.198:9092
	args[1]:topic
	args[2]:username
	args[3]:password

*/
func Initial(service_name string) {
	//
	//get ENV
	err_env := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err_env!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			panic(err)
		}
	}
	env:=os.Getenv("ENV")//2: local,1: development,0:product
	mode:=2
	dest:=0
	if env=="" || env=="local"{//local
		mode=2
		dest=0
	}else{
		if env=="prd"{
			mode=0
			dest=0
		}else if env=="dev"{
			mode=3
			dest=0
		}else{
			mode=1
			dest=0
		}
	}
	//
	log.service = service_name
	hostname, _ := os.Hostname()
	log.host=hostname
	var err error
	log.mode = mode
	log.dest = dest
	log.env  = env
	var cfg zap.Config
	cfg = zap.NewProductionConfig()	
	// cfg.OutputPaths = []string{"judger.log"}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	if mode==2{
		cfg.Encoding="console"
	}else{
		cfg.Encoding="json"
	}
	log.logEngineer, err = cfg.Build()
	defer log.logEngineer.Sync()
	if err != nil {
		panic(err)
	}
	//
}
func SetDestKafka(config_map map[string]string){
	if config_map["BROKERS"]==""{
		panic("Logger SetDestKafka: LOGGER KAFKA BROKER not found")
	}
	if config_map["TOPIC"]==""{
		panic("Logger SetDestKafka: LOGGER KAFKA TOPIC not found")
	}
	brokers:= config_map["BROKERS"]
	topic:=config_map["TOPIC"]
	var err error
	log.dest=2
	stderr := zapx.SinkURL{url.URL{Opaque: "stderr"}}
	sinkUrl := zapx.SinkURL{url.URL{Scheme: "kafka", Host: brokers, RawQuery: fmt.Sprintf("topic=%s&username=%s&password=%s",topic,config_map["USERNAME"],config_map["PASSWORD"])}}
	log.logCacheEngineer,err= zapx.NewCachedLoggerConfig().AddSinks(stderr,sinkUrl).Build(log.mode)
	if err!=nil{
		dlog.Fatalf(err.Error())
	}
	defer log.logCacheEngineer.Flush(&wg)
	dlog.Println("Logger connected to Kafka success")
}
func Push(metric_name string,t int) {
	if log.dest==2{
		log.logCacheEngineer.Info(
			"",
			zap.String("id",uuid.New().String()),
			zap.String("host", log.host),
			zap.String("service", log.service),
			zap.String("metric", metric_name),
			zap.Int("t", t),
			zap.String("env", log.env),
		)
		defer log.logCacheEngineer.Flush(&wg)
	}else{
		dlog.Println(metric_name,t)
		/*log.logEngineer.Info(
			"",
			zap.String("id",uuid.New().String()),
			zap.String("msg",msg),
			zap.String("host", log.host),
			zap.String("service", log.service),
			zap.String("key_msg", key_msg),
		)
		defer log.logEngineer.Sync()*/
	}
	
}

/*
- 0: production
- 1: deveopment
- 2: local
*/
func LogMode() int{
	return log.mode
}

func ItoString(value interface{}) string {
	if value==nil{
		return ""
	}
	str := fmt.Sprintf("%v", value)
	return str
}
func StructToJson(v interface{}) string{
	if v==nil{
		return ""
	}
	out, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return (string(out))
}