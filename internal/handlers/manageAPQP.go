package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"apiTrackingSystem/internal/utils"
)

type SysAPQP struct {
	ID                int64            `db:"ma_id" json:"ma_id"`
	MppID             int64            `db:"mpp_id" json:"mpp_id"`
	SdID              int64            `json:"sd_id"`
	Aname             utils.NullString `db:"aname" json:"aname"`
	Mppname           utils.NullString `db:"mpp_name" json:"mpp_name"`
	MppOrder          int64            `db:"mpp_order" json:"mpp_order"`
	Name              utils.NullString `db:"ma_name" json:"ma_name"`
	Status            string           `db:"ma_status" json:"ma_status"`
	CreatedAt         *time.Time       `db:"ma_created_at" json:"ma_created_at"`
	CreatedBy         utils.NullString `db:"ma_created_by" json:"ma_created_by"`
	UpdatedAt         *time.Time       `db:"ma_updated_at" json:"ma_updated_at"`
	UpdatedBy         utils.NullString `db:"ma_updated_by" json:"ma_updated_by"`
	UpdateByFirstName utils.NullString `db:"ma_updated_by_firstname" json:"ma_updated_by_firstname"`
	UpdateByLastName  utils.NullString `db:"ma_updated_by_lastname" json:"ma_updated_by_lastname"`
	Type              utils.NullString `db:"ma_type" json:"ma_type"`
}
type SysAPQPDetail struct {
	MaID      int64      `db:"ma_id" json:"ma_id"`
	SdID      int64      `db:"sd_id" json:"sd_id"`
	MadStatus string     `db:"mad_status" json:"mad_status"`
	UpdatedBy string     `db:"mad_updated_by" json:"mad_updated_by"`
	UpdatedAt *time.Time `db:"mad_updated_at" json:"mad_updated_at"`
	CreatedBy string     `db:"mad_created_by" json:"mad_created_by"`
	CreatedAt *time.Time `db:"mad_created_at" json:"mad_created_at"`
}

func ListAPQP(c *fiber.Ctx, db *sqlx.DB) error {
	mppID := c.Query("mpp_id")
	query := `SELECT
    mst_apqp.ma_id AS ma_id,
    mpp.mpp_id AS mpp_id,
    GROUP_CONCAT(DISTINCT sd.sd_dept_aname SEPARATOR ', ') AS aname,
    mpp.mpp_name AS mpp_name,
    mpp.mpp_order AS mpp_order,
    ma_name AS ma_name,
    ma_status AS ma_status,
    ma_updated_at AS ma_updated_at,
    ma_updated_by AS ma_updated_by,
    ma_type AS ma_type,
    su.su_firstname AS ma_updated_by_firstname,
    su.su_lastname AS ma_updated_by_lastname
  FROM mst_apqp
  LEFT JOIN sys_user su ON ma_updated_by = su_emp_code
  LEFT JOIN mst_project_phase mpp ON mpp.mpp_id = mst_apqp.mpp_id
  RIGHT JOIN mst_apqp_detail mad ON mad.ma_id = mst_apqp.ma_id
  LEFT JOIN sys_department sd ON sd.sd_id = mad.sd_id
  WHERE 1=1`

	args := []interface{}{}
	if mppID != "" {
		query += " AND mpp_id = ?"
		args = append(args, mppID)
	}
	query += " AND mad.mad_status = 'active' GROUP BY mst_apqp.ma_id ORDER BY ma_id ASC"

	var apqps []SysAPQP
	var err error
	err = db.Select(&apqps, query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(apqps)
}

func GetAPQP(c *fiber.Ctx, db *sqlx.DB) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id required"})
	}
	var a SysAPQP
	query := `SELECT ma_id AS ma_id, mpp_id AS mpp_id, ma_name AS ma_name, ma_status AS ma_status, ma_created_at AS ma_created_at, ma_created_by AS ma_created_by, ma_updated_at AS ma_updated_at, ma_updated_by AS ma_updated_by, ma_type AS ma_type FROM mst_apqp WHERE ma_id = ? LIMIT 1`
	if err := db.Get(&a, query, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(a)
}

