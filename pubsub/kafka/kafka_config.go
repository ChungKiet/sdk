package kafka
import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"github.com/Shopify/sarama"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"
	"github.com/goonma/sdk/config/vault"
	//"github.com/xdg-go/scram"
	"fmt"
	"time"
	"github.com/google/uuid"
)

func NewProducerConfig(storeCfg map[string]string) *sarama.Config{
	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 3
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Return.Successes = true
	conf.Producer.Idempotent = true
	conf.Net.MaxOpenRequests = 1
	conf.Metadata.Full = true
	conf.Version = sarama.V2_1_0_0
	conf.ClientID = "Goonma-SDK-1.0.1-Producer"
	conf.Metadata.Full = true
	if storeCfg["USERNAME"]!="" && storeCfg["PASSWORD"]!=""{
		conf.Net.SASL.Enable = true
		conf.Net.SASL.User = storeCfg["USERNAME"]
		conf.Net.SASL.Password =storeCfg["PASSWORD"]
		conf.Net.SASL.Handshake = true
		/*if strings.ToLower(storeCfg["SHA_ALGORITHM"]) == "sha512" {
			conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		} else {
			conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	
		}*/
	}
	if strings.ToLower(storeCfg["USE_SSL"])=="yes" || strings.ToLower(storeCfg["USE_SSL"]) =="true"{
		conf.Net.TLS.Enable = true
		conf.Net.TLS.Config = createTLSConfiguration(storeCfg)
	}
	return conf
}

func NewConsumerConfig(storeCfg map[string]string) *sarama.Config{
	conf := sarama.NewConfig()
	conf.Metadata.Full = true
	//conf.Version = sarama.V0_10_0_0
	conf.ClientID = "Goonma-SDK-1.0.1-Consumer"+"-"+uuid.New().String()
	conf.Metadata.Full = true
	conf.Consumer.Group.Session.Timeout = 30 * time.Second
	conf.Consumer.MaxProcessingTime = 60*60*time.Second
	conf.Consumer.Group.Rebalance.Timeout =  60*time.Second
	conf.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	if storeCfg["USERNAME"]!="" && storeCfg["PASSWORD"]!=""{
		conf.Net.SASL.Enable = true
		conf.Net.SASL.User = storeCfg["USERNAME"]
		conf.Net.SASL.Password =storeCfg["PASSWORD"]
		conf.Net.SASL.Handshake = true
		/*if strings.ToLower(storeCfg["SHA_ALGORITHM"]) == "sha512" {
			conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		} else {
			conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	
		}*/
	}
	if strings.ToLower(storeCfg["USE_SSL"])=="yes" || strings.ToLower(storeCfg["USE_SSL"]) =="true"{
		conf.Net.TLS.Enable = true
		conf.Net.TLS.Config = createTLSConfiguration(storeCfg)
	}
	return conf
}
func createTLSConfiguration(storeCfg map[string]string) (t *tls.Config) {
	verify_ssl:=false
	if strings.ToLower(storeCfg["VERIFY_SSL"])=="true" || strings.ToLower(storeCfg["VERIFY_SSL"])=="yes"{
		verify_ssl=true
	} 
	t = &tls.Config{
		InsecureSkipVerify: verify_ssl,
	}
	if storeCfg["SSL_CERT"]!="" && storeCfg["SSL_KEY"]!="" && storeCfg["SSL_CA"]!= "" {
		cert, err := tls.X509KeyPair([]byte(storeCfg["CERT"]),[]byte(storeCfg["KEY"]))
		if err != nil {
			log.Error(err.Error(),"KAFKA_SSL_CONFIG")
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(storeCfg["SSL_CA"]))
		verify_ssl:=false
		if storeCfg["VERIFY_SSL"]!=""{
			if strings.ToLower(storeCfg["VERIFY_SSL"])=="yes" || strings.ToLower(storeCfg["VERIFY_SSL"])=="true"{
				verify_ssl=true
			}
		}
		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: verify_ssl,
		}
	}
	return t
}
func GetConfig(vault *vault.Vault,config_path string) map[string]string{
	config_path=strings.TrimSuffix(config_path,"/")
	m:=utils.DictionaryString()
	m["BROKERS"]=vault.ReadVAR(fmt.Sprintf("%s/BROKERS",config_path))
	m["TOPIC"]=vault.ReadVAR(fmt.Sprintf("%s/TOPIC",config_path))
	m["USERNAME"]=vault.ReadVAR(fmt.Sprintf("%s/USERNAME",config_path))
	m["PASSWORD"]=vault.ReadVAR(fmt.Sprintf("%s/PASSWORD",config_path))
	m["USE_SSL"]=vault.ReadVAR(fmt.Sprintf("%s/USE_SSL",config_path))
	m["SSL_CA"]=vault.ReadVAR(fmt.Sprintf("%s/SSL_CA",config_path))
	m["SSL_CERT"]=vault.ReadVAR(fmt.Sprintf("%s/SSL_CERT",config_path))
	m["SSL_KEY"]=vault.ReadVAR(fmt.Sprintf("%s/SSL_KEY",config_path))
	m["VERIFY_SSL"]=vault.ReadVAR(fmt.Sprintf("%s/VERIFY_SSL",config_path))
	m["NUM_CONSUMER"]=vault.ReadVAR(fmt.Sprintf("%s/NUM_CONSUMER",config_path))
	m["CONSUMER_GROUP"]=vault.ReadVAR(fmt.Sprintf("%s/CONSUMER_GROUP",config_path))
	m["NUM_POD"]=vault.ReadVAR(fmt.Sprintf("%s/NUM_POD",config_path))
	return m
}

