// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://somewhere.com/",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/orders": {
            "get": {
                "description": "get all order",
                "produces": [
                    "application/json"
                ],
                "summary": "getOrders",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/oms.OrderForSwag"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "new 1 Order api",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "newOrder",
                "parameters": [
                    {
                        "description": "Order data for sending to executor ",
                        "name": "Order",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/oms.OrderForSwag"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/orders/:id": {
            "get": {
                "description": "get 1 order",
                "produces": [
                    "application/json"
                ],
                "summary": "getOrder",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id of order",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/oms.OrderForSwag"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "oms.OrderForSwag": {
            "type": "object",
            "properties": {
                "account": {
                    "type": "string"
                },
                "avg_px": {
                    "type": "string"
                },
                "clord_id": {
                    "type": "string"
                },
                "closed": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "maturity_day": {
                    "type": "integer"
                },
                "maturity_month_year": {
                    "type": "string"
                },
                "open": {
                    "type": "string"
                },
                "ord_type": {
                    "type": "string"
                },
                "order_id": {
                    "type": "string"
                },
                "price": {
                    "type": "string"
                },
                "put_or_call": {
                    "type": "string"
                },
                "quantity": {
                    "type": "string"
                },
                "security_desc": {
                    "type": "string"
                },
                "security_type": {
                    "type": "string"
                },
                "session_id": {
                    "type": "string"
                },
                "side": {
                    "type": "string"
                },
                "stop_price": {
                    "type": "string"
                },
                "strike_price": {
                    "type": "string"
                },
                "symbol": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{"https", "http"},
	Title:            "Traderui API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