func InsertAPQP(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		MppID     int64   `json:"mpp_id"`
		Name      string  `json:"ma_name"`
		CreatedBy string  `json:"ma_created_by"`
		SdIDs     []int64 `json:"sd_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_apqp WHERE ma_name = ? AND mpp_id = ?`, body.Name, body.MppID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}

	res, err := tx.Exec(`INSERT INTO mst_apqp (mpp_id, ma_name, ma_status, ma_created_at, ma_created_by, ma_updated_at, ma_updated_by, ma_type) VALUES (?, ?, 'active', ?, ?, ?, ?, ?)`, body.MppID, body.Name, now, body.CreatedBy, now, body.CreatedBy, "text")
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "cannot get last insert id", "detail": err.Error()})
	}

	if len(body.SdIDs) > 0 {
		stmt, err := tx.Prepare(`INSERT INTO mst_apqp_detail (ma_id, sd_id, mad_status, mad_created_at, mad_created_by, mad_updated_at, mad_updated_by) VALUES (?, ?, 'active', ?, ?, ?, ?)`)
		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(5)
		}
		defer stmt.Close()
		for _, sd := range body.SdIDs {
			if _, err := stmt.Exec(lastID, sd, now, body.CreatedBy, now, body.CreatedBy); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(5)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}

	return c.Status(201).JSON(1)
}

func UpdateAPQP(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64   `json:"ma_id"`
		MppID     int64   `json:"mpp_id"`
		Name      string  `json:"ma_name"`
		UpdatedBy string  `json:"ma_updated_by"`
		SdIDs     []int64 `json:"sd_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ma_id is required"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_name is required"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_updated_by is required"})
	}

	// check duplicate name within same master plan excluding current id
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM mst_apqp WHERE ma_name = ? AND mpp_id = ? AND ma_id <> ?`, body.Name, body.MppID, body.ID); err != nil {
		return c.Status(500).JSON(5)
	}
	if count > 0 {
		return c.Status(200).JSON(2)
	}

	now := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return c.Status(500).JSON(5)
	}

	// update master
	if _, err := tx.Exec(`UPDATE mst_apqp SET mpp_id = ?, ma_name = ?, ma_type = ?, ma_updated_at = ?, ma_updated_by = ? WHERE ma_id = ?`, body.MppID, body.Name, "text", now, body.UpdatedBy, body.ID); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}

	var existing []int64
	if err := tx.Select(&existing, `SELECT sd_id FROM mst_apqp_detail WHERE ma_id = ?`, body.ID); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(2)
	}

	existMap := make(map[int64]bool, len(existing))
	for _, s := range existing {
		existMap[s] = true
	}

	incomingMap := make(map[int64]bool, len(body.SdIDs))
	for _, sd := range body.SdIDs {
		incomingMap[sd] = true
		if existMap[sd] {
			if _, err := tx.Exec(`UPDATE mst_apqp_detail SET mad_status = 'active', mad_updated_at = ?, mad_updated_by = ? WHERE ma_id = ? AND sd_id = ?`, now, body.UpdatedBy, body.ID, sd); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(5)
			}
		} else {
			if _, err := tx.Exec(`INSERT INTO mst_apqp_detail (ma_id, sd_id, mad_status, mad_created_at, mad_created_by, mad_updated_at, mad_updated_by) VALUES (?, ?, 'active', ?, ?, ?, ?)`, body.ID, sd, now, body.UpdatedBy, now, body.UpdatedBy); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(5)
			}
		}
	}

	// inactivate details that are not present in incoming array
	for _, s := range existing {
		if !incomingMap[s] {
			if _, err := tx.Exec(`UPDATE mst_apqp_detail SET mad_status = 'inactive', mad_updated_at = ?, mad_updated_by = ? WHERE ma_id = ? AND sd_id = ?`, now, body.UpdatedBy, body.ID, s); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(5)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(5)
	}

	return c.Status(200).JSON(1)
}

