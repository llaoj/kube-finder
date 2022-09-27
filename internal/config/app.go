package config

type App struct {
	Name          string            `yaml:"name"`
	LogLevel      string            `yaml:"logLevel"`
	HttpPort      string            `yaml:"httpPort"`
	Kubeconfig    string            `yaml:"kubeconfig"`
	CgroupPIDRoot string            `yaml:"cgroupPIDRoot"`
	Clients       map[string]string `yaml:"clients"`
}
