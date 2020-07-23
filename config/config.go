package config

type Config struct {
	//RemoteSource configures how to read a remotely served records.mapstructure file
	RemoteSource RemoteSourceConfig `mapstructure:"remote_source"`
	//LocalSource configures how to read a local records.mapstructure file
	LocalSource LocalSourceConfig `mapstructure:"local_source"`
	//Destination configures how to write the output file
	Destination DestinationConfig `mapstructure:"destination"`
}

type RemoteSourceConfig struct {
	Endpoint           string          `mapstructure:"endpoint"`
	InsecureSkipVerify bool            `mapstructure:"insecure_skip_verify"`
	CACerts            string          `mapstructure:"ca_certs"`
	BasicAuth          BasicAuthConfig `mapstructure:"basic_auth"`
	ScrapeInterval     int             `mapstructure:"scrape_interval"`
	Include            MatchSpec       `mapstructure:"include"`
}

type BasicAuthConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type DestinationConfig struct {
	File string `mapstructure:"file"`
}

type LocalSourceConfig struct {
	File           string `mapstructure:"file"`
	ScrapeInterval int    `mapstructure:"scrape_interval"`
}

type MatchSpec struct {
	Deployments []string `mapstructure:"deployments"`
}
