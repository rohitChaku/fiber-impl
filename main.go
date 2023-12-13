package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type orderedReqLog struct {
	Level           string `json:"level,omitempty"`
	ServerTimestamp string `json:"@timestamp,omitempty"`
	HttpMethod      string `json:"http_method,omitempty"`
	HttpUri         string `json:"http_uri,omitempty"`
	QueryString     string `json:"query_string,omitempty"`
	Body            string `json:"body,omitempty"`
	Status          int    `json:"status,omitempty"`
	ResBody         string `json:"res_body,omitempty"`
	TimeTaken       int64  `json:"time_taken_ms,omitempty"`
	PerfStats       string `json:"perf-stats,omitempty"`
	ClientIp        string `json:"client_ip,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	Port            string `json:"port,omitempty"`
	Host            string `json:"host,omitempty"`
	Module          string `json:"module,omitempty"`
	ReqHeaders      string `json:"req_headers,omitempty"`
}

func DefaultLogger(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
	logMap := orderedReqLog{
		Level:           "RLOG",
		ServerTimestamp: c.Context().Time().Format("2006-01-02T15:04:05.000Z0700"),
		HttpMethod:      c.Method(),
		HttpUri:         c.Path(),
		QueryString:     c.Request().URI().QueryArgs().String(),
		Body:            string(c.Body()),
		Status:          c.Response().StatusCode(),
		TimeTaken:       time.Since(c.Context().Time()).Milliseconds(),
		ResBody:         string(c.Response().Body()),
		ClientIp:        c.IP(),
		Protocol:        c.Protocol(),
		Port:            c.Port(),
		Host:            c.Hostname(),
		Module:          "MIDDLEWARE",
	}
	reqHeaders := make([]string, 0)
	for k, v := range c.GetReqHeaders() {
		reqHeaders = append(reqHeaders, k+"="+strings.Join(v, ","))
	}
	logMap.ReqHeaders = strings.Join(reqHeaders, "&")
	logMap.PerfStats = "Random Perf Stats"
	// Marshall Struct without HTML escape
	var log bytes.Buffer
	e := json.NewEncoder(&log)
	e.SetEscapeHTML(false)
	if err := e.Encode(logMap); err != nil {
		return 0, err
	}
	return output.WriteString(log.String())
}

// User struct for object binding
type User struct {
	Id    string   `params:"id" json:"id"`
	Name  string   `form:"name,default=Rohit" default:"Rohit" validate:"required" json:"name"`
	Email string   `form:"email,default=rohit@example.com" default:"rohit@example.com" validate:"required" json:"email"`
	Age   int      `form:"age,default=18" default:"18" validate:"max=100" json:"age"`
	Items []string `form:"items,default=initiated" default:"initiated" validate:"omitempty,dive,oneofarray=initiated ringing answered completed" json:"items"`
}

// MyRequest is a struct representing the incoming request body
type MyRequest struct {
	MyField *bool `json:"myField" form:"myField"`
}

func removeJSONExtension() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the current path
		path := c.Path()

		// Check if the path ends with ".json"
		if strings.HasSuffix(path, ".json") {
			// Remove ".json" from the end of the path
			newPath := strings.TrimSuffix(path, ".json")

			// Update the request path
			c.Path(newPath)

			log.Println("Rewrite Path")
		}

		// Continue processing
		return c.Next()
	}
}

func Serve() {
	validate := validator.New()
	validate.RegisterValidation("oneofarray", func(fl validator.FieldLevel) bool {
		// Get the field value
		options := strings.Fields(fl.Field().String())

		// Get the allowed options from the struct tag
		allowedOptions := fl.Param()

		// Split the allowed options into a slice
		allowedOptionsSlice := strings.Fields(allowedOptions)

		// Run the HasAny check
		for _, option := range options {
			if !slices.Contains(allowedOptionsSlice, option) {
				return false
			}
		}

		return true
	})
	// Create a new Fiber instance
	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format: "${custom_tag}\n",
		CustomTags: map[string]logger.LogFunc{
			"custom_tag": DefaultLogger,
		},
	}))
	app.Use(recover.New())
	app.Use(pprof.New())

	// Middleware for logging
	app.Use(func(c *fiber.Ctx) error {
		log.Println("Middleware 1: Before request")
		c.Next()
		log.Println("Middleware 1: After request")
		fmt.Printf("ROUTE NAME: %+v\n", c.Route().Name)
		return nil
	})

	// Middleware for logging
	app.Use(func(c *fiber.Ctx) error {
		log.Println("Middleware 2: Before request")
		c.Next()
		log.Println("Middleware 2: After request")
		return nil
	})

	app.Use(removeJSONExtension())

	// Endpoint to handle POST requests with JSON input + Form Data
	app.Post("/api/user", func(c *fiber.Ctx) error {
		// Create an instance of the User struct
		user := new(User)

		// Bind the JSON input to the User struct
		if err := c.BodyParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := validate.Struct(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Do something with the user object (e.g., save it to a database)
		log.Printf("Received user: %+v", user)

		time.Sleep(300 * time.Millisecond)

		return c.JSON(fiber.Map{
			"message": "User created successfully",
			"user":    user,
		})
	}).Name("Random Name")

	// Endpoint to handle POST requests with JSON input
	app.Post("/api/user/:id", func(c *fiber.Ctx) error {
		// Create an instance of the User struct
		user := new(User)

		if err := c.ParamsParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Bind the JSON input to the User struct
		if err := c.BodyParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := validate.Struct(user); err != nil {
			return err
		}

		// Do something with the user object (e.g., save it to a database)
		log.Printf("Received user: %+v", user)

		return c.JSON(fiber.Map{
			"message": "User created successfully",
			"user":    user,
		})
	})

	app.Post("/panic", func(c *fiber.Ctx) error {
		panic("panic now")
	})

	api := app.Group("/api") // /api

	v1 := api.Group("/v1") // /api/v1
	v1.Get("/list", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "/v1/list"})
	}) // /api/v1/list
	v1.Get("/user", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "/v1/user"})
	}) // /api/v1/user

	// Endpoint to handle POST requests with JSON input
	app.Post("/api/user/default/:id", func(c *fiber.Ctx) error {
		// Create an instance of the User struct
		user := new(User)
		MapFormDefault(&user)
		if err := c.ParamsParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Bind the JSON input to the User struct
		if err := c.BodyParser(user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if err := validate.Struct(user); err != nil {
			return err
		}

		// Do something with the user object (e.g., save it to a database)
		log.Printf("Received user: %+v", user)

		return c.JSON(fiber.Map{
			"message": "User created successfully",
			"user":    user,
		})
	})

	app.Post("/example", func(c *fiber.Ctx) error {
		// Parse the request body into a MyRequest struct
		var req MyRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Check if MyField is present and explicitly set to true or false
		if req.MyField != nil {
			fmt.Printf("myField is explicitly set to: %v\n", *req.MyField)
		} else {
			fmt.Println("myField is not present in the request body\n")
		}

		return c.SendString("Request processed")
	})

	// Start the Fiber application on port 3000
	log.Fatal(app.Listen(":3000"))
}

func main() {
	Serve()
}
