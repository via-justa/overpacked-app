package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/via-justa/overpacked-app/backend/internal/app"
	"github.com/via-justa/overpacked-app/backend/internal/auth"
	"github.com/via-justa/overpacked-app/backend/internal/http/handlers"
	"github.com/via-justa/overpacked-app/backend/internal/store"
	"gopkg.in/yaml.v3"
)

type openAPIDoc struct {
	OpenAPI    string              `yaml:"openapi"`
	Info       infoObject          `yaml:"info"`
	Servers    []serverObject      `yaml:"servers"`
	Paths      map[string]pathItem `yaml:"paths"`
	Components componentsObject    `yaml:"components"`
}

type infoObject struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

type serverObject struct {
	URL string `yaml:"url"`
}

type pathItem struct {
	Get    *operationObject `yaml:"get,omitempty"`
	Post   *operationObject `yaml:"post,omitempty"`
	Put    *operationObject `yaml:"put,omitempty"`
	Patch  *operationObject `yaml:"patch,omitempty"`
	Delete *operationObject `yaml:"delete,omitempty"`
}

type operationObject struct {
	Summary     string                    `yaml:"summary"`
	OperationID string                    `yaml:"operationId"`
	Tags        []string                  `yaml:"tags,omitempty"`
	RequestBody *requestBodyObject        `yaml:"requestBody,omitempty"`
	Responses   map[string]responseObject `yaml:"responses"`
}

type requestBodyObject struct {
	Required bool                       `yaml:"required"`
	Content  map[string]mediaTypeObject `yaml:"content"`
}

type responseObject struct {
	Description string                     `yaml:"description"`
	Content     map[string]mediaTypeObject `yaml:"content,omitempty"`
}

type mediaTypeObject struct {
	Schema schemaObject `yaml:"schema"`
}

type schemaObject struct {
	Ref        string                  `yaml:"$ref,omitempty"`
	Type       string                  `yaml:"type,omitempty"`
	Properties map[string]schemaObject `yaml:"properties,omitempty"`
	Required   []string                `yaml:"required,omitempty"`
}

type componentsObject struct {
	Schemas         map[string]schemaObject         `yaml:"schemas"`
	SecuritySchemes map[string]securitySchemeObject `yaml:"securitySchemes"`
}

type securitySchemeObject struct {
	Type   string `yaml:"type"`
	Scheme string `yaml:"scheme"`
	Bearer string `yaml:"bearerFormat,omitempty"`
}

type route struct {
	Method  string
	Pattern string
}

const errorSchemaRef = "#/components/schemas/ErrorResponse"

func main() {
	authService, err := auth.NewService("openapi", "openapi", "openapi-dev-secret")
	if err != nil {
		panic(fmt.Errorf("init auth service: %w", err))
	}
	router := app.NewRouter(handlers.NewAuthHandler(authService), store.New(nil), "openapi")

	routes, err := collectRoutes(router)
	if err != nil {
		panic(fmt.Errorf("collect routes: %w", err))
	}

	doc := buildDoc(routes)
	out, err := yaml.Marshal(doc)
	if err != nil {
		panic(fmt.Errorf("marshal openapi: %w", err))
	}

	outputPath := filepath.Join("..", "dev", "openapi.yaml")
	if err := os.WriteFile(outputPath, out, 0o644); err != nil {
		panic(fmt.Errorf("write openapi file: %w", err))
	}

	fmt.Printf("generated %s\n", outputPath)
}

