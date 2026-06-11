package handlers

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"
)

const testMailRecipient = "noraphat_j@tbkk.co.th"

// RunSendMailNoraphatTest runs the same workbook-generation flow as SendMailAuto,
// but always sends to the fixed test recipient.
func RunSendMailNoraphatTest(db *sqlx.DB) error {
	sqlQuery := `SELECT
    ip.ip_code,
	x.mmp_id,
    ip.ip_part_name,
    ip.ip_part_no,
    ip.ip_model,
    x.item_name,
    x.item_type,
    x.start_date,
    x.end_date,
    x.ipid_status,
    x.ia_type,
    x.ia_status,
    x.ap_id,
    GROUP_CONCAT(DISTINCT su.su_id ORDER BY su.su_id SEPARATOR '/') AS owner_su_ids,
    GROUP_CONCAT(
        DISTINCT CONCAT('K.', su.su_firstname)
        ORDER BY su.su_firstname
        SEPARATOR '/'
    ) AS owner_names,
    GROUP_CONCAT(
        DISTINCT su.su_emp_code
        ORDER BY su.su_emp_code
        SEPARATOR '/'
    ) AS owner_emp_codes
FROM
(
    SELECT
        ai.ip_id AS ip_id,
        pid.ref_id AS ref_id,
        ai.iai_name AS item_name,
		ai.mpp_id AS mmp_id,
		ai.iai_id AS ap_id,
        pid.ipid_type AS item_type,
        pid.su_id AS owner_su_id,
        pid.ipid_start_date AS start_date,
        pid.ipid_end_date AS end_date,
        pid.ipid_id AS ipid_id,
        pid.ipid_status AS ipid_status,
        ia.ia_type AS ia_type,
        ia.ia_status AS ia_status
    FROM info_project_item_detail pid
    JOIN info_apqp_item ai
        ON ai.iai_id = pid.ref_id
       AND pid.ipid_type = 'apqp'
    LEFT JOIN info_approval ia
        ON ia.ipid_id = pid.ipid_id
       AND ia.ia_is_action = 1

    UNION ALL

    SELECT
        pi.ip_id AS ip_id,
        pid.ref_id AS ref_id,
        pi.ipi_name AS item_name,
		6 AS mmp_id,
		pi.ipi_id AS ap_id,
        pid.ipid_type AS item_type,
        pid.su_id AS owner_su_id,
        pid.ipid_start_date AS start_date,
        pid.ipid_end_date AS end_date,
        pid.ipid_id AS ipid_id,
        pid.ipid_status AS ipid_status,
        ia.ia_type AS ia_type,
        ia.ia_status AS ia_status
    FROM info_project_item_detail pid
    JOIN info_ppap_item pi
        ON pi.ipi_id = pid.ref_id
       AND pid.ipid_type = 'ppap'
    LEFT JOIN info_approval ia
        ON ia.ipid_id = pid.ipid_id
       AND ia.ia_is_action = 1
) x
LEFT JOIN sys_user su
    ON su.su_id = x.owner_su_id
LEFT JOIN info_project ip
    ON x.ip_id = ip.ip_id
WHERE x.ipid_status IN ('waiting', 'delay', 'inprogress', 'done', 'reject')

AND NOT (
    x.ipid_status = 'inprogress'
    AND EXISTS (
        SELECT 1
        FROM info_project_item_detail pid2
        WHERE pid2.ref_id = x.ref_id
          AND pid2.ipid_type = x.item_type
          AND pid2.ipid_status IN ('waiting', 'delay', 'done', 'reject')
    )
)

GROUP BY
    ip.ip_code,
    ip.ip_part_name,
    ip.ip_part_no,
    ip.ip_model,
	x.mmp_id,
    x.item_name,
    x.item_type,
    x.start_date,
    x.end_date,
    x.ipid_status,
    x.ia_type,
    x.ia_status
ORDER BY
	 x.mmp_id ASC,
	 x.ap_id ASC;`

	var rows []MailData
	if err := db.Select(&rows, sqlQuery); err != nil {
		return fmt.Errorf("failed to query data: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	reSheet := regexp.MustCompile(`[\\/:*?\[\]]`)
	sanitize := func(s string) string {
		s = reSheet.ReplaceAllString(s, "_")
		if s == "" {
			s = "sheet"
		}
		if len(s) > 31 {
			s = s[:31]
		}
		return s
	}

	f := excelize.NewFile()

	baseStyle, titleStyle, headerStyle, projectHeaderStyle, stDone, stDelay, stInprog, stWaiting, stReject, err := buildStyles(f)
	if err != nil {
		return fmt.Errorf("failed to build styles: %w", err)
	}

	modelGroups := map[string]map[string][]MailData{}
	for _, it := range rows {
		m := getStringValue(it.IpModel)
		ipc := getStringValue(it.ProjectCode)
		if _, ok := modelGroups[m]; !ok {
			modelGroups[m] = map[string][]MailData{}
		}
		modelGroups[m][ipc] = append(modelGroups[m][ipc], it)
	}

	models := make([]string, 0, len(modelGroups))
	for m := range modelGroups {
		models = append(models, m)
	}
	sort.Strings(models)

	firstSheet := true
	for _, model := range models {
		sheetName := sanitize(model)

		if firstSheet {
			_ = f.SetSheetName("Sheet1", sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		applySheetSetup(f, sheetName)
		buildTitleBlock(f, sheetName, titleStyle)

		currentRow := 5
		ipcodes := make([]string, 0, len(modelGroups[model]))
		for ip := range modelGroups[model] {
			ipcodes = append(ipcodes, ip)
		}
		sort.Strings(ipcodes)

		for _, ipcode := range ipcodes {
			group := modelGroups[model][ipcode]
			if len(group) == 0 {
				continue
			}
			first := group[0]

			currentRow = writeProjectBlock(
				f, sheetName, currentRow,
				getStringValue(first.ProjectCode),
				getStringValue(first.PartNo),
				getStringValue(first.PartName),
				getStringValue(first.IpModel),
				"-",
				group,
				baseStyle, headerStyle, projectHeaderStyle,
				stDone, stDelay, stInprog, stWaiting, stReject,
			)
		}
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return fmt.Errorf("failed to write excel file: %w", err)
	}

	subject := "TBKK Project Management Tracking Status"
	attachName := "TrackingProjects.xlsx"
	body := `<html>
<body style="margin:0;padding:0;font-family:Arial,sans-serif;background-color:#f5f7fa;color:#1f2937;">
	<div style="max-width:600px;margin:0 auto;padding:32px 24px;">
		<div style="background-color:#ffffff;border-radius:12px;padding:32px;box-shadow:0 4px 16px rgba(15,23,42,0.08);">
			<p style="margin:0 0 16px 0;font-size:16px;line-height:24px;">Please review the attached file and kindly proceed with the assigned action items.</p>
			<a href="http://192.168.161.205:4009/login" style="display:inline-block;padding:12px 24px;background-color:#0d6efd;color:#ffffff;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;">Open Project Tracking</a>
		</div>
	</div>
</body>
</html>`
	if err := SendMailWithAttachment([]string{testMailRecipient}, subject, body, "text/html; charset=utf-8", attachName, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to send test email: %w", err)
	}

	return nil
}

// SendMailNoraphatTest sends a standalone test email with the same workbook logic as SendMailAuto.
func SendMailNoraphatTest(c *fiber.Ctx, db *sqlx.DB) error {
	if err := RunSendMailNoraphatTest(db); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success":   false,
			"recipient": testMailRecipient,
			"error":     err.Error(),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"success":   true,
		"recipient": testMailRecipient,
		"message":   "Test tracking email sent",
	})
}
