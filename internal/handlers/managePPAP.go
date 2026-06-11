package handlers

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// SysPPAPItem represents a row in mst_ppap_item
type SysPPAPItem struct {
	ID                int64            `db:"mpi_id" json:"mpi_id"`
	Name              utils.NullString `db:"mpi_name" json:"mpi_name"`
	Status            string           `db:"mpi_status" json:"mpi_status"`
	CreatedAt         *time.Time       `db:"mpi_created_at" json:"mpi_created_at"`
	CreatedBy         utils.NullString `db:"mpi_created_by" json:"mpi_created_by"`
	UpdatedAt         *time.Time       `db:"mpi_updated_at" json:"mpi_updated_at"`
	UpdatedBy         utils.NullString `db:"mpi_updated_by" json:"mpi_updated_by"`
	UpdateByFirstName utils.NullString `db:"mpi_updated_by_firstname" json:"mpi_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"mpi_updated_by_lastname" json:"mpi_updated_by_lastname"`
}

// htmlEscapeNullString safely returns escaped string for utils.NullString
func htmlEscapeNullString(ns utils.NullString) string {
	if !ns.Valid {
		return ""
	}
	return html.EscapeString(ns.String)
}

func ListPPAPItems(c *fiber.Ctx, db *sqlx.DB) error {
	var list []SysPPAPItem
	query := `SELECT mpi_id AS mpi_id,
					 mpi_name AS mpi_name,
					 mpi_status AS mpi_status,
					 mpi_updated_at AS mpi_updated_at,
					 mpi_updated_by AS mpi_updated_by,
					 su.su_firstname AS mpi_updated_by_firstname,
					 su.su_lastname AS mpi_updated_by_lastname
				FROM mst_ppap_item 
				LEFT JOIN sys_user su ON mpi_updated_by = su_emp_code
				ORDER BY mpi_id ASC`
	if err := db.Select(&list, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(list)
}

func InsertPPAPItem(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		Name      string `json:"mpi_name"`
		CreatedBy string `json:"mpi_created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_created_by is required"})
	}

	// duplicate check
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_ppap_item WHERE mpi_name = ?`, body.Name); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	if _, err := db.Exec(`INSERT INTO mst_ppap_item (mpi_name, mpi_status, mpi_created_at, mpi_created_by, mpi_updated_at, mpi_updated_by) VALUES (?, ?, ?, ?, ?, ?)`, body.Name, "active", now, body.CreatedBy, now, body.CreatedBy); err != nil {
		return c.Status(500).JSON(5)
	}
	return c.Status(201).JSON(1)
}

func UpdatePPAPItem(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mpi_id"`
		Name      string `json:"mpi_name"`
		UpdatedBy string `json:"mpi_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_id is required"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_name is required"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_updated_by is required"})
	}

	// duplicate check excluding current id
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_ppap_item WHERE mpi_name = ? AND mpi_id <> ?`, body.Name, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()
	res, err := db.Exec(`UPDATE mst_ppap_item SET mpi_name = ?, mpi_updated_at = ?, mpi_updated_by = ? WHERE mpi_id = ?`, body.Name, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "ppap item not found"})
	}
	return c.Status(200).JSON(1)
}

func UpdatePPAPItemStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"mpi_id"`
		Status    string `json:"mpi_status"`
		UpdatedBy string `json:"mpi_updated_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "mpi_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE mst_ppap_item SET mpi_status = ?, mpi_updated_at = ?, mpi_updated_by = ? WHERE mpi_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "ppap item not found"})
	}
	return c.Status(200).JSON(1)
}

