package twitter

type Config struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
} // Config

func init() {
	config := Config{}
	// Twitter API credentials
	config.ConsumerKey = "wmoPCSpLArkwc1TcQ1XNokPFA"
	config.ConsumerSecret = "FNkplAcqAOGOLjK0zUUtQCoB17y32G4xJmUg490UbYvroYRlHN"
	config.AccessToken = "1055003306509115395-GixVy9W42It7018BuOMZKJuiJO5zrS"
	config.AccessTokenSecret = "ujq00W66PsK6J49wia3qbXUuZrZwvzndtUGbzA0GqqU8E"
} // init

func config() *Config {
	config := Config{
		ConsumerKey:       "wmoPCSpLArkwc1TcQ1XNokPFA",
		ConsumerSecret:    "FNkplAcqAOGOLjK0zUUtQCoB17y32G4xJmUg490UbYvroYRlHN",
		AccessToken:       "1055003306509115395-GixVy9W42It7018BuOMZKJuiJO5zrS",
		AccessTokenSecret: "ujq00W66PsK6J49wia3qbXUuZrZwvzndtUGbzA0GqqU8E",
	} // config
	return &config
} // Init

func GetConfig() Config {
	config := config()
	config.SetConfig(*config)
	return *config
} // GetConfig

func (c *Config) GetConsumerKey() string {
	return c.ConsumerKey
} // GetConsumerKey

func (c *Config) GetConsumerSecret() string {
	return c.ConsumerSecret
} // GetConsumerSecret

func (c *Config) GetAccessToken() string {
	return c.AccessToken
} // GetAccessToken
func (c *Config) GetAccessTokenSecret() string {
	return c.AccessTokenSecret
} // GetAccessTokenSecret
func (c *Config) SetConsumerKey(consumerKey string) {
	c.ConsumerKey = consumerKey
} // SetConsumerKey

func (c *Config) SetConsumerSecret(consumerSecret string) {
	c.ConsumerSecret = consumerSecret
} // SetConsumerSecret
func (c *Config) SetAccessToken(accessToken string) {
	c.AccessToken = accessToken
} // SetAccessToken

func (c *Config) SetAccessTokenSecret(accessTokenSecret string) {
	c.AccessTokenSecret = accessTokenSecret
} // SetAccessTokenSecret
func (c *Config) GetConfig() Config {
	return *c
} // GetConfig
func (c *Config) SetConfig(config Config) {
	c.ConsumerKey = config.ConsumerKey
	c.ConsumerSecret = config.ConsumerSecret
	c.AccessToken = config.AccessToken
	c.AccessTokenSecret = config.AccessTokenSecret
} // SetConfig

func (c *Config) GetConfigFile() string {
	return "config.json"
} // GetConfigFile