func collectRoutes(r chi.Router) ([]route, error) {
	collected := make([]route, 0)
	err := chi.Walk(r, func(method string, routePattern string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if method == "*" {
			return nil
		}
		collected = append(collected, route{Method: strings.ToLower(method), Pattern: routePattern})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(collected, func(i, j int) bool {
		if collected[i].Pattern == collected[j].Pattern {
			return collected[i].Method < collected[j].Method
		}
		return collected[i].Pattern < collected[j].Pattern
	})

	return collected, nil
}

func buildDoc(routes []route) openAPIDoc {
	doc := openAPIDoc{
		OpenAPI: "3.1.0",
		Info:    infoObject{Title: "Packing List API", Version: "0.1.0"},
		Servers: []serverObject{{URL: "http://localhost:8000"}},
		Paths:   map[string]pathItem{},
		Components: componentsObject{
			Schemas: map[string]schemaObject{
				"AuthLoginRequest": {
					Type: "object",
					Properties: map[string]schemaObject{
						"username": {Type: "string"},
						"password": {Type: "string"},
					},
					Required: []string{"username", "password"},
				},
				"AuthRefreshRequest": {
					Type: "object",
					Properties: map[string]schemaObject{
						"refresh_token": {Type: "string"},
					},
					Required: []string{"refresh_token"},
				},
				"AuthTokenResponse": {
					Type: "object",
					Properties: map[string]schemaObject{
						"access_token":  {Type: "string"},
						"refresh_token": {Type: "string"},
						"token_type":    {Type: "string"},
						"expires_in":    {Type: "integer"},
					},
					Required: []string{"access_token", "refresh_token", "token_type", "expires_in"},
				},
				"ErrorResponse": {
					Type: "object",
					Properties: map[string]schemaObject{
						"error": {Type: "string"},
					},
					Required: []string{"error"},
				},
			},
			SecuritySchemes: map[string]securitySchemeObject{
				"bearerAuth": {
					Type:   "http",
					Scheme: "bearer",
					Bearer: "JWT",
				},
			},
		},
	}

	for _, rt := range routes {
		op := operationForRoute(rt)
		item := doc.Paths[rt.Pattern]
		switch rt.Method {
		case "get":
			item.Get = &op
		case "post":
			item.Post = &op
		case "put":
			item.Put = &op
		case "patch":
			item.Patch = &op
		case "delete":
			item.Delete = &op
		}
		doc.Paths[rt.Pattern] = item
	}

	return doc
}

func operationForRoute(rt route) operationObject {
	key := rt.Method + " " + rt.Pattern
	defaultOp := operationObject{
		Summary:     strings.ToUpper(rt.Method) + " " + rt.Pattern,
		OperationID: operationID(rt.Method, rt.Pattern),
		Tags:        []string{firstTag(rt.Pattern)},
		Responses: map[string]responseObject{
			"200": {Description: "OK"},
		},
	}

	switch key {
	case "get /health":
		return operationObject{
			Summary:     "Health check",
			OperationID: "healthCheck",
			Tags:        []string{"health"},
			Responses: map[string]responseObject{
				"200": {
					Description: "Service healthy",
					Content: map[string]mediaTypeObject{
						"text/plain": {Schema: schemaObject{Type: "string"}},
					},
				},
			},
		}
	case "post /api/v1/auth/login":
		return operationObject{
			Summary:     "Login with configured app credentials",
			OperationID: "authLogin",
			Tags:        []string{"auth"},
			RequestBody: jsonBodyRef("#/components/schemas/AuthLoginRequest"),
			Responses: map[string]responseObject{
				"200": jsonResp("Authenticated", "#/components/schemas/AuthTokenResponse"),
				"400": jsonResp("Invalid request body", errorSchemaRef),
				"401": jsonResp("Invalid credentials", errorSchemaRef),
			},
		}
	case "post /api/v1/auth/refresh":
		return operationObject{
			Summary:     "Refresh JWT token pair",
			OperationID: "authRefresh",
			Tags:        []string{"auth"},
			RequestBody: jsonBodyRef("#/components/schemas/AuthRefreshRequest"),
			Responses: map[string]responseObject{
				"200": jsonResp("Token refreshed", "#/components/schemas/AuthTokenResponse"),
				"400": jsonResp("Invalid request body", errorSchemaRef),
				"401": jsonResp("Invalid refresh token", errorSchemaRef),
			},
		}
	case "post /api/v1/auth/logout":
		return operationObject{
			Summary:     "Logout current client session",
			OperationID: "authLogout",
			Tags:        []string{"auth"},
			Responses: map[string]responseObject{
				"204": {Description: "Logged out"},
			},
		}
	default:
		return defaultOp
	}
}

func jsonBodyRef(ref string) *requestBodyObject {
	return &requestBodyObject{
		Required: true,
		Content: map[string]mediaTypeObject{
			"application/json": {Schema: schemaObject{Ref: ref}},
		},
	}
}

func jsonResp(desc, ref string) responseObject {
	return responseObject{
		Description: desc,
		Content: map[string]mediaTypeObject{
			"application/json": {Schema: schemaObject{Ref: ref}},
		},
	}
}

func operationID(method, pattern string) string {
	clean := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_").Replace(pattern)
	clean = strings.Trim(clean, "_")
	return method + "_" + clean
}

func firstTag(pattern string) string {
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "default"
	}
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return parts[2]
	}
	return parts[0]
}
