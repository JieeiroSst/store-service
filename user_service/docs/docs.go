// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/authentication": {
            "post": {
                "description": "Authentication account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Authentication Account",
                "parameters": [
                    {
                        "description": "token",
                        "name": "token",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Token"
                        }
                    },
                    {
                        "type": "string",
                        "description": "username",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/api/v1/lock_user": {
            "post": {
                "description": "LockAccount account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "LockAccount Account",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/api/v1/login": {
            "post": {
                "description": "login account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Login Account",
                "parameters": [
                    {
                        "description": "user in json",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Login"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/api/v1/sign-up": {
            "post": {
                "description": "SignUp account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "SignUp Account",
                "parameters": [
                    {
                        "description": "user in json login",
                        "name": "user",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/model.Users"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/api/v1/update/profile": {
            "put": {
                "description": "UpdateProfile account",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "UpdateProfile Account",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "id in json login",
                        "name": "id",
                        "in": "path"
                    },
                    {
                        "description": "user in json",
                        "name": "user",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/model.Users"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.Login": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.Token": {
            "type": "object",
            "properties": {
                "encodeToken": {
                    "type": "string"
                }
            }
        },
        "model.Users": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "checked": {
                    "type": "boolean"
                },
                "create_time": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "sex": {
                    "type": "string"
                },
                "update_time": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
