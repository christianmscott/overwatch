package spec

type Config struct {
	Server ServerConfig `yaml:"server,omitempty"`
	Checks []CheckSpec  `yaml:"checks"`
	Alerts AlertsConfig `yaml:"alerts,omitempty"`
	Worker WorkerConfig `yaml:"worker,omitempty"`
}

type ServerConfig struct {
	BindAddress string `yaml:"bind_address,omitempty"`
	BindPort    int    `yaml:"bind_port,omitempty"`
	Concurrency int    `yaml:"concurrency,omitempty"`
}

type AlertsConfig struct {
	Webhooks []WebhookConfig `yaml:"webhooks,omitempty"`
	SMTP     *SMTPConfig     `yaml:"smtp,omitempty"`
}

type WebhookConfig struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Timeout Duration          `yaml:"timeout,omitempty"`
}

type SMTPConfig struct {
	Host       string   `yaml:"host"`
	Port       int      `yaml:"port"`
	Username   string   `yaml:"username,omitempty"`
	Password   string   `yaml:"password,omitempty"`
	From       string   `yaml:"from"`
	Recipients []string `yaml:"recipients"`
	TLS        bool     `yaml:"tls"`
}

type WorkerConfig struct {
	Concurrency  int      `yaml:"concurrency,omitempty"`
	PollInterval Duration `yaml:"poll_interval,omitempty"`
	APIEndpoint  string   `yaml:"api_endpoint,omitempty"`
	APIToken     string   `yaml:"api_token,omitempty"`
}
