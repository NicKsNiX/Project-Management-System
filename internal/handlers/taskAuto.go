package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func CheckAndUpdateDelayStatus(db *sqlx.DB) error {
	// ── Step 1: mark overdue detail rows as 'delay' ──────────────────────────
	queryDetail := `
		UPDATE info_project_item_detail
		SET
			ipid_status     = 'delay',
			ipid_updated_at = NOW(),
			ipid_updated_by = 'system'
		WHERE
			ipid_end_date IS NOT NULL
			AND ipid_end_date < CURDATE()
			AND ipid_status NOT IN ('done', 'reject', 'delay')
			AND (
				ipid_status <> 'waiting'
				OR (
					EXISTS (
						SELECT 1
						FROM info_approval ia
						WHERE ia.ipid_id = info_project_item_detail.ipid_id
							AND ia.ia_type = 'Leader'
							AND ia.ia_status = 'waiting'
							AND ia.ia_status_flg = 'active'
					)
					AND NOT EXISTS (
						SELECT 1
						FROM info_approval ia
						WHERE ia.ipid_id = info_project_item_detail.ipid_id
							AND ia.ia_type = 'PJ'
							AND ia.ia_status_flg = 'active'
					)
				)
			)
	`

	result, err := db.Exec(queryDetail)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	log.Printf("[taskAuto] CheckAndUpdateDelayStatus: updated %d info_project_item_detail row(s) to 'delay'", rows)

	queryMasterPlan := `
		UPDATE info_project_master_plan ipmp
		INNER JOIN mst_master_plan      mmp   ON mmp.mmp_name  = ipmp.ipmp_name
		INNER JOIN mst_master_plan_detail mmpd ON mmpd.mmp_id  = mmp.mmp_id
		INNER JOIN mst_apqp             ma    ON ma.ma_id      = mmpd.ma_id
		INNER JOIN info_apqp_item       iai   ON iai.iai_name  = ma.ma_name
		INNER JOIN info_project_item_detail ipid
			ON  ipid.ref_id      = iai.iai_id
			AND ipid.ipid_type   = 'apqp'
			AND ipid.ipid_end_date IS NOT NULL
			AND ipid.ipid_end_date < CURDATE()
			AND ipid.ipid_status = 'delay'
		SET
			ipmp.ipmp_status = 'delay'
		WHERE
			ipmp.ip_id = iai.ip_id
			AND ipmp.ipmp_status NOT IN ('done', 'reject', 'delay')
	`

	resultMP, err := db.Exec(queryMasterPlan)
	if err != nil {
		return err
	}
	rowsMP, _ := resultMP.RowsAffected()
	log.Printf("[taskAuto] CheckAndUpdateDelayStatus: updated %d info_project_master_plan row(s) to 'delay'", rowsMP)

	return nil
}

func durationUntilMidnight() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
	return time.Until(next)
}

func StartDelayStatusScheduler(db *sqlx.DB) {
	go func() {
		// Run once right away on startup
		if err := CheckAndUpdateDelayStatus(db); err != nil {
			log.Printf("[taskAuto] initial delay-check error: %v", err)
		}

		for {
			wait := durationUntilMidnight()
			log.Printf("[taskAuto] next delay-check in %.2f hours (at midnight)", wait.Hours())

			timer := time.NewTimer(wait)
			<-timer.C

			if err := CheckAndUpdateDelayStatus(db); err != nil {
				log.Printf("[taskAuto] scheduled delay-check error: %v", err)
			}
		}
	}()
}

func durationUntilNextMailDay(hour, minute int) time.Duration {
	now := time.Now()
	mailDays := map[time.Weekday]bool{
		time.Monday:    true,
		time.Wednesday: true,
		time.Friday:    true,
	}

	// Search up to 7 days ahead
	for daysAhead := 0; daysAhead <= 7; daysAhead++ {
		candidate := time.Date(now.Year(), now.Month(), now.Day()+daysAhead, hour, minute, 0, 0, now.Location())
		if mailDays[candidate.Weekday()] && candidate.After(now) {
			return time.Until(candidate)
		}
	}
	// Fallback: 7 days (should never reach here)
	return 7 * 24 * time.Hour
}

func StartSendMailScheduler(db *sqlx.DB) {
	go func() {
		for {
			wait := durationUntilNextMailDay(8, 0)
			log.Printf("[taskAuto] next send-mail scheduled in %.2f hours (Mon/Wed/Fri 08:00)", wait.Hours())

			timer := time.NewTimer(wait)
			<-timer.C

			if err := RunSendMailAuto(db); err != nil {
				log.Printf("[taskAuto] scheduled send-mail error: %v", err)
			}
		}
	}()
}

func TriggerDelayCheck(c *fiber.Ctx, db *sqlx.DB) error {
	if err := CheckAndUpdateDelayStatus(db); err != nil {
		log.Printf("[taskAuto] TriggerDelayCheck error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "delay status check completed",
	})
}
