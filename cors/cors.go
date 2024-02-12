package cors

import (
	"strconv"
	"strings"

	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/collection"
)

type Cors struct {
	config Config
}

func (c Cors) HandlePreflightRequest(request contracts.RequestContract) contracts.ResponseContract {
	resp := content.NewResponse(nil, nil, 204)
	c.ConfigureAllowedOrigin(request, resp)
	if resp.Header("Access-Control-Allow-Origin") != "" {
		c.ConfigureAllowCredentials(resp)
		c.ConfigureAllowedMethods(request, resp)
		c.ConfigureAllowedHeaders(request, resp)
		c.ConfigureMaxAge(resp)
	}

	return c.VaryHeader(resp, "Access-Control-Request-Method")
}

func (c Cors) AddActualRequestHeaders(request contracts.RequestContract, response contracts.ResponseContract) contracts.ResponseContract {
	c.ConfigureAllowedOrigin(request, response)
	if response.Header("Access-Control-Allow-Origin") != "" {
		c.ConfigureAllowCredentials(response)
		c.ConfigureExposedHeaders(response)
	}
	return response
}

func (c Cors) ConfigureAllowedOrigin(request contracts.RequestContract, response contracts.ResponseContract) {
	origin := request.HeaderString("Origin")
	allowAny := collection.Contains(c.config.AllowedOrigins, "*")
	originAllowed := collection.Contains(c.config.AllowedOrigins, origin)
	if allowAny && !c.config.AllowCredentials {
		response.SetHeader("Access-Control-Allow-Origin", "*")
	} else {
		c.VaryHeader(response, "Origin")

		if origin != "" && (allowAny || originAllowed) {
			response.SetHeader("Access-Control-Allow-Origin", origin)
		}
	}
}

func (c Cors) ConfigureAllowCredentials(response contracts.ResponseContract) {
	if c.config.AllowCredentials {
		response.SetHeader("Access-Control-Allow-Credentials", "true")
	}
}

func (c Cors) ConfigureAllowedMethods(request contracts.RequestContract, response contracts.ResponseContract) {
	var methods string
	if collection.Contains(c.config.AllowedMethods, "*") {
		methods = request.HeaderString("Access-Control-Request-Method")
		c.VaryHeader(response, "Access-Control-Request-Method")
	} else {
		methods = strings.Join(c.config.AllowedMethods, ",")
	}
	response.SetHeader("Access-Control-Allow-Methods", methods)
}

func (c Cors) ConfigureAllowedHeaders(request contracts.RequestContract, response contracts.ResponseContract) {
	var methods string
	if collection.Contains(c.config.AllowedHeaders, "*") {
		methods = request.HeaderString("Access-Control-Request-Headers")
		c.VaryHeader(response, "Access-Control-Request-Headers")
	} else {
		methods = strings.Join(c.config.AllowedHeaders, ",")
	}
	response.SetHeader("Access-Control-Allow-Headers", methods)
}

func (c Cors) ConfigureMaxAge(response contracts.ResponseContract) {
	if c.config.MaxAge > 0 {
		response.SetHeader("Access-Control-Max-Age", strconv.Itoa(c.config.MaxAge))
	}
}
func (c Cors) ConfigureExposedHeaders(response contracts.ResponseContract) {
	if len(c.config.ExposedHeaders) > 0 {
		response.SetHeader("Access-Control-Expose-Headers", strings.Join(c.config.ExposedHeaders, ","))
	}
}

func (c Cors) VaryHeader(response contracts.ResponseContract, header string) contracts.ResponseContract {
	vary := response.Header("Vary")
	if vary == "" {
		response.SetHeader("Vary", header)
	} else if !collection.Contains(collection.Map(strings.Split(vary, ","), func(v string) string { return strings.TrimSpace(v) }), header) {
		response.SetHeader("Vary", vary+", "+header)
	}

	return response
}