func GetListPPAPItems(c *fiber.Ctx, db *sqlx.DB) error {
	var ppaps []struct {
		MpiID   int64            `db:"mpi_id" json:"mpi_id"`
		MpiName utils.NullString `db:"mpi_name" json:"mpi_name"`
	}
	query := `SELECT mpi_id, mpi_name FROM mst_ppap_item WHERE mpi_status = 'active' ORDER BY mpi_id ASC`
	if err := db.Select(&ppaps, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(ppaps)
}

func GetListPPAPStep4(c *fiber.Ctx, db *sqlx.DB) error {
	var ppaps []struct {
		MpiID         int64            `db:"mpi_id" json:"mpi_id"`
		MpiName       utils.NullString `db:"mpi_name" json:"mpi_name"`
		IpidStartDate *time.Time       `db:"ipid_start_date" json:"ipid_start_date"`
		IpidEndDate   *time.Time       `db:"ipid_end_date" json:"ipid_end_date"`
		SdID          utils.NullInt64  `db:"sd_id" json:"sd_id"`
		SuID          utils.NullInt64  `db:"su_id" json:"su_id"`
		IpidLineCode  utils.NullString `db:"ipid_line_code" json:"ipid_line_code"`
		Status        int              `db:"status" json:"status"`
	}

	// optional mmm_id filter (passed as query param)
	mmmIDStr := strings.TrimSpace(c.Query("mmm_id"))
	var mmmID int64 = 0
	if mmmIDStr != "" {
		if v, err := strconv.ParseInt(mmmIDStr, 10, 64); err == nil {
			mmmID = v
		}
	}

	// optional ip_id filter (passed as query param)
	ipIDStr := strings.TrimSpace(c.Query("ip_id"))
	var ipID int64 = 0
	if ipIDStr != "" {
		if v, err := strconv.ParseInt(ipIDStr, 10, 64); err == nil {
			ipID = v
		}
	}

	query := `SELECT
  mpi.mpi_id,
  mpi.mpi_name,
  ipid.ipid_start_date AS ipid_start_date,
  ipid.ipid_end_date AS ipid_end_date,
  ipid.sd_id AS sd_id,
  ipid.su_id AS su_id,
  ipid.ipid_line_code AS ipid_line_code,
    CASE
                WHEN ipid.ipid_id IS NOT NULL THEN 1
                    WHEN ipid.ipid_id IS NULL
                         AND NOT EXISTS (
                             SELECT 1
                             FROM info_ppap_item x
                             WHERE x.ip_id = ?
                         )
                         AND mpd.mpd_id IS NOT NULL
                    THEN 1
                ELSE 0
            END AS status
FROM
  mst_ppap_item mpi
  LEFT JOIN mst_ppap_detail mpd ON mpi.mpi_id = mpd.mpi_id
  AND mpd.mmm_id = ?
  AND mpd.mpd_status = 'active'
  LEFT JOIN info_ppap_item ipi ON mpi.mpi_name = ipi.ipi_name AND ipi.ip_id = ?
  LEFT JOIN info_project_item_detail ipid ON ipid.ref_id = ipi.ipi_id AND ipid.ipid_type = 'ppap'
WHERE
  mpi_status = 'active'
ORDER BY
  mpi.mpi_id ASC`

	if err := db.Select(&ppaps, query, ipID, mmmID, ipID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(ppaps)
}

func InsertPPAPItemStep4(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		IpID          int64         `json:"ip_id"`
		IpiName       []string      `json:"ipi_name"`
		SuID          []string      `json:"su_id"`
		IpidLineCode  []string      `json:"ipid_line_code"`
		IpidStartDate StringOrArray `json:"ipid_start_date"`
		IpidEndDate   StringOrArray `json:"ipid_end_date"`
		CreatedBy     string        `json:"created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.IpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "created_by is required"})
	}

	n := len(body.IpiName)
	if n > 0 {
		if len(body.SuID) != 0 && len(body.SuID) != n {
			return c.Status(400).JSON(fiber.Map{"error": "su_id length must match ipi_name length"})
		}
		if len(body.IpidLineCode) != 0 && len(body.IpidLineCode) != n {
			return c.Status(400).JSON(fiber.Map{"error": "ipid_line_code length must match ipi_name length"})
		}
	}

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now()

	parseCSVInts := func(s string) []int64 {
		s = strings.TrimSpace(s)
		if s == "" || s == "null" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]int64, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if id, err := strconv.ParseInt(p, 10, 64); err == nil {
				out = append(out, id)
			}
		}
		return out
	}

	parseCSVStrings := func(s string) []string {
		s = strings.TrimSpace(s)
		if s == "" || s == "null" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	type existingPPAPDetail struct {
		IpidID    int64          `db:"ipid_id"`
		Start     sql.NullTime   `db:"ipid_start_date"`
		End       sql.NullTime   `db:"ipid_end_date"`
		StatusFlg sql.NullString `db:"ipid_status_flg"`
	}

	shouldUpdateExistingPPAPDetail := func(existing existingPPAPDetail, parsedStart, parsedEnd *time.Time) bool {
		if existing.StatusFlg.Valid && existing.StatusFlg.String == "sendbyphase" {
			return false
		}
		needUpdate := false
		if parsedStart == nil && existing.Start.Valid {
			needUpdate = true
		}
		if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
			needUpdate = true
		}
		if !needUpdate {
			if parsedEnd == nil && existing.End.Valid {
				needUpdate = true
			}
			if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
				needUpdate = true
			}
		}
		if !needUpdate && (!existing.StatusFlg.Valid || existing.StatusFlg.String != "sendandsubmit") {
			needUpdate = true
		}
		return needUpdate
	}

	updateExistingPPAPDetail := func(ipidID int64, startDate, endDate interface{}) error {
		_, err := tx.Exec(`UPDATE info_project_item_detail SET
			ipid_start_date = ?,
			ipid_end_date = ?,
			ipid_updated_at = ?,
			ipid_updated_by = ?,
			ipid_status_flg = CASE
				WHEN ipid_status_flg = 'sendbyphase' THEN ipid_status_flg
				ELSE 'sendandsubmit'
			END
		WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID)
		return err
	}

	insertPPAPDetail := func(refID int64, sd, su, line, startDate, endDate interface{}) (int64, error) {
		res, err := tx.Exec(`INSERT INTO info_project_item_detail
			(ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_status_flg, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			refID, sd, su, line, "ppap", startDate, endDate, "inprogress", "sendandsubmit", now, body.CreatedBy, now, body.CreatedBy,
		)
		if err != nil {
			return 0, err
		}
		id, _ := res.LastInsertId()
		return id, nil
	}

	var results []map[string]int64

	if n > 0 {
		if !(len(body.IpidStartDate) == 0 || len(body.IpidStartDate) == 1 || len(body.IpidStartDate) == n) {
			return c.Status(400).JSON(fiber.Map{"error": "ipid_start_date must be empty, single value, or array matching ipi_name length"})
		}
		if !(len(body.IpidEndDate) == 0 || len(body.IpidEndDate) == 1 || len(body.IpidEndDate) == n) {
			return c.Status(400).JSON(fiber.Map{"error": "ipid_end_date must be empty, single value, or array matching ipi_name length"})
		}

		parsedStarts := make([]*time.Time, n)
		parsedEnds := make([]*time.Time, n)

		for i := 0; i < n; i++ {
			var s string
			if len(body.IpidStartDate) == 1 {
				s = body.IpidStartDate[0]
			} else if len(body.IpidStartDate) == n {
				s = body.IpidStartDate[i]
			}
			s = strings.TrimSpace(s)
			if s != "" && !strings.EqualFold(s, "null") {
				if t, err := time.Parse("2006-01-02", s); err == nil {
					parsedStarts[i] = &t
				} else if t, err := time.Parse(time.RFC3339, s); err == nil {
					parsedStarts[i] = &t
				} else {
					return c.Status(400).JSON(fiber.Map{"error": "invalid ipid_start_date format"})
				}
			}

			var e string
			if len(body.IpidEndDate) == 1 {
				e = body.IpidEndDate[0]
			} else if len(body.IpidEndDate) == n {
				e = body.IpidEndDate[i]
			}
			e = strings.TrimSpace(e)
			if e != "" && !strings.EqualFold(e, "null") {
				if t, err := time.Parse("2006-01-02", e); err == nil {
					parsedEnds[i] = &t
				} else if t, err := time.Parse(time.RFC3339, e); err == nil {
					parsedEnds[i] = &t
				} else {
					return c.Status(400).JSON(fiber.Map{"error": "invalid ipid_end_date format"})
				}
			}
		}

		for i := 0; i < n; i++ {
			name := strings.TrimSpace(body.IpiName[i])
			if name == "" {
				return c.Status(400).JSON(fiber.Map{"error": "ipi_name values must not be empty"})
			}

			parsedStart := parsedStarts[i]
			parsedEnd := parsedEnds[i]
			var startDate interface{} = nil
			var endDate interface{} = nil
			if parsedStart != nil {
				startDate = *parsedStart
			}
			if parsedEnd != nil {
				endDate = *parsedEnd
			}

			var ipiID int64
			err = tx.Get(&ipiID, `SELECT ipi_id FROM info_ppap_item WHERE ip_id = ? AND ipi_name = ?`, body.IpID, name)
			if err != nil {
				if err == sql.ErrNoRows {
					res, err := tx.Exec(`INSERT INTO info_ppap_item (ip_id, ipi_name, ipi_created_at, ipi_created_by) VALUES (?, ?, ?, ?)`, body.IpID, name, now, body.CreatedBy)
					if err != nil {
						return c.Status(500).JSON(5)
					}
					id, _ := res.LastInsertId()
					ipiID = id
				} else {
					return c.Status(500).JSON(5)
				}
			}

			suPerRow := len(body.SuID) == n
			globalSu := []int64{}
			if len(body.SuID) == 1 {
				globalSu = parseCSVInts(body.SuID[0])
			}
			suList := []int64{}
			if suPerRow {
				suList = parseCSVInts(body.SuID[i])
			}

			sdList := []int64{}
			if suPerRow && len(suList) > 0 {
				for _, su := range suList {
					var sd sql.NullInt64
					if err := tx.Get(&sd, `SELECT sd_id FROM sys_user WHERE su_id = ? LIMIT 1`, su); err != nil {
						if err == sql.ErrNoRows {
							continue
						}
						return c.Status(500).JSON(fiber.Map{"error": "failed to fetch sd_id for su_id", "detail": err.Error()})
					}
					if sd.Valid && sd.Int64 != 0 {
						sdList = append(sdList, sd.Int64)
					}
				}
			} else if len(globalSu) > 0 {
				for _, su := range globalSu {
					var sd sql.NullInt64
					if err := tx.Get(&sd, `SELECT sd_id FROM sys_user WHERE su_id = ? LIMIT 1`, su); err != nil {
						if err == sql.ErrNoRows {
							continue
						}
						return c.Status(500).JSON(fiber.Map{"error": "failed to fetch sd_id for su_id", "detail": err.Error()})
					}
					if sd.Valid && sd.Int64 != 0 {
						sdList = append(sdList, sd.Int64)
					}
				}
			}

			linePerRow := len(body.IpidLineCode) == n
			globalLine := []string{}
			if len(body.IpidLineCode) == 1 {
				globalLine = parseCSVStrings(body.IpidLineCode[0])
			}
			lineList := []string{}
			if linePerRow {
				lineList = parseCSVStrings(body.IpidLineCode[i])
			}

			if len(sdList) == 0 {
				if suPerRow && len(suList) > 0 {
					for _, su := range suList {
						var lineVal interface{} = nil
						if len(lineList) == 1 {
							if v, err := strconv.ParseInt(lineList[0], 10, 64); err == nil {
								lineVal = v
							}
						} else if len(globalLine) == 1 {
							if v, err := strconv.ParseInt(globalLine[0], 10, 64); err == nil {
								lineVal = v
							}
						}

						var existing existingPPAPDetail
						sel := `SELECT ipid_id, ipid_start_date, ipid_end_date, ipid_status_flg FROM info_project_item_detail WHERE ref_id = ? AND sd_id IS NULL AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
						err = tx.Get(&existing, sel, ipiID, su, su, lineVal, lineVal)

						var ipidID int64
						if err == nil {
							needUpdate := shouldUpdateExistingPPAPDetail(existing, parsedStart, parsedEnd)
							ipidID = existing.IpidID
							if needUpdate {
								if err := updateExistingPPAPDetail(ipidID, startDate, endDate); err != nil {
									return c.Status(500).JSON(5)
								}
							}
						} else {
							if err != sql.ErrNoRows {
								return c.Status(500).JSON(5)
							}
							ipidID, err = insertPPAPDetail(ipiID, nil, su, lineVal, startDate, endDate)
							if err != nil {
								return c.Status(500).JSON(5)
							}
						}
						results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
					}
					continue
				}

				if len(globalSu) > 0 {
					for _, su := range globalSu {
						var lineVal interface{} = nil
						if len(globalLine) == 1 {
							if v, err := strconv.ParseInt(globalLine[0], 10, 64); err == nil {
								lineVal = v
							}
						}

						var existing existingPPAPDetail
						sel := `SELECT ipid_id, ipid_start_date, ipid_end_date, ipid_status_flg FROM info_project_item_detail WHERE ref_id = ? AND sd_id IS NULL AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
						err = tx.Get(&existing, sel, ipiID, su, su, lineVal, lineVal)

						var ipidID int64
						if err == nil {
							needUpdate := shouldUpdateExistingPPAPDetail(existing, parsedStart, parsedEnd)
							ipidID = existing.IpidID
							if needUpdate {
								if err := updateExistingPPAPDetail(ipidID, startDate, endDate); err != nil {
									return c.Status(500).JSON(5)
								}
							}
						} else {
							if err != sql.ErrNoRows {
								return c.Status(500).JSON(5)
							}
							ipidID, err = insertPPAPDetail(ipiID, nil, su, lineVal, startDate, endDate)
							if err != nil {
								return c.Status(500).JSON(5)
							}
						}
						results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
					}
					continue
				}

				{
					var existing existingPPAPDetail
					sel := `SELECT ipid_id, ipid_start_date, ipid_end_date, ipid_status_flg FROM info_project_item_detail WHERE ref_id = ? AND sd_id IS NULL AND su_id IS NULL AND ipid_line_code IS NULL AND ipid_type = 'ppap' LIMIT 1`
					err = tx.Get(&existing, sel, ipiID)

					var ipidID int64
					if err == nil {
						needUpdate := shouldUpdateExistingPPAPDetail(existing, parsedStart, parsedEnd)
						ipidID = existing.IpidID
						if needUpdate {
							if err := updateExistingPPAPDetail(ipidID, startDate, endDate); err != nil {
								return c.Status(500).JSON(5)
							}
						}
					} else {
						if err != sql.ErrNoRows {
							return c.Status(500).JSON(5)
						}
						ipidID, err = insertPPAPDetail(ipiID, nil, nil, nil, startDate, endDate)
						if err != nil {
							return c.Status(500).JSON(5)
						}
					}
					results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
					continue
				}
			}

			for idx, sd := range sdList {
				var suVal interface{} = nil
				if suPerRow {
					if len(suList) == len(sdList) && idx < len(suList) {
						suVal = suList[idx]
					} else if len(suList) > 0 {
						suVal = suList[0]
					}
				} else if len(globalSu) > 0 {
					suVal = globalSu[0]
				}

				var lineVal interface{} = nil
				if linePerRow {
					if len(lineList) == len(sdList) && idx < len(lineList) {
						if v, err := strconv.ParseInt(lineList[idx], 10, 64); err == nil {
							lineVal = v
						}
					} else if len(lineList) > 0 {
						if v, err := strconv.ParseInt(lineList[0], 10, 64); err == nil {
							lineVal = v
						}
					}
				} else if len(globalLine) > 0 {
					if v, err := strconv.ParseInt(globalLine[0], 10, 64); err == nil {
						lineVal = v
					}
				}

				if suPerRow && len(suList) > 1 && len(sdList) == 1 {
					for _, su := range suList {
						var existing existingPPAPDetail
						sel := `SELECT ipid_id, ipid_start_date, ipid_end_date, ipid_status_flg FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
						err = tx.Get(&existing, sel, ipiID, sd, sd, su, su, lineVal, lineVal)

						var ipidID int64
						if err == nil {
							needUpdate := shouldUpdateExistingPPAPDetail(existing, parsedStart, parsedEnd)
							ipidID = existing.IpidID
							if needUpdate {
								if err := updateExistingPPAPDetail(ipidID, startDate, endDate); err != nil {
									return c.Status(500).JSON(5)
								}
							}
						} else {
							if err != sql.ErrNoRows {
								return c.Status(500).JSON(5)
							}
							ipidID, err = insertPPAPDetail(ipiID, sd, su, lineVal, startDate, endDate)
							if err != nil {
								return c.Status(500).JSON(5)
							}
						}
						results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
					}
					continue
				}

				if !suPerRow && len(globalSu) > 1 {
					for _, su := range globalSu {
						var existing existingPPAPDetail
						sel := `SELECT ipid_id, ipid_start_date, ipid_end_date, ipid_status_flg FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
						err = tx.Get(&existing, sel, ipiID, sd, sd, su, su, lineVal, lineVal)

						var ipidID int64
						if err == nil {
							needUpdate := shouldUpdateExistingPPAPDetail(existing, parsedStart, parsedEnd)
							ipidID = existing.IpidID
							if needUpdate {
								if err := updateExistingPPAPDetail(ipidID, startDate, endDate); err != nil {
									return c.Status(500).JSON(5)
								}
							}
						} else {
							if err != sql.ErrNoRows {
								return c.Status(500).JSON(5)
							}
							ipidID, err = insertPPAPDetail(ipiID, sd, su, lineVal, startDate, endDate)
							if err != nil {
								return c.Status(500).JSON(5)
							}
						}
						results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
					}
					continue
				}

				var existing existingPPAPDetail
				sel := `SELECT ipid_id, ipid_start_date, ipid_end_date, ipid_status_flg FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
				err = tx.Get(&existing, sel, ipiID, sd, sd, suVal, suVal, lineVal, lineVal)

				var ipidID int64
				if err == nil {
					needUpdate := shouldUpdateExistingPPAPDetail(existing, parsedStart, parsedEnd)
					ipidID = existing.IpidID
					if needUpdate {
						if err := updateExistingPPAPDetail(ipidID, startDate, endDate); err != nil {
							return c.Status(500).JSON(5)
						}
					}
				} else {
					if err != sql.ErrNoRows {
						return c.Status(500).JSON(5)
					}
					ipidID, err = insertPPAPDetail(ipiID, sd, suVal, lineVal, startDate, endDate)
					if err != nil {
						return c.Status(500).JSON(5)
					}
				}
				results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
			}
		}
	}

	if _, err := tx.Exec(`UPDATE info_project SET ip_status = ?, ip_updated_at = ?, ip_updated_by = ? WHERE ip_id = ?`, "inprogress", now, body.CreatedBy, body.IpID); err != nil {
		return c.Status(500).JSON(5)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(5)
	}

	// Send mail in background (same pattern as InsertAPQPByPhase)
	go func() {
		type mailRow struct {
			IpCode     sql.NullString `db:"ip_code"`
			IpPartName sql.NullString `db:"ip_part_name"`
			IpPartNo   sql.NullString `db:"ip_part_no"`
			IpModel    sql.NullString `db:"ip_model"`
			ItemName   sql.NullString `db:"item_name"`
			ItemType   sql.NullString `db:"item_type"`
			OwnerSuID  sql.NullInt64  `db:"owner_su_id"`
			StartDate  *time.Time     `db:"start_date"`
			EndDate    *time.Time     `db:"end_date"`
		}

		var allRows []mailRow
		if err := db.Select(&allRows, `
			SELECT
				ip.ip_code, ip.ip_part_name, ip.ip_part_no, ip.ip_model,
				x.item_name, x.item_type, x.owner_su_id, x.start_date, x.end_date
			FROM (
				SELECT ai.iai_name AS item_name, pid.ipid_type AS item_type,
				       pid.su_id AS owner_su_id,
				       pid.ipid_start_date AS start_date, pid.ipid_end_date AS end_date
				FROM info_project_item_detail pid
				JOIN info_apqp_item ai ON ai.iai_id = pid.ref_id
				WHERE pid.ipid_type = 'apqp' AND ai.ip_id = ? AND pid.ipid_status_flg = 'sendandsubmit'
				UNION ALL
				SELECT pi.ipi_name AS item_name, pid.ipid_type AS item_type,
				       pid.su_id AS owner_su_id,
				       pid.ipid_start_date AS start_date, pid.ipid_end_date AS end_date
				FROM info_project_item_detail pid
				JOIN info_ppap_item pi ON pi.ipi_id = pid.ref_id
				WHERE pid.ipid_type = 'ppap' AND pi.ip_id = ? AND pid.ipid_status_flg = 'sendandsubmit'
			) x
			LEFT JOIN info_project ip ON ip.ip_id = ?
			ORDER BY x.item_type ASC, x.start_date ASC, x.item_name ASC
		`, body.IpID, body.IpID, body.IpID); err != nil {
			log.Printf("InsertPPAPItemStep4 mail: query allRows error: %v", err)
			return
		}
		if len(allRows) == 0 {
			return
		}

		// Collect unique su_ids from allRows
		seenSu := map[int64]struct{}{}
		for _, r := range allRows {
			if r.OwnerSuID.Valid && r.OwnerSuID.Int64 != 0 {
				seenSu[r.OwnerSuID.Int64] = struct{}{}
			}
		}
		if len(seenSu) == 0 {
			return
		}

		subject := "TBKK Project Control Notification : waiting upload file"

		for suID := range seenSu {
			var email string
			if err := db.Get(&email, `SELECT su_email FROM sys_user WHERE su_id = ? AND su_status = 'active' AND su_email IS NOT NULL AND su_email <> '' LIMIT 1`, suID); err != nil {
				continue
			}
			toEmail := strings.TrimSpace(email)
			if toEmail == "" {
				continue
			}

			var firstname string
			var fn, ln sql.NullString
			_ = db.QueryRow(`SELECT su_firstname, su_lastname FROM sys_user WHERE su_id = ? LIMIT 1`, suID).Scan(&fn, &ln)
			if fn.Valid && fn.String != "" {
				firstname = fn.String
			} else {
				firstname = fmt.Sprintf("%v", suID)
			}

			// CC from sys_workflow by sd_id
			var sdID sql.NullInt64
			_ = db.Get(&sdID, `SELECT sd_id FROM sys_user WHERE su_id = ? LIMIT 1`, suID)
			var ccEmails []string
			if sdID.Valid {
				ccRows, err := db.Queryx(`
					SELECT su.su_email FROM sys_workflow sw
					JOIN sys_user su ON sw.su_id = su.su_id
					WHERE sw.sd_id = ? AND sw.sw_status = 'active'
					  AND su.su_status = 'active' AND su.su_email IS NOT NULL AND su.su_email <> ''
				`, sdID.Int64)
				if err == nil {
					for ccRows.Next() {
						var cc sql.NullString
						if err := ccRows.Scan(&cc); err == nil && cc.Valid {
							if e := strings.TrimSpace(cc.String); e != "" && !strings.EqualFold(e, toEmail) {
								ccEmails = append(ccEmails, e)
							}
						}
					}
					ccRows.Close()
				}
			}

			// Filter rows for this recipient
			var rows []mailRow
			for _, r := range allRows {
				if r.OwnerSuID.Valid && r.OwnerSuID.Int64 == suID {
					rows = append(rows, r)
				}
			}
			if len(rows) == 0 {
				continue
			}

			var sb strings.Builder
			sb.WriteString("<h3 style='font-family: Arial, sans-serif; color:#1f2d3d;'>Dear, K." + html.EscapeString(firstname) + "</h3>")
			sb.WriteString("<h4>You have new project please <b style='color:#0952c0;'>Upload file</b> in website Project Management</h4>")
			sb.WriteString("<html><body style='font-family: Arial, sans-serif; background:#f6f8fb; padding:20px;'>")
			sb.WriteString("<div style='margin:auto; background:#ffffff; border-radius:10px; border:1px solid #e0e6ed; padding:20px;'>")
			sb.WriteString("<div style='font-size:18px; font-weight:bold; color:#1f2d3d;'>Project Detail</div>")
			sb.WriteString("<div style='font-size:12px; color:#6b7280;'>PPAP Item Information</div>")
			sb.WriteString("<hr style='border:none; border-top:1px dashed #d1d5db; margin:15px 0;'>")
			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse;'><tr>")
			getStr := func(ns sql.NullString) string {
				if ns.Valid {
					return ns.String
				}
				return "-"
			}
			sb.WriteString("<td style='width:50%; border:1px solid #e5e7eb; border-radius:8px; padding:12px;'><div style='font-size:12px; color:#374151;'>PROJECT CODE</div><div style='font-size:14px; font-weight:bold; color:#2563eb;'>#" + html.EscapeString(getStr(rows[0].IpCode)) + "</div></td>")
			sb.WriteString("<td style='width:50%; border:1px solid #e5e7eb; border-radius:8px; padding:12px;'><div style='font-size:12px; color:#374151;'>MODEL</div><div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStr(rows[0].IpModel)) + "</div></td>")
			sb.WriteString("</tr><tr>")
			sb.WriteString("<td style='border:1px solid #e5e7eb; border-radius:8px; padding:12px;'><div style='font-size:12px; color:#374151;'>PART NO</div><div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStr(rows[0].IpPartNo)) + "</div></td>")
			sb.WriteString("<td style='border:1px solid #e5e7eb; border-radius:8px; padding:12px;'><div style='font-size:12px; color:#374151;'>PART NAME</div><div style='font-size:14px; font-weight:bold; color:#2563eb;'>" + html.EscapeString(getStr(rows[0].IpPartName)) + "</div></td>")
			sb.WriteString("</tr></table></div><br>")
			sb.WriteString("<table width='100%' cellpadding='10' cellspacing='0' style='border-collapse:collapse; border:1px solid #e5e7eb;'>")
			sb.WriteString("<thead><tr style='background:#f3f4f6; border-bottom:2px solid #d1d5db;'>")
			for _, col := range []string{"No.", "Item Name", "Item Type", "Start Date", "End Date"} {
				sb.WriteString("<th style='text-align:left; font-weight:bold; color:#374151; font-size:13px;'>" + html.EscapeString(col) + "</th>")
			}
			sb.WriteString("</tr></thead><tbody>")
			for idx, r := range rows {
				var startStr, endStr string
				if r.StartDate != nil {
					startStr = r.StartDate.Format("2006-01-02")
				}
				if r.EndDate != nil {
					endStr = r.EndDate.Format("2006-01-02")
				}
				rowBg := ""
				if idx%2 == 0 {
					rowBg = " style='background:#fbfdff;'"
				}
				itemName := ""
				if r.ItemName.Valid {
					itemName = r.ItemName.String
				}
				itemType := ""
				if r.ItemType.Valid {
					itemType = r.ItemType.String
				}
				sb.WriteString("<tr" + rowBg + ">")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + strconv.Itoa(idx+1) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(itemName) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(itemType) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(startStr) + "</td>")
				sb.WriteString("<td style='border-bottom:1px solid #e5e7eb; padding:10px; font-size:13px;'>" + html.EscapeString(endStr) + "</td>")
				sb.WriteString("</tr>")
			}
			sb.WriteString("</tbody></table>")
			sb.WriteString("<div style='margin-top:20px;'><a href='http://192.168.161.205:4009/login' style='display:inline-block; padding:10px 20px; background:#2563eb; color:#fff; text-decoration:none; border-radius:6px; font-weight:bold; font-size:14px;'>Open Project Management</a></div>")
			sb.WriteString("<div style='margin-top:30px; padding-top:20px; border-top:1px solid #e5e7eb; font-size:13px; color:#6b7280;'><p>Best Regards,<br><strong>System Service Department</strong></p></div>")
			sb.WriteString("</body></html>")

			if mailErr := SendMailWithCC([]string{toEmail}, ccEmails, subject, sb.String(), "text/html; charset=utf-8"); mailErr != nil {
				log.Printf("InsertPPAPItemStep4 mail: send error to %s: %v", toEmail, mailErr)
			}
		}
	}()

	return c.Status(201).JSON(1)
}

