package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// CheckAndUpdateDelayStatus finds every row in info_project_item_detail whose
// ipid_end_date has already passed (today > ipid_end_date, i.e. at least 1 day
// overdue) and whose status is not yet terminal (done / reject / delay), then
// flips ipid_status to 'delay'.
func CheckAndUpdateDelayStatus(db *sqlx.DB) error {
	query := `
		UPDATE info_project_item_detail
		SET
			ipid_status     = 'delay',
			ipid_updated_at = NOW(),
			ipid_updated_by = 'system'
		WHERE
			ipid_end_date IS NOT NULL
			AND ipid_end_date < CURDATE()
			AND ipid_status NOT IN ('done', 'reject', 'delay')
	`

	result, err := db.Exec(query)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	log.Printf("[taskAuto] CheckAndUpdateDelayStatus: updated %d row(s) to 'delay'", rows)
	return nil
}

// durationUntilMidnight returns the duration from now until the next 00:00:00.
func durationUntilMidnight() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
	return time.Until(next)
}

// StartDelayStatusScheduler runs CheckAndUpdateDelayStatus once immediately on
// startup, then waits until the next midnight (00:00:00) and repeats every day.
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

// durationUntilNextMailDay returns the duration from now until the next
// Monday, Wednesday, or Friday at the given hour:minute (24-h clock).
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

// StartSendMailScheduler sends the automated project tracking email every
// Monday, Wednesday, and Friday at 08:00.
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

// TriggerDelayCheck is an HTTP handler that lets an authorised caller run the
// delay-status check on demand.
//
// POST /apiTrackingSystem/task/triggerDelayCheck
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
