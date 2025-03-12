package base

var Config config

type config struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	DataSource string `json:"dataSource"`

	Dir string `json:"-"`
}
