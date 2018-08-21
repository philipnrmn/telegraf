package api

type Container struct {
	Id string `json:"id"`

	StatsdHost string `json:"statsd_host,omitempty"`

	StatsdPort float32 `json:"statsd_port,omitempty"`
}
