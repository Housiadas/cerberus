package elasticsearch

// Config holds settings for the Elasticsearch client.
type Config struct {
	Addresses []string
	Username  string
	Password  string
}
