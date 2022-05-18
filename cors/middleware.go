package cors

import (
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/pipeline"
)

type Config struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type Middleware struct {
	cors Cors
}

func (m *Middleware) Handle(request contracts.RequestContract, next pipeline.PipeHandler) contracts.ResponseContract {
	if m.isPreflight(request) {
		return m.cors.HandlePreflightRequest(request)
	}
	resp := next(request)
	if request.GetMethod() == "OPTIONS" {
		return m.cors.VaryHeader(resp, "Access-Control-Request-Method")
	}

	return m.cors.AddActualRequestHeaders(request, resp)
}

func (m *Middleware) isPreflight(request contracts.RequestContract) bool {
	return request.GetMethod() == "OPTIONS" && len(request.Header("Access-Control-Request-Method")) > 0
}

func NewMiddleware(config Config) *Middleware {
	return &Middleware{
		cors: Cors{config: config},
	}
}
