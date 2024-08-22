package config

type Server struct {
	// "localhost"
	Address string
	// "8080"
	Port              int
	TokenSymmetricKey string
}
