package config


type Config struct {
	Server string `json:"server"`
	TarPath  string `json:"tarPath"`
	WebPath string `json:"webPath"`
	ApiKey string	`json:"apiKey"`
	WebSite string `json:"webSite"`
}