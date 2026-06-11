package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"apiTrackingSystem/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// Login - accepts username/password, calls external auth service, inserts/updates department and user
func Login(c *fiber.Ctx, db *sqlx.DB) error {
	var loginReq models.LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if loginReq.Username == "" || loginReq.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "username and password required"})
	}

	// Prepare request to external auth service
	extURL := "http://192.168.161.102:9999/login"
	reqBody := map[string]string{
		"username": loginReq.Username,
		"password": loginReq.Password,
	}
	b, _ := json.Marshal(reqBody)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Post(extURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "failed to contact auth service", "detail": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(502).JSON(fiber.Map{"error": "auth service returned non-200", "status": resp.Status})
	}

	var extResp models.ExternalLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&extResp); err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "invalid response from auth service", "detail": err.Error()})
	}

	// Check if status is EXPIRED
	if extResp.Status == "EXPIRED" {
		return c.Status(200).JSON(fiber.Map{"status": "EXPIRED"})
	}

	// Use department fields from external response
	sdSecCode := extResp.User.DepartmentCode // division maps to sd_sec_code
	sdName := extResp.User.Department
	sdCode := extResp.User.Division
	sdExtensionCode := extResp.User.ExtensionName
	now := time.Now()

	var suID int64
	var dbSpg sql.NullInt64
	var userSdID sql.NullInt64
	row := db.QueryRowx("SELECT su_id, spg_id, sd_id FROM sys_user WHERE su_username = ? LIMIT 1", extResp.User.Username)
	userErr := row.Scan(&suID, &dbSpg, &userSdID)
	if userErr != nil && userErr != sql.ErrNoRows {
		return c.Status(500).JSON(fiber.Map{"error": "failed to query user", "detail": userErr.Error()})
	}

	var sdID int64

	// Resolve department by sd_sec_code only.
	err = db.Get(&sdID, "SELECT sd_id FROM sys_department WHERE sd_sec_code = ? LIMIT 1", sdSecCode)
	if err != nil {
		if err == sql.ErrNoRows {
			res, insErr := db.Exec(`INSERT INTO sys_department (sd_name, sd_sec_name, sd_code, sd_status, sd_created_at, sd_created_by, sd_updated_at, sd_updated_by, sd_sec_code, sd_extention_code) VALUES (?, ?, ?, 'active', ?, ?, ?, ?, ?, ?)`,
				sdName, sdName, sdCode, now, "system", now, "system", sdSecCode, sdExtensionCode)
			if insErr != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to insert department", "detail": insErr.Error()})
			}
			newID, _ := res.LastInsertId()
			sdID = newID
		} else {
			return c.Status(500).JSON(fiber.Map{"error": "failed to query department", "detail": err.Error()})
		}
	} else {
		_, err = db.Exec(`UPDATE sys_department SET sd_name = ?, sd_sec_name = ?, sd_code = ?, sd_extention_code = ?, sd_updated_at = ?, sd_updated_by = ? WHERE sd_id = ?`,
			sdName, sdName, sdCode, sdExtensionCode, now, "system", sdID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update department", "detail": err.Error()})
		}
	}

	// default spgID
	spgID := int64(2)
	if userErr != nil {
		if userErr == sql.ErrNoRows {
			// insert new user, use default spgID
			_, err := db.Exec(`INSERT INTO sys_user (su_username, su_emp_code, su_firstname, su_lastname, su_email, su_status, spg_id, sd_id, su_created_at, su_created_by, su_updated_at, su_updated_by) VALUES (?, ?, ?, ?, ?, 'active', ?, ?, ?, ?, ?, ?)`,
				extResp.User.Username, extResp.User.EmployeeID, extResp.User.Name, extResp.User.Surname, extResp.User.Email, spgID, sdID, now, extResp.User.EmployeeID, now, extResp.User.EmployeeID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to insert user", "detail": err.Error()})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": "failed to query user", "detail": userErr.Error()})
		}
	}
	// else {
	// 	// existing user: if database has spg_id set, use it
	// 	if dbSpg.Valid {
	// 		spgID = dbSpg.Int64
	// 	}
	// 	// update existing user fields and link department
	// 	_, err := db.Exec(`UPDATE sys_user SET su_emp_code = ?, su_firstname = ?, su_lastname = ?, su_email = ?, sd_id = ?, spg_id = ?, su_updated_at = ?, su_updated_by = ? WHERE su_id = ?`,
	// 		extResp.User.EmployeeID, extResp.User.Name, extResp.User.Surname, extResp.User.Email, sdID, spgID, now, extResp.User.EmployeeID, suID)
	// 	if err != nil {
	// 		return c.Status(500).JSON(fiber.Map{"error": "failed to update user", "detail": err.Error()})
	// 	}
	// }

	// Return the required payload
	out := fiber.Map{
		"su_id":       suID,
		"sd_id":       sdID,
		"username":    extResp.User.Username,
		"displayName": extResp.User.DisplayName,
		"spg_id":      spgID,
		"employeeID":  extResp.User.EmployeeID,
	}

	return c.Status(200).JSON(out)
}
