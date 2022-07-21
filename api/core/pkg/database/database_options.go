package database

type Options struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
	SSLMode  bool   `json:"ssl_mode" yaml:"sslMode"`
	Timezone string `json:"timezone" yaml:"timezone"`
}
