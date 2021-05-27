package client

type Client interface{
	GetServerVersion() (string, error)
}