func MergeConfig(global,local map[string]string) map[string]string{
	m:=utils.DictionaryString()
	if utils.Map_contains(global,"BROKERS") || utils.Map_contains(local,"BROKERS"){
		if utils.Map_contains(local,"BROKERS") && local["BROKERS"]!=""{
			m["BROKERS"]=local["BROKERS"]
		}else{
			m["BROKERS"]=global["BROKERS"]
		}
	}
	if utils.Map_contains(global,"TOPIC") || utils.Map_contains(local,"TOPIC"){
		if utils.Map_contains(local,"TOPIC") &&  local["TOPIC"]!=""{
			m["TOPIC"]=local["TOPIC"]
		}else{
			m["TOPIC"]=global["TOPIC"]
		}
	}
	if utils.Map_contains(global,"NUM_CONSUMER") || utils.Map_contains(local,"NUM_CONSUMER"){
		if utils.Map_contains(local,"NUM_CONSUMER") && local["NUM_CONSUMER"]!=""{
			m["NUM_CONSUMER"]=local["NUM_CONSUMER"]
		}else{
			m["NUM_CONSUMER"]=global["NUM_CONSUMER"]
		}
	}
	if utils.Map_contains(global,"USERNAME") || utils.Map_contains(local,"USERNAME"){
		if utils.Map_contains(local,"USERNAME") && local["USERNAME"]!=""{
			m["USERNAME"]=local["USERNAME"]
		}else{
			m["USERNAME"]=global["USERNAME"]
		}
	}
	if utils.Map_contains(global,"PASSWORD") || utils.Map_contains(local,"PASSWORD"){
		if utils.Map_contains(local,"PASSWORD") && local["PASSWORD"]!=""{
			m["PASSWORD"]=local["PASSWORD"]
		}else{
			m["PASSWORD"]=global["PASSWORD"]
		}
	}
	if utils.Map_contains(global,"USE_SSL") || utils.Map_contains(local,"USE_SSL"){
		if utils.Map_contains(local,"USE_SSL") && local["USE_SSL"]!=""{
			m["USE_SSL"]=local["USE_SSL"]
		}else{
			m["USE_SSL"]=global["USE_SSL"]
		}
	}
	if utils.Map_contains(global,"CONSUMER_GROUP") || utils.Map_contains(local,"CONSUMER_GROUP"){
		if utils.Map_contains(local,"CONSUMER_GROUP") && local["CONSUMER_GROUP"]!=""{
			m["CONSUMER_GROUP"]=local["CONSUMER_GROUP"]
		}else{
			m["CONSUMER_GROUP"]=global["CONSUMER_GROUP"]
		}
	}
	if utils.Map_contains(global,"SSL_CA") || utils.Map_contains(local,"SSL_CA"){
		if utils.Map_contains(local,"SSL_CA") && local["SSL_CA"]!=""{
			m["SSL_CA"]=local["SSL_CA"]
		}else{
			m["SSL_CA"]=global["SSL_CA"]
		}
	}
	if utils.Map_contains(global,"SSL_CERT") || utils.Map_contains(local,"SSL_CERT"){
		if utils.Map_contains(local,"SSL_CERT") && local["SSL_CERT"]!=""{
			m["SSL_CERT"]=local["SSL_CERT"]
		}else{
			m["SSL_CERT"]=global["SSL_CERT"]
		}
	}
	if utils.Map_contains(global,"SSL_KEY") || utils.Map_contains(local,"SSL_KEY"){
		if utils.Map_contains(local,"SSL_KEY") && local["SSL_KEY"]!=""{
			m["SSL_KEY"]=local["SSL_KEY"]
		}else{
			m["SSL_KEY"]=global["SSL_KEY"]
		}
	}
	if utils.Map_contains(global,"VERIFY_SSL") || utils.Map_contains(local,"VERIFY_SSL"){
		if utils.Map_contains(local,"VERIFY_SSL") && local["VERIFY_SSL"]!=""{
			m["VERIFY_SSL"]=local["VERIFY_SSL"]
		}else{
			m["VERIFY_SSL"]=global["VERIFY_SSL"]
		}
	}
	if utils.Map_contains(global,"NUM_POD") || utils.Map_contains(local,"NUM_POD"){
		if utils.Map_contains(local,"NUM_POD") && local["NUM_POD"]!=""{
			m["NUM_POD"]=local["NUM_POD"]
		}else{
			m["NUM_POD"]=global["NUM_POD"]
		}
	}
	return m
}