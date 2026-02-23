package main

import (
	"fmt"
	"log"
	"strings"

	"apiTrackingSystem/config"
	"apiTrackingSystem/database"
	"apiTrackingSystem/internal/handlers"
	"apiTrackingSystem/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db := database.MustOpen(cfg)

	// เช็คว่าเชื่อมต่อสำเร็จหรือไม่
	if db != nil {
		host, name := parseMySQLDSN(cfg.DBDSN)
		if host == "" && name == "" {
			fmt.Println("Database connected successfully!")
		} else {
			log.Printf("Database connected: host=%s db=%s", host, name)
		}
	}

	app := fiber.New()

	// CORS middleware - allow browser origin to call this API
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://192.168.161.205:4001",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Start background scheduler: marks overdue items as 'delay' at midnight
	handlers.StartDelayStatusScheduler(db)

	// Start background scheduler: sends tracking email every Mon/Wed/Fri at 08:00
	handlers.StartSendMailScheduler(db)

	// Setup all routes (pass db)
	routes.Setup(app, db)

	// รันเซิร์ฟเวอร์
	addr := cfg.AppAddr
	if addr == "" {
		addr = ":9005"
	}
	log.Printf("listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func parseMySQLDSN(dsn string) (host string, dbname string) {
	// extract host inside tcp(...)
	if idx := strings.Index(dsn, "tcp("); idx != -1 {
		start := idx + len("tcp(")
		if end := strings.Index(dsn[start:], ")"); end != -1 {
			host = dsn[start : start+end]
		}
	}

	// extract dbname after ")/" up to ? or end
	if idx := strings.Index(dsn, ")/"); idx != -1 {
		start := idx + len(")/")
		if end := strings.Index(dsn[start:], "?"); end != -1 {
			dbname = dsn[start : start+end]
		} else {
			dbname = dsn[start:]
		}
	}

	return
}
