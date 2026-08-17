package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	// โหลดไฟล์ .env
	// English: load .env in development
	_ = godotenv.Load()

	// โหลดค่าการตั้งค่าจาก .env
	// English: load config from env
	cfg := config.Load()

	// ทดสอบการเชื่อมต่อกับฐานข้อมูล
	// English: connect database
	db := database.MustOpen(cfg)

	// เช็คว่าเชื่อมต่อสำเร็จหรือไม่
	// English: print database connection info
	if db != nil {
		host, name := parseMySQLDSN(cfg.DBDSN)
		if host == "" && name == "" {
			fmt.Println("Database connected successfully!")
		} else {
			log.Printf("Database connected: host=%s db=%s", host, name)
		}
	}

	// กำหนดขนาด upload สูงสุดจาก .env
	// English: max upload size from env, default 1 GB
	maxUploadMB := getEnvInt("MAX_UPLOAD_SIZE_MB", 1024)

	// สร้างเซิร์ฟเวอร์ Fiber พร้อมรองรับไฟล์ใหญ่
	// English: create Fiber app with large upload support
	app := fiber.New(fiber.Config{
		BodyLimit:         maxUploadMB * 1024 * 1024, // MB -> bytes
		StreamRequestBody: true,
		ReadBufferSize:    64 * 1024,

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// log error ไว้ดูใน server
			// English: log server-side error
			log.Printf("Fiber error: method=%s path=%s ip=%s err=%v",
				c.Method(), c.Path(), c.IP(), err)

			// ถ้า body ใหญ่เกิน limit ของ Fiber
			// English: if request body exceeds Fiber limit
			if strings.Contains(strings.ToLower(err.Error()), "body too large") {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error":  "request body too large",
					"detail": fmt.Sprintf("upload exceeds MAX_UPLOAD_SIZE_MB=%d", maxUploadMB),
				})
			}

			// default handler
			// English: fallback error response
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error":  "server error",
				"detail": err.Error(),
			})
		},
	})

	// CORS middleware - allow browser origin to call this API
	// English: allow frontend origin
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://192.168.161.205:4009,https://pms-demo.tbkk.co.th",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	// GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o apiTrackingSystem9.exe .
	// health check
	// English: simple health route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":             "ok",
			"max_upload_size_mb": maxUploadMB,
		})
	})

	// เริ่มงาน background อัตโนมัติ
	// English: start background auto tasks
	handlers.StartDelayStatusScheduler(db)
	handlers.StartSendMailScheduler(db)

	// Setup all routes (pass db)
	// English: register routes
	routes.Setup(app, db)

	// รันเซิร์ฟเวอร์
	// English: start server
	addr := cfg.AppAddr
	if addr == "" {
		addr = ":9004"
	}

	log.Printf("Listening on %s | MAX_UPLOAD_SIZE_MB=%d", addr, maxUploadMB)

	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}

	return n
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
