package utils

type urlStruct struct {
	AuthenticationServiceUrl string
	ConfigServiceUrl         string
	FileConfigServiceUrl     string
	FileServiceUrl           string
	PubSubServiceUrl         string
	RedisServiceUrl          string
	MyIP                     string
}

var URL urlStruct

func InitUrl() {
	URL = urlStruct{
		AuthenticationServiceUrl: "http://authentication-service:3000",
		ConfigServiceUrl:         "http://config-service:3002",
		FileConfigServiceUrl:     "http://192.168.1.3:3000",
		FileServiceUrl:           "http://192.168.1.10:3001",
		PubSubServiceUrl:         "192.168.1.10:3000",
		RedisServiceUrl:          "192.168.1.50:6379",
		MyIP:                     "192.168.1.11",
	}
}
