package main

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	secretEnvError = errors.New("missing SECRET environment variable")
)

type webhook struct {
	secret string
	e      nostrEvent
}

type nostrEvent struct {
	Id           string `json:"id"`
	Pubkey       string `json:"pubkey"`
	CreatedAtRaw int64  `json:"created_at"`
	CreatedAt    *time.Time
	Kind         int      `json:"kind"`
	Tags         []string `json:"tags"`
	Content      string   `json:"content"`
	Sig          string   `json:"sig"`
}

func main() {
	w := webhook{}
	if err := w.readSecret(); err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Use(recover.New())
	app.Use(healthcheck.New())
	app.Use(func(c *fiber.Ctx) error {
		slog.Info("request", "ip", c.IP(), "port", c.Port(), "protocol", c.Protocol(), "method", c.Method(), "path", c.Path())
		slog.Info("request headers", "headers", c.GetReqHeaders())
		slog.Info("request body", "body", string(c.Body()))
		return c.Next()
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Route("/webhook", w.webhookRoute)

	app.Listen(":3000")
}

func (w *webhook) readSecret() error {
	ok := false
	w.secret, ok = os.LookupEnv("SECRET")
	if !ok {
		return secretEnvError
	}
	return nil
}

func (w *webhook) webhookRoute(r fiber.Router) {
	r.Get("/test", w.testHandler)
	r.Post("/test", w.testHandler)
}

func (w *webhook) testHandler(c *fiber.Ctx) error {
	err := bearerAuthentication(c, w.secret)
	if err != nil {
		slog.Error("authentication failed", "err", err, "header", c.GetReqHeaders())
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err = w.handleEvent(c)
	if err != nil {
		slog.Error("error parsing event", "err", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	slog.Info("event", "event", w.e)

	return c.SendString("ok")
}

func bearerAuthentication(c *fiber.Ctx, secret string) error {
	auth := c.Get("Authorization", "")
	if auth == "" {
		return errors.New("missing authorization header")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return errors.New("invalid authorization header")
	}

	if parts[0] != "Bearer" || parts[1] != secret {
		return errors.New("invalid bearer token")
	}

	return nil
}

func (w *webhook) handleEvent(c *fiber.Ctx) error {
	if err := c.BodyParser(&w.e); err != nil {
		return err
	}

	createdAt := time.Unix(w.e.CreatedAtRaw, 0)
	w.e.CreatedAt = &createdAt

	return nil
}
