// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "Hickar",
            "url": "https://hickar.space",
            "email": "hickar@icloud.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/authorize": {
            "post": {
                "description": "Method for authorizing user with credentials, returning signed jwt in response",
                "consumes": [
                    "application/json"
                ],
                "summary": "Authorize user with username/password",
                "parameters": [
                    {
                        "description": "JSON with credentials",
                        "name": "login_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.AuthUserInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.AuthResponse"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "token": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "404": {
                        "description": ""
                    },
                    "422": {
                        "description": ""
                    }
                }
            }
        },
        "/authorize/email/challenge/{code}": {
            "get": {
                "description": "Method for enabling user via verification message sent by email",
                "summary": "Enable user",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Confirmation code",
                        "name": "confirmation_code",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.AuthResponse"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "token": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "404": {
                        "description": ""
                    },
                    "422": {
                        "description": ""
                    }
                }
            }
        },
        "/user": {
            "post": {
                "description": "Create new user with credentials provided in request. Response contains user JWT.",
                "consumes": [
                    "application/json"
                ],
                "summary": "Create new user",
                "parameters": [
                    {
                        "description": "JSON with user credentials",
                        "name": "new_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.CreateUserInput"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.AuthResponse"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "token": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "409": {
                        "description": ""
                    },
                    "422": {
                        "description": ""
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Method for updating user info: name, bio, avatar and birth date",
                "consumes": [
                    "application/json"
                ],
                "summary": "Update user info",
                "parameters": [
                    {
                        "description": "JSON with user info",
                        "name": "update_user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.UpdateUserInput"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    },
                    "401": {
                        "description": ""
                    },
                    "403": {
                        "description": ""
                    },
                    "404": {
                        "description": ""
                    },
                    "422": {
                        "description": ""
                    }
                }
            }
        },
        "/user/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get user by id",
                "consumes": [
                    "application/json"
                ],
                "summary": "Get user",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.UpdateUserInput"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "avatar": {
                                            "type": "string"
                                        },
                                        "bio": {
                                            "type": "string"
                                        },
                                        "birth_date": {
                                            "type": "string"
                                        },
                                        "name": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "401": {
                        "description": ""
                    },
                    "403": {
                        "description": ""
                    },
                    "404": {
                        "description": ""
                    },
                    "422": {
                        "description": ""
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Delete user by id",
                "consumes": [
                    "application/json"
                ],
                "summary": "Delete user",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": ""
                    },
                    "401": {
                        "description": ""
                    },
                    "403": {
                        "description": ""
                    },
                    "404": {
                        "description": ""
                    },
                    "422": {
                        "description": ""
                    }
                }
            }
        }
    },
    "definitions": {
        "api.AuthResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "api.AuthUserInput": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string",
                    "maxLength": 128
                },
                "password": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 6
                }
            }
        },
        "api.CreateUserInput": {
            "type": "object",
            "required": [
                "email",
                "name",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string",
                    "maxLength": 128
                },
                "name": {
                    "type": "string",
                    "maxLength": 128
                },
                "password": {
                    "type": "string",
                    "maxLength": 64,
                    "minLength": 6
                }
            }
        },
        "api.UpdateUserInput": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "avatar": {
                    "type": "string"
                },
                "bio": {
                    "type": "string",
                    "maxLength": 512
                },
                "birth_date": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "maxLength": 128
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "localhost:8080",
	BasePath:    "/api",
	Schemes:     []string{},
	Title:       "Gin-Rush API",
	Description: "Minimal API written on gin framework",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
