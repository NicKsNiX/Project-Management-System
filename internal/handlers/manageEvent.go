package handlers

import (
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type CustomerEvent struct {
	ID                 int64            `db:"mce_id" json:"mce_id"`
	MmmID              utils.NullInt64  `db:"mmm_id" json:"mmm_id"`
	MMMName            utils.NullString `db:"mmm_model" json:"mmm_model"`
	MMCustomerName     utils.NullString `db:"mmm_customer_name" json:"mmm_customer_name"`
	Name               utils.NullString `db:"mce_name" json:"mce_name"`
	Status             string           `db:"mce_status" json:"mce_status"`
	CreatedAt          *time.Time       `db:"mce_created_at" json:"mce_created_at"`
	CreatedBy          utils.NullString `db:"mce_created_by" json:"mce_created_by"`
	UpdatedAt          *time.Time       `db:"mce_updated_at" json:"mce_updated_at"`
	UpdatedBy          utils.NullString `db:"mce_updated_by" json:"mce_updated_by"`
	UpdatedByFirstName utils.NullString `db:"su_firstname" json:"updated_by_first_name"`
	UpdatedByLastName  utils.NullString `db:"su_lastname" json:"updated_by_last_name"`
}

func ListCustomerEvents(c *fiber.Ctx, db *sqlx.DB) error {
	status := c.Query("mce_status")
	query := `SELECT mce_id, mce.mmm_id, mce_name, mce_status,
					 mmm.mmm_model,mmm.mmm_customer_name,
	                 mce_created_at, mce_created_by,
	                 mce_updated_at, mce_updated_by,
	                 su.su_firstname, su.su_lastname
	          FROM mst_customer_event mce
			  LEFT JOIN mst_model_master mmm ON mce.mmm_id = mmm.mmm_id
	          LEFT JOIN sys_user su ON su.su_emp_code = mce.mce_updated_by
	          WHERE 1=1`
	args := []interface{}{}
	if status != "" {
		query += " AND mce_status = ?"
		args = append(args, status)
	}
	query += " ORDER BY mce_id ASC"

	var list []CustomerEvent
	if err := db.Select(&list, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func InsertCustomerEvent(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		MmmID     int64  `json:"mmm_id"`
		Name      string `json:"mce_name"`
		CreatedBy string `json:"mce_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mce_name is required"})
	}

	// Duplicate check (same name + mmm_id)
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_customer_event WHERE mce_name = ? AND mmm_id = ?`, body.Name, body.MmmID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	_, err := db.Exec(`INSERT INTO mst_customer_event
	                   (mmm_id, mce_name, mce_status, mce_created_at, mce_created_by, mce_updated_at, mce_updated_by)
	                   VALUES (?, ?, 'active', ?, ?, ?, ?)`,
		body.MmmID, body.Name, now, body.CreatedBy, now, body.CreatedBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "insert error", "detail": err.Error()})
	}
	return c.Status(201).JSON(1)
}

func UpdateCustomerEvent(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mce_id"`
		MmmID     int64  `json:"mmm_id"`
		Name      string `json:"mce_name"`
		UpdatedBy string `json:"mce_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mce_id is required"})
	}

	// Duplicate check (same name + mmm_id, excluding current record)
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_customer_event WHERE mce_name = ? AND mmm_id = ? AND mce_id <> ?`, body.Name, body.MmmID, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_customer_event
	                     SET mmm_id = ?, mce_name = ?, mce_updated_at = ?, mce_updated_by = ?
	                     WHERE mce_id = ?`,
		body.MmmID, body.Name, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update error", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "customer event not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdateCustomerEventStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mce_id"`
		Status    string `json:"mce_status"`
		UpdatedBy string `json:"mce_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mce_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "mce_status must be 'active' or 'inactive'"})
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_customer_event SET mce_status = ?, mce_updated_at = ?, mce_updated_by = ? WHERE mce_id = ?`,
		body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update error", "detail": err.Error()})
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "customer event not found"})
	}
	return c.Status(200).JSON(1)
}
