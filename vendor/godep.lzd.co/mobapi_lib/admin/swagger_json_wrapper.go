package admin

import (
	"strings"

	"github.com/sergei-svistunov/gorpc/transport/http_json"
	"godep.lzd.co/mobapi_lib/handlersmanager"
)

func NewSwaggerJSONCallbacks(serviceID, venture string) http_json.SwaggerJSONCallbacks {
	return http_json.SwaggerJSONCallbacks{
		Process: func(swagger *http_json.Swagger) {
			swagger.Info.Title = serviceID
			swagger.Info.Description += `<h3>Access Token</h3><p>Some handlers require access token. It must be provided in <code>X-API-TOKEN</code> header. There are guest and auth tokens which can be acquired in dedicated \"customer\" handlers.</p><p></p>`
			swagger.SecurityDefinitions = http_json.SecurityDefinitions{
				"api_token": &http_json.SecurityScheme{
					Type:        "apiKey",
					Description: "API token",
					Name:        "X-API-TOKEN",
					In:          "header",
					Xtensions:   []string{"mobapi_auth"},
				},
			}

			for _, item := range swagger.Paths {
				for _, operation := range item {
					for _, param := range operation.Parameters {
						if strings.ToLower(param.Name) == "lang" {
							var langs string
							switch strings.ToLower(venture) {
							case "ph", "sg":
								langs += "'en'"
							case "id":
								langs += "'id'"
							case "vn":
								langs += "'vi', 'en'"
							case "my":
								langs += "'ms', 'en'"
							case "th":
								langs += "'th', 'en'"
							default:
								langs += "[unknown]"
							}
							param.Description = "Possible values: <b>" + langs + "</b>.<br/>" + param.Description
							break
						}
					}

					if tokenType, ok := operation.ExtraData.(handlersmanager.TokenType); ok {
						var tokenInfo string
						switch tokenType {
						case handlersmanager.TokenTypeAuthorized:
							tokenInfo = `Handler requires <strong>auth token</strong>.`
						case handlersmanager.TokenTypeAny:
							tokenInfo = `Handler accepts any token but it's not mandatory.`
						case handlersmanager.TokenTypeGuest:
							tokenInfo = `Handler requires <strong>guest token</strong>.`
						}
						if tokenInfo != "" {
							operation.Description += "<br/>" + tokenInfo
							operation.Security = append(operation.Security, &http_json.SecurityRequirement{"api_token": []string{}})
						}
					}
				}
			}
		},
	}
}