func InsertPPAPItemStep4Draft(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		IpID          int64         `json:"ip_id"`
		IpiName       []string      `json:"ipi_name"`
		SdID          []string      `json:"sd_id"`
		SuID          []string      `json:"su_id"`
		IpidLineCode  []string      `json:"ipid_line_code"`
		IpidStartDate StringOrArray `json:"ipid_start_date"`
		IpidEndDate   StringOrArray `json:"ipid_end_date"`
		CreatedBy     string        `json:"created_by"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.IpID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ip_id is required"})
	}
	if body.CreatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "created_by is required"})
	}
	// arrays must align with ipi_name length when PPAP items are provided
	n := len(body.IpiName)
	if n > 0 {
		if len(body.SdID) != 0 && len(body.SdID) != n {
			return c.Status(400).JSON(fiber.Map{"error": "sd_id length must match ipi_name length"})
		}
		if len(body.SuID) != 0 && len(body.SuID) != n {
			return c.Status(400).JSON(fiber.Map{"error": "su_id length must match ipi_name length"})
		}
		if len(body.IpidLineCode) != 0 && len(body.IpidLineCode) != n {
			return c.Status(400).JSON(fiber.Map{"error": "ipid_line_code length must match ipi_name length"})
		}
	}

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "database error", "detail": err.Error()})
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now()

	markAPQPNullStatusAsDraft := func() error {
		_, err := tx.Exec(`
			UPDATE info_project_item_detail ipid
			JOIN info_apqp_item iai ON iai.iai_id = ipid.ref_id
			SET ipid.ipid_status_flg = 'draft',
				ipid.ipid_updated_at = ?,
				ipid.ipid_updated_by = ?
			WHERE iai.ip_id = ?
			  AND ipid.ipid_type = 'apqp'
			  AND ipid.ipid_status_flg IS NULL
		`, now, body.CreatedBy, body.IpID)
		return err
	}

	if n == 0 {
		if err := markAPQPNullStatusAsDraft(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "update apqp draft status failed", "detail": err.Error()})
		}

		if _, err := tx.Exec(`UPDATE info_project SET ip_status = ?, ip_updated_at = ?, ip_updated_by = ? WHERE ip_id = ?`, "draft", now, body.CreatedBy, body.IpID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		if err := tx.Commit(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(1)
	}

	// Support ipid_start_date / ipid_end_date as 0, 1 or N values (N == len(ipi_name)).
	if !(len(body.IpidStartDate) == 0 || len(body.IpidStartDate) == 1 || len(body.IpidStartDate) == n) {
		return c.Status(400).JSON(fiber.Map{"error": "ipid_start_date must be empty, single value, or array matching ipi_name length"})
	}
	if !(len(body.IpidEndDate) == 0 || len(body.IpidEndDate) == 1 || len(body.IpidEndDate) == n) {
		return c.Status(400).JSON(fiber.Map{"error": "ipid_end_date must be empty, single value, or array matching ipi_name length"})
	}

	// pre-parse per-row dates into slices of *time.Time for easy comparison later
	parsedStarts := make([]*time.Time, n)
	parsedEnds := make([]*time.Time, n)
	for i := 0; i < n; i++ {
		// start
		var s string
		if len(body.IpidStartDate) == 1 {
			s = body.IpidStartDate[0]
		} else if len(body.IpidStartDate) == n {
			s = body.IpidStartDate[i]
		}
		s = strings.TrimSpace(s)
		if s != "" && !strings.EqualFold(s, "null") {
			if t, err := time.Parse("2006-01-02", s); err == nil {
				parsedStarts[i] = &t
			} else if t, err := time.Parse(time.RFC3339, s); err == nil {
				parsedStarts[i] = &t
			} else {
				return c.Status(400).JSON(fiber.Map{"error": "invalid ipid_start_date format"})
			}
		}

		// end
		var e string
		if len(body.IpidEndDate) == 1 {
			e = body.IpidEndDate[0]
		} else if len(body.IpidEndDate) == n {
			e = body.IpidEndDate[i]
		}
		e = strings.TrimSpace(e)
		if e != "" && !strings.EqualFold(e, "null") {
			if t, err := time.Parse("2006-01-02", e); err == nil {
				parsedEnds[i] = &t
			} else if t, err := time.Parse(time.RFC3339, e); err == nil {
				parsedEnds[i] = &t
			} else {
				return c.Status(400).JSON(fiber.Map{"error": "invalid ipid_end_date format"})
			}
		}
	}

	// helper: parse CSV string into ints
	parseCSVInts := func(s string) []int64 {
		s = strings.TrimSpace(s)
		if s == "" || s == "null" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]int64, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if id, err := strconv.ParseInt(p, 10, 64); err == nil {
				out = append(out, id)
			}
		}
		return out
	}

	parseCSVStrings := func(s string) []string {
		s = strings.TrimSpace(s)
		if s == "" || s == "null" {
			return nil
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	// results to return: per index, the ipi_id and created ipid_id
	var results []map[string]int64

	for i := 0; i < n; i++ {
		name := strings.TrimSpace(body.IpiName[i])
		if name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "ipi_name values must not be empty"})
		}

		// per-row parsed dates and interface values used for SQL params
		parsedStart := parsedStarts[i]
		parsedEnd := parsedEnds[i]
		var startDate interface{} = nil
		var endDate interface{} = nil
		if parsedStart != nil {
			startDate = *parsedStart
		}
		if parsedEnd != nil {
			endDate = *parsedEnd
		}

		var ipiID int64
		// check existing ipi for same ip_id and name
		err = tx.Get(&ipiID, `SELECT ipi_id FROM info_ppap_item WHERE ip_id = ? AND ipi_name = ?`, body.IpID, name)
		if err != nil {
			if err == sql.ErrNoRows {
				res, err := tx.Exec(`INSERT INTO info_ppap_item (ip_id, ipi_name, ipi_created_at, ipi_created_by) VALUES (?, ?, ?, ?)`, body.IpID, name, now, body.CreatedBy)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": err.Error()})
				}
				id, _ := res.LastInsertId()
				ipiID = id
			} else {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
		}

		// build lists per-row
		sdList := []int64{}
		if len(body.SdID) == n {
			sdList = parseCSVInts(body.SdID[i])
		}

		// su mapping: per-row or global
		suPerRow := len(body.SuID) == n
		globalSu := []int64{}
		if len(body.SuID) == 1 {
			globalSu = parseCSVInts(body.SuID[0])
		}
		suList := []int64{}
		if suPerRow {
			suList = parseCSVInts(body.SuID[i])
		}

		// line code mapping
		linePerRow := len(body.IpidLineCode) == n
		globalLine := []string{}
		if len(body.IpidLineCode) == 1 {
			globalLine = parseCSVStrings(body.IpidLineCode[0])
		}
		lineList := []string{}
		if linePerRow {
			lineList = parseCSVStrings(body.IpidLineCode[i])
		}

		// If no sdList but suList/globalSu provided, create rows with sd NULL
		if len(sdList) == 0 {
			// Use suPerRow -> insert for each su token, else if globalSu exists use globalSu tokens
			if suPerRow && len(suList) > 0 {
				for _, su := range suList {
					// line mapping: use first of lineList/globalLine or nil
					var lineVal interface{} = nil
					if len(lineList) == 1 {
						if v, err := strconv.ParseInt(lineList[0], 10, 64); err == nil {
							lineVal = v
						}
					} else if len(globalLine) == 1 {
						if v, err := strconv.ParseInt(globalLine[0], 10, 64); err == nil {
							lineVal = v
						}
					}
					// upsert detail for (ref_id, sd=NULL, su, lineVal)
					var existing struct {
						IpidID int64        `db:"ipid_id"`
						Start  sql.NullTime `db:"ipid_start_date"`
						End    sql.NullTime `db:"ipid_end_date"`
					}
					sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND sd_id IS NULL AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
					err = tx.Get(&existing, sel, ipiID, su, su, lineVal, lineVal)
					var ipidID int64
					if err == nil {
						needUpdate := false
						if parsedStart == nil && existing.Start.Valid {
							needUpdate = true
						}
						if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
							needUpdate = true
						}
						if !needUpdate {
							if parsedEnd == nil && existing.End.Valid {
								needUpdate = true
							}
							if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
								needUpdate = true
							}
						}
						ipidID = existing.IpidID
						if needUpdate {
							if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID); err != nil {
								return c.Status(500).JSON(fiber.Map{"error": err.Error()})
							}
						}
					} else {
						if err != sql.ErrNoRows {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						res2, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, ipiID, su, lineVal, "ppap", startDate, endDate, "inprogress", now, body.CreatedBy, now, body.CreatedBy)
						if err != nil {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						id, _ := res2.LastInsertId()
						ipidID = id
					}
					results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
				}
				// done sd==NULL case for this row
				continue
			}
			// if globalSu exists but sd empty, still create rows with NULL sd but no su assigned per your earlier logic -> skip here
			if len(globalSu) > 0 {
				for _, su := range globalSu {
					var lineVal interface{} = nil
					if len(globalLine) == 1 {
						if v, err := strconv.ParseInt(globalLine[0], 10, 64); err == nil {
							lineVal = v
						}
					}
					var existing struct {
						IpidID int64        `db:"ipid_id"`
						Start  sql.NullTime `db:"ipid_start_date"`
						End    sql.NullTime `db:"ipid_end_date"`
					}
					sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND sd_id IS NULL AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
					err = tx.Get(&existing, sel, ipiID, su, su, lineVal, lineVal)
					var ipidID int64
					if err == nil {
						needUpdate := false
						if parsedStart == nil && existing.Start.Valid {
							needUpdate = true
						}
						if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
							needUpdate = true
						}
						if !needUpdate {
							if parsedEnd == nil && existing.End.Valid {
								needUpdate = true
							}
							if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
								needUpdate = true
							}
						}
						ipidID = existing.IpidID
						if needUpdate {
							if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID); err != nil {
								return c.Status(500).JSON(fiber.Map{"error": err.Error()})
							}
						}
					} else {
						if err != sql.ErrNoRows {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						res2, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, ipiID, su, lineVal, "ppap", startDate, endDate, "inprogress", now, body.CreatedBy, now, body.CreatedBy)
						if err != nil {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						id, _ := res2.LastInsertId()
						ipidID = id
					}
					results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
				}
				continue
			}

			// fallback: create a single detail row with NULL sd, NULL su, NULL line
			{
				var existing struct {
					IpidID int64        `db:"ipid_id"`
					Start  sql.NullTime `db:"ipid_start_date"`
					End    sql.NullTime `db:"ipid_end_date"`
				}
				sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND sd_id IS NULL AND su_id IS NULL AND ipid_line_code IS NULL AND ipid_type = 'ppap' LIMIT 1`
				err = tx.Get(&existing, sel, ipiID)
				var ipidID int64
				if err == nil {
					needUpdate := false
					if parsedStart == nil && existing.Start.Valid {
						needUpdate = true
					}
					if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
						needUpdate = true
					}
					if !needUpdate {
						if parsedEnd == nil && existing.End.Valid {
							needUpdate = true
						}
						if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
							needUpdate = true
						}
					}
					ipidID = existing.IpidID
					if needUpdate {
						if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID); err != nil {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
					}
				} else {
					if err != sql.ErrNoRows {
						return c.Status(500).JSON(fiber.Map{"error": err.Error()})
					}
					res2, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, NULL, NULL, NULL, ?, ?, ?, ?, ?, ?, ?, ?)`, ipiID, "ppap", startDate, endDate, "inprogress", now, body.CreatedBy, now, body.CreatedBy)
					if err != nil {
						return c.Status(500).JSON(fiber.Map{"error": err.Error()})
					}
					id, _ := res2.LastInsertId()
					ipidID = id
				}
				results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
				continue
			}
		}

		// otherwise iterate sdList and map su/line per rules
		for idx, sd := range sdList {
			// determine su for this sd
			var suVal interface{} = nil
			if suPerRow {
				// if suList length equals sdList length, map by index
				if len(suList) == len(sdList) && idx < len(suList) {
					suVal = suList[idx]
				} else if len(suList) > 0 {
					// if suList has multiple tokens, expand into multiple inserts below
					// we'll handle expansion: if suList length>1 and len(sdList)==1 we create multiple rows
					suVal = suList[0]
				}
			} else if len(globalSu) > 0 {
				// use first of globalSu as default mapping; if len>1 we'll expand below
				suVal = globalSu[0]
			}

			// determine line for this sd
			var lineVal interface{} = nil
			if linePerRow {
				if len(lineList) == len(sdList) && idx < len(lineList) {
					if v, err := strconv.ParseInt(lineList[idx], 10, 64); err == nil {
						lineVal = v
					}
				} else if len(lineList) > 0 {
					if v, err := strconv.ParseInt(lineList[0], 10, 64); err == nil {
						lineVal = v
					}
				}
			} else if len(globalLine) > 0 {
				if v, err := strconv.ParseInt(globalLine[0], 10, 64); err == nil {
					lineVal = v
				}
			}

			// If suList (per-row) has multiple tokens and only one sd, expand
			if suPerRow && len(suList) > 1 && len(sdList) == 1 {
				for _, su := range suList {
					var existing struct {
						IpidID int64        `db:"ipid_id"`
						Start  sql.NullTime `db:"ipid_start_date"`
						End    sql.NullTime `db:"ipid_end_date"`
					}
					sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
					err = tx.Get(&existing, sel, ipiID, sd, sd, su, su, lineVal, lineVal)
					var ipidID int64
					if err == nil {
						needUpdate := false
						if parsedStart == nil && existing.Start.Valid {
							needUpdate = true
						}
						if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
							needUpdate = true
						}
						if !needUpdate {
							if parsedEnd == nil && existing.End.Valid {
								needUpdate = true
							}
							if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
								needUpdate = true
							}
						}
						ipidID = existing.IpidID
						if needUpdate {
							if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID); err != nil {
								return c.Status(500).JSON(fiber.Map{"error": err.Error()})
							}
						}
					} else {
						if err != sql.ErrNoRows {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						res2, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, ipiID, sd, su, lineVal, "ppap", startDate, endDate, "inprogress", now, body.CreatedBy, now, body.CreatedBy)
						if err != nil {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						id, _ := res2.LastInsertId()
						ipidID = id
					}
					results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
				}
				continue
			}

			// If globalSu has multiple tokens and sdList exists, create cross-product
			if !suPerRow && len(globalSu) > 1 {
				for _, su := range globalSu {
					var existing struct {
						IpidID int64        `db:"ipid_id"`
						Start  sql.NullTime `db:"ipid_start_date"`
						End    sql.NullTime `db:"ipid_end_date"`
					}
					sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
					err = tx.Get(&existing, sel, ipiID, sd, sd, su, su, lineVal, lineVal)
					var ipidID int64
					if err == nil {
						needUpdate := false
						if parsedStart == nil && existing.Start.Valid {
							needUpdate = true
						}
						if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
							needUpdate = true
						}
						if !needUpdate {
							if parsedEnd == nil && existing.End.Valid {
								needUpdate = true
							}
							if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
								needUpdate = true
							}
						}
						ipidID = existing.IpidID
						if needUpdate {
							if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID); err != nil {
								return c.Status(500).JSON(fiber.Map{"error": err.Error()})
							}
						}
					} else {
						if err != sql.ErrNoRows {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						res2, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, ipiID, sd, su, lineVal, "ppap", startDate, endDate, "inprogress", now, body.CreatedBy, now, body.CreatedBy)
						if err != nil {
							return c.Status(500).JSON(fiber.Map{"error": err.Error()})
						}
						id, _ := res2.LastInsertId()
						ipidID = id
					}
					results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
				}
				continue
			}

			// Default: single insert using sd and suVal (may be nil)
			var existing struct {
				IpidID int64        `db:"ipid_id"`
				Start  sql.NullTime `db:"ipid_start_date"`
				End    sql.NullTime `db:"ipid_end_date"`
			}
			sel := `SELECT ipid_id, ipid_start_date, ipid_end_date FROM info_project_item_detail WHERE ref_id = ? AND ((sd_id IS NULL AND ? IS NULL) OR sd_id = ?) AND ((su_id IS NULL AND ? IS NULL) OR su_id = ?) AND ((ipid_line_code IS NULL AND ? IS NULL) OR ipid_line_code = ?) AND ipid_type = 'ppap' LIMIT 1`
			err = tx.Get(&existing, sel, ipiID, sd, sd, suVal, suVal, lineVal, lineVal)
			var ipidID int64
			if err == nil {
				needUpdate := false
				if parsedStart == nil && existing.Start.Valid {
					needUpdate = true
				}
				if parsedStart != nil && (!existing.Start.Valid || !existing.Start.Time.Equal(*parsedStart)) {
					needUpdate = true
				}
				if !needUpdate {
					if parsedEnd == nil && existing.End.Valid {
						needUpdate = true
					}
					if parsedEnd != nil && (!existing.End.Valid || !existing.End.Time.Equal(*parsedEnd)) {
						needUpdate = true
					}
				}
				ipidID = existing.IpidID
				if needUpdate {
					if _, err := tx.Exec(`UPDATE info_project_item_detail SET ipid_start_date = ?, ipid_end_date = ?, ipid_updated_at = ?, ipid_updated_by = ? WHERE ipid_id = ?`, startDate, endDate, now, body.CreatedBy, ipidID); err != nil {
						return c.Status(500).JSON(fiber.Map{"error": err.Error()})
					}
				}
			} else {
				if err != sql.ErrNoRows {
					return c.Status(500).JSON(fiber.Map{"error": err.Error()})
				}
				res2, err := tx.Exec(`INSERT INTO info_project_item_detail (ref_id, sd_id, su_id, ipid_line_code, ipid_type, ipid_start_date, ipid_end_date, ipid_status, ipid_created_at, ipid_created_by, ipid_updated_at, ipid_updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, ipiID, sd, suVal, lineVal, "ppap", startDate, endDate, "inprogress", now, body.CreatedBy, now, body.CreatedBy)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": err.Error()})
				}
				id, _ := res2.LastInsertId()
				ipidID = id
			}
			results = append(results, map[string]int64{"ipi_id": ipiID, "ipid_id": ipidID})
		}
	}

	if err := markAPQPNullStatusAsDraft(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "update apqp draft status failed", "detail": err.Error()})
	}

	// mark project as draft when all inserts/updates succeeded
	if _, err := tx.Exec(`UPDATE info_project SET ip_status = ?, ip_updated_at = ?, ip_updated_by = ? WHERE ip_id = ?`, "draft", now, body.CreatedBy, body.IpID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(1)
}
