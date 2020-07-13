package fass

import (
	"os"
	"path"
)

type Config struct {
	ListenAddress string
	MailHost string
	MailFrom string
	MailUseAuth bool
	MailAuthIdent string
	MailAuthUser string
	MailAuthPass string
}

func (c Config) path() string {
	return "/home/alex/.config/fass/config.json"
}

func (c Config) Store() error {
	err := os.MkdirAll(path.Dir(c.path()), 0755)
	if err != nil {
		return err
	}

	return marshalToFile(c.path(), c)
}

func DefaultConfig() Config {
	return Config {
		ListenAddress: "localhost:8080",
		MailHost: "localhost",
		MailFrom: "fass@localhost",
		MailUseAuth: false,
		MailAuthIdent: "",
		MailAuthUser: "",
		MailAuthPass: "",
	};
}

func LoadConfig() (config Config, err error) {
	err = unmarshalFromFile(config.path(), &config)
	return
}
