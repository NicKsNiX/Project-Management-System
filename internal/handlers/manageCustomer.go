package handlers

import (
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Customer struct {
	ID        int64            `db:"mc_id" json:"mc_id"`
	Name      utils.NullString `db:"mc_name" json:"mc_name"`
	Status    utils.NullString `db:"mc_status" json:"mc_status"`
	CreatedAt *time.Time       `db:"mc_created_at" json:"mc_created_at"`
	CreatedBy utils.NullString `db:"mc_created_by" json:"mc_created_by"`
	UpdatedAt *time.Time       `db:"mc_updated_at" json:"mc_updated_at"`
	UpdatedBy utils.NullString `db:"mc_updated_by" json:"mc_updated_by"`
}

func ListCustomers(c *fiber.Ctx, db *sqlx.DB) error {
	query := `SELECT
					mc_id,
					mc_name,
					mc_status,
					mc_created_at,
					mc_created_by,
					mc_updated_at,
					mc_updated_by
				FROM mst_customer
				WHERE mc_status = 'active'
				ORDER BY mc_id ASC`

	var list []Customer
	if err := db.Select(&list, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}