func UpdateAPQPStatus(c *fiber.Ctx, db *sqlx.DB) error {
	var body struct {
		ID        int64  `json:"ma_id"`
		Status    string `json:"ma_status"`
		UpdatedBy string `json:"ma_updated_by"`
		SdID      int64  `json:"sd_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body", "detail": err.Error()})
	}
	if body.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "ma_id is required"})
	}
	if body.Status != "active" && body.Status != "inactive" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_status must be 'active' or 'inactive'"})
	}
	if body.UpdatedBy == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ma_updated_by is required"})
	}
	now := time.Now()
	res, err := db.Exec(`UPDATE mst_apqp SET ma_status = ?, ma_updated_at = ?, ma_updated_by = ? WHERE ma_id = ?`, body.Status, now, body.UpdatedBy, body.ID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "apqp not found"})
	}
	return c.Status(200).JSON(1)
}

func SelectPhaseAPQP(c *fiber.Ctx, db *sqlx.DB) error {
	var phases []struct {
		MppID     int64            `db:"mpp_id" json:"mpp_id"`
		PhaseName utils.NullString `db:"mpp_name" json:"mpp_name"`
	}
	query := `SELECT mpp_id, mpp_name FROM mst_project_phase WHERE mpp_status = 'active' ORDER BY mpp_id ASC`
	if err := db.Select(&phases, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(phases)
}

func SelectAPQPS(c *fiber.Ctx, db *sqlx.DB) error {
	var apqps []struct {
		MaID   int64            `db:"ma_id" json:"ma_id"`
		MppID  int64            `db:"mpp_id" json:"mpp_id"`
		MaName utils.NullString `db:"ma_name" json:"ma_name"`
	}
	query := `SELECT ma_id, mpp_id, ma_name FROM mst_apqp WHERE ma_status = 'active' ORDER BY mpp_id ASC`
	if err := db.Select(&apqps, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(apqps)
}

func GetListAPQPPhase(c *fiber.Ctx, db *sqlx.DB) error {
	var ids []int64
	var ipID int64
	if v := c.Query("ip_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			ipID = parsed
		}
	} else {
		return c.Status(400).JSON(fiber.Map{
			"error": "ip_id query parameter is required",
		})
	}
	// 1) รับจาก query string
	// รองรับ: ?mmp_id=1,2,3
	if q := c.Query("mmp_id"); q != "" {
		parts := strings.Split(q, ",")
		for _, p := range parts {
			if v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
				ids = append(ids, v)
			}
		}
	}

	// รองรับ: ?mmp_id=1&mmp_id=2
	if len(ids) == 0 {
		qs := c.Queries()["mmp_id"]
		if qs != "" {
			for _, p := range strings.Split(qs, ",") {
				if v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
					ids = append(ids, v)
				}
			}
		}
	}

	if len(ids) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "mmp_id query parameter is required",
		})
	}

	query := `
		SELECT
    COALESCE(iai.ip_id, ?) AS ip_id,
    mmp.mmp_id,
    ma.mpp_id AS mpp_id,
    ma.ma_id AS ma_id,
    mmpd.ma_id AS mmpd_ma_id,
    ma.ma_name AS ma_name,
    ipid.ipid_start_date AS ipid_start_date,
    ipid.ipid_end_date AS ipid_end_date,
    ipid.sd_id AS sd_id,
    ipid.su_id AS su_id,
    ipid.ipid_line_code AS ipid_line_code,
    ipmp.ipmp_start_date,
    ipmp.ipmp_end_date,
    CASE
        WHEN EXISTS (
            SELECT 1
            FROM mst_apqp ma2
            LEFT JOIN info_apqp_item iai2
                ON iai2.iai_name = ma2.ma_name
                AND iai2.ip_id = ?
            LEFT JOIN info_project_item_detail ipid2
                ON ipid2.ref_id = iai2.iai_id
                AND ipid2.ipid_type = 'apqp'
            LEFT JOIN mst_master_plan_detail mmpd2
                ON mmpd2.ma_id = ma2.ma_id
                AND mmpd2.mmpd_status = 'active'
            WHERE mmpd2.mmp_id = mmpd.mmp_id
              AND ipid2.ipid_status_flg = 'sendbyphase'
        )
        AND COALESCE(ipid.ipid_status_flg, '') <> 'sendbyphase'
        THEN 0

        WHEN COALESCE(ipid.ipid_status_flg, '') = 'cancel' THEN 0

        WHEN EXISTS (
            SELECT 1
            FROM mst_apqp ma5
            LEFT JOIN info_apqp_item iai5
                ON iai5.iai_name = ma5.ma_name
                AND iai5.ip_id = ?
            LEFT JOIN info_project_item_detail ipid5
                ON ipid5.ref_id = iai5.iai_id
                AND ipid5.ipid_type = 'apqp'
            LEFT JOIN mst_master_plan_detail mmpd5
                ON mmpd5.ma_id = ma5.ma_id
                AND mmpd5.mmpd_status = 'active'
            WHERE mmpd5.mmp_id = mmpd.mmp_id
              AND ipid5.ipid_status_flg = 'sendandsubmit'
        )
        AND COALESCE(ipid.ipid_status_flg, '') <> 'sendandsubmit'
        THEN 0

        WHEN ipid.ipid_id IS NOT NULL THEN 1

        WHEN ipid.ipid_id IS NULL
             AND NOT EXISTS (
                 SELECT 1
                 FROM info_apqp_item x
                 WHERE x.ip_id = ?
             )
             AND mtd.mmp_id IS NOT NULL
             AND ipmp.ipmp_id IS NOT NULL
        THEN 1

        ELSE 0
    END AS status,

CASE
    WHEN COALESCE(ipid.ipid_status_flg, '') = 'cancel' THEN 0

    WHEN EXISTS (
    SELECT 1
    FROM mst_apqp ma6
    LEFT JOIN info_apqp_item iai6
        ON iai6.iai_name = ma6.ma_name
        AND iai6.ip_id = ?
    LEFT JOIN info_project_item_detail ipid6
        ON ipid6.ref_id = iai6.iai_id
        AND ipid6.ipid_type = 'apqp'
    LEFT JOIN mst_master_plan_detail mmpd6
        ON mmpd6.ma_id = ma6.ma_id
        AND mmpd6.mmpd_status = 'active'
    LEFT JOIN mst_master_plan mmp6
        ON mmp6.mmp_id = mmpd6.mmp_id
    LEFT JOIN info_project_master_plan ipmp6
        ON ipmp6.ipmp_name = mmp6.mmp_name
        AND ipmp6.ip_id = ?
    WHERE ma6.mpp_id = ma.mpp_id
      AND ipmp6.ipmp_order_no = ipmp.ipmp_order_no
      AND ipid6.ipid_status_flg = 'sendbyphase'
)
THEN 0

WHEN EXISTS (
    SELECT 1
    FROM mst_apqp ma7
    LEFT JOIN info_apqp_item iai7
        ON iai7.iai_name = ma7.ma_name
        AND iai7.ip_id = ?
    LEFT JOIN info_project_item_detail ipid7
        ON ipid7.ref_id = iai7.iai_id
        AND ipid7.ipid_type = 'apqp'
    LEFT JOIN mst_master_plan_detail mmpd7
        ON mmpd7.ma_id = ma7.ma_id
        AND mmpd7.mmpd_status = 'active'
    LEFT JOIN mst_master_plan mmp7
        ON mmp7.mmp_id = mmpd7.mmp_id
    LEFT JOIN info_project_master_plan ipmp7
        ON ipmp7.ipmp_name = mmp7.mmp_name
        AND ipmp7.ip_id = ?
    WHERE ipmp7.ipmp_order_no = ipmp.ipmp_order_no
      AND ipid7.ipid_status_flg = 'draft'
)
THEN
    CASE
        WHEN ipid.ipid_status_flg = 'draft' THEN 1
        ELSE 0
    END

    WHEN ipmp.ipmp_id IS NOT NULL
         AND ipid.ipid_status_flg IS NULL THEN 1

    WHEN ipid.ipid_status_flg = 'draft' THEN 1

    ELSE 0
END AS status2,

    ipid.ipid_status_flg

FROM mst_apqp ma
LEFT JOIN info_apqp_item iai
    ON iai.iai_name = ma.ma_name
    AND iai.ip_id = ?
LEFT JOIN info_project_item_detail ipid
    ON ipid.ref_id = iai.iai_id
    AND ipid.ipid_type = 'apqp'
LEFT JOIN mst_master_plan_detail mmpd
    ON mmpd.ma_id = ma.ma_id
    AND mmpd.mmpd_status = 'active'
LEFT JOIN mst_template_detail mtd
    ON mmpd.mmp_id = mtd.mmp_id
    AND mmpd.mmp_id IN (?)
    AND mtd.mt_id = (SELECT mt_id FROM info_project WHERE ip_id = ?)
LEFT JOIN mst_master_plan mmp
    ON mmpd.mmp_id = mmp.mmp_id
LEFT JOIN info_project_master_plan ipmp
    ON ipmp.ipmp_name = mmp.mmp_name
    AND ipmp.ip_id = ?
WHERE ma.ma_status = 'active'
GROUP BY
    mmpd.ma_id,
    ma.ma_name,
    ma.mpp_id,
    ma.ma_id,
    ipid.ipid_start_date,
    ipid.ipid_end_date,
    ipid.sd_id,
    ipid.su_id,
    ipid.ipid_line_code,
    ipid.ipid_status_flg,
    ipmp.ipmp_start_date,
    ipmp.ipmp_end_date,
    ipid.ipid_id,
    ipmp.ipmp_id,
    mtd.mmp_id,
    iai.ip_id,
    mmp.mmp_id
ORDER BY ma.mpp_id, ma.ma_id ASC;
	`
	q, args, err := sqlx.In(query, ipID, ipID, ipID, ipID, ipID, ipID, ipID, ipID, ipID, ids, ipID, ipID)
	if err != nil {
		return c.Status(500).JSON(5)
	}
	q = db.Rebind(q)

	var list []struct {
		MmpID         utils.NullInt64  `db:"mmp_id" json:"mmp_id"`
		IpID          int64            `db:"ip_id" json:"ip_id"`
		MppID         int64            `db:"mpp_id" json:"mpp_id"`
		MaID          int64            `db:"ma_id" json:"ma_id"`
		MmpdMaID      utils.NullInt64  `db:"mmpd_ma_id" json:"mmpd_ma_id"`
		MaName        utils.NullString `db:"ma_name" json:"ma_name"`
		IpidStartDate *time.Time       `db:"ipid_start_date" json:"ipid_start_date"`
		IpidEndDate   *time.Time       `db:"ipid_end_date" json:"ipid_end_date"`
		SdID          utils.NullInt64  `db:"sd_id" json:"sd_id"`
		SuID          utils.NullInt64  `db:"su_id" json:"su_id"`
		IpidLineCode  utils.NullString `db:"ipid_line_code" json:"ipid_line_code"`
		IpmpStartDate *time.Time       `db:"ipmp_start_date" json:"ipmp_start_date"`
		IpmpEndDate   *time.Time       `db:"ipmp_end_date" json:"ipmp_end_date"`
		Status        int              `db:"status" json:"status"`
		Status2       utils.NullInt64  `db:"status2" json:"status2"`
		StatusFlag    utils.NullString `db:"ipid_status_flg" json:"ipid_status_flg"`
	}

	if err := db.Select(&list, q, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "query error",
			"detail": err.Error(),
		})
	}

	return c.JSON(list)
}
