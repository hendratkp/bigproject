package nsq

import (
	"log"
	"os"

	nsq "github.com/bitly/go-nsq"
	logging "gopkg.in/tokopedia/logging.v1"
)

type ServerConfig struct {
	Name string
}

type Config struct {
	Server ServerConfig
}

type NSQModule struct {
	cfg *Config
	q   *nsq.Consumer
}

func NewNSQModule() *NSQModule {

	var cfg Config

	ok := logging.ReadModuleConfig(&cfg, "config", "bigproject") || logging.ReadModuleConfig(&cfg, "files/etc/bigproject", "bigproject")
	if !ok {
		// when the app is run with -e switch, this message will automatically be redirected to the log file specified
		log.Fatalln("failed to read config")
	}

	// this message only shows up if app is run with -debug option, so its great for debugging
	logging.Debug.Println("nsq init called", cfg.Server.Name)

	// contohnya: caranya ciptakan nsq consumer
	nsqCfg := nsq.NewConfig()
	q := createNewConsumer(nsqCfg, "bigproject_user", "reloaded", handler)
	q.SetLogger(log.New(os.Stderr, "nsq:", log.Ltime), nsq.LogLevelError)
	q.ConnectToNSQLookupd("127.0.0.1:4161")

	return &NSQModule{
		cfg: &cfg,
		q:   q,
	}

}

func handler(msg *nsq.Message) error {
	log.Println("got message :", string(msg.Body))
	msg.Finish()
	return nil
}

func createNewConsumer(nsqCfg *nsq.Config, topic string, channel string, handler nsq.HandlerFunc) *nsq.Consumer {
	q, err := nsq.NewConsumer(topic, channel, nsqCfg)
	if err != nil {
		log.Fatal("failed to create consumer for ", topic, channel, err)
	}
	q.AddHandler(handler)
	return q
}
