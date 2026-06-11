package handlers

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"regexp"
	"sort"
	"strings"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

// MailData represents email data for project items
type MailData struct {
	ProjectCode   sql.NullString `db:"ip_code"`
	MmpID         sql.NullInt64  `db:"mmp_id"`
	ApID          sql.NullInt64  `db:"ap_id"`
	PartName      sql.NullString `db:"ip_part_name"`
	PartNo        sql.NullString `db:"ip_part_no"`
	IpModel       sql.NullString `db:"ip_model"`
	ItemName      sql.NullString `db:"item_name"`
	ItemType      sql.NullString `db:"item_type"`
	StartDate     sql.NullTime   `db:"start_date"`
	EndDate       sql.NullTime   `db:"end_date"`
	IpidStatus    sql.NullString `db:"ipid_status"`
	IaType        sql.NullString `db:"ia_type"`
	IaStatus      sql.NullString `db:"ia_status"`
	OwnerSuIDs    sql.NullString `db:"owner_su_ids"`
	OwnerNames    sql.NullString `db:"owner_names"`
	OwnerEmpCodes sql.NullString `db:"owner_emp_codes"`
}

func SendMail(to []string, subject, body, contentType string) error {
	_ = godotenv.Load(".env")
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if from == "" {
		from = user
	}
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}

	if host == "" || port == "" || user == "" || pass == "" {
		return fmt.Errorf("smtp configuration incomplete: host=%q port=%q user=%q from=%q", host, port, user, from)
	}

	addr := host + ":" + port
	hdr := make(map[string]string)
	hdr["From"] = from
	hdr["To"] = strings.Join(to, ", ")
	hdr["Subject"] = subject
	hdr["MIME-Version"] = "1.0"
	hdr["Content-Type"] = contentType

	var msg strings.Builder
	for k, v := range hdr {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Dial and create client so we can STARTTLS before AUTH
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial smtp server: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	// STARTTLS if supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: host}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	}

	auth := utils.LoginAuth(user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err := w.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("writing message failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing DATA writer failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		log.Printf("warning: QUIT returned error: %v", err)
	}
	return nil
}

// SendMailWithCC sends an email with optional CC recipients.
func SendMailWithCC(to []string, cc []string, subject, body, contentType string) error {
	_ = godotenv.Load(".env")
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if from == "" {
		from = user
	}
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}

	if host == "" || port == "" || user == "" || pass == "" {
		return fmt.Errorf("smtp configuration incomplete: host=%q port=%q user=%q from=%q", host, port, user, from)
	}

	addr := host + ":" + port
	hdr := make(map[string]string)
	hdr["From"] = from
	hdr["To"] = strings.Join(to, ", ")
	if len(cc) > 0 {
		hdr["Cc"] = strings.Join(cc, ", ")
	}
	hdr["Subject"] = subject
	hdr["MIME-Version"] = "1.0"
	hdr["Content-Type"] = contentType

	var msg strings.Builder
	for k, v := range hdr {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial smtp server: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: host}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	}

	auth := utils.LoginAuth(user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	allRcpt := append(to, cc...)
	for _, rcpt := range allRcpt {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err := w.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("writing message failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing DATA writer failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		log.Printf("warning: QUIT returned error: %v", err)
	}
	return nil
}

// SendMailWithAttachment sends an email with a single attachment (xlsx or other binary).
func SendMailWithAttachment(to []string, subject, body, contentType, attachmentName string, attachmentData []byte) error {
	_ = godotenv.Load(".env")
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if from == "" {
		from = user
	}
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}

	if host == "" || port == "" || user == "" || pass == "" {
		return fmt.Errorf("smtp configuration incomplete: host=%q port=%q user=%q from=%q", host, port, user, from)
	}

	addr := host + ":" + port

	// multipart/mixed with boundary
	boundary := "====BOUNDARY======"

	hdr := make(map[string]string)
	hdr["From"] = from
	hdr["To"] = strings.Join(to, ", ")
	hdr["Subject"] = subject
	hdr["MIME-Version"] = "1.0"
	hdr["Content-Type"] = "multipart/mixed; boundary=" + boundary

	var msg strings.Builder
	for k, v := range hdr {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n")

	// body part
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: " + contentType + "\r\n\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n")

	// attachment part (base64)
	msg.WriteString("--" + boundary + "\r\n")
	mimeType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(attachmentName), ".xlsx") {
		mimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	msg.WriteString("Content-Type: " + mimeType + "; name=\"" + attachmentName + "\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("Content-Disposition: attachment; filename=\"" + attachmentName + "\"\r\n\r\n")

	enc := base64.StdEncoding.EncodeToString(attachmentData)
	// wrap lines at 76 chars per RFC
	for i := 0; i < len(enc); i += 76 {
		end := i + 76
		if end > len(enc) {
			end = len(enc)
		}
		msg.WriteString(enc[i:end] + "\r\n")
	}
	msg.WriteString("\r\n--" + boundary + "--\r\n")

	// send via SMTP client (so we can STARTTLS + custom auth)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial smtp server: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: host}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	}

	auth := utils.LoginAuth(user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err := w.Write([]byte(msg.String())); err != nil {
		return fmt.Errorf("writing message failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing DATA writer failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		log.Printf("warning: QUIT returned error: %v", err)
	}
	return nil
}

// ============================================================================
// Template Layout Helper Functions
// ============================================================================

// applySheetSetup sets column widths for the sheet
func applySheetSetup(f *excelize.File, sheet string) {
	_ = f.SetColWidth(sheet, "A", "A", 3.91)
	_ = f.SetColWidth(sheet, "B", "B", 10)
	_ = f.SetColWidth(sheet, "C", "C", 45.09)
	_ = f.SetColWidth(sheet, "D", "D", 12.09)
	_ = f.SetColWidth(sheet, "E", "E", 14.73)
	_ = f.SetColWidth(sheet, "F", "F", 14.63)
	_ = f.SetColWidth(sheet, "G", "G", 18)
	_ = f.SetColWidth(sheet, "H", "H", 25.91)
}

// buildStyles creates and returns style format IDs for the workbook
func buildStyles(f *excelize.File) (base, title, header, projectHeader, statusDone, statusDelay, statusInprog, statusWaiting, statusReject int, err error) {
	// Base style with borders and center alignment
	base, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	projectHeader, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Title style: bold, large font, centered
	title, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 28,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Header style: borders and center alignment
	header, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Status Done: white bold text on green background
	statusDone, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"92D050"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Status Delay: white bold text on red background
	statusDelay, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"FF0000"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Status In Progress: black bold text on yellow background
	statusInprog, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "000000",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"FFC000"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Status Waiting: white bold text on light purple background
	statusWaiting, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"9B59B6"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	// Status Reject: white bold text on orange background
	statusReject, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"FF6600"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return
	}

	return
}

// buildTitleBlock creates a merged title block at the top of the sheet
func buildTitleBlock(f *excelize.File, sheet string, titleStyle int) {
	_ = f.MergeCell(sheet, "B1", "H4")
	_ = f.SetCellValue(sheet, "B1", "Project Management Tracking")
	_ = f.SetCellStyle(sheet, "B1", "H4", titleStyle)
}

// statusStyle returns the appropriate style based on status value
func statusStyle(status string, done, delay, inprog, waiting, reject, base int) int {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "done":
		return done
	case "delay":
		return delay
	case "inprogress":
		return inprog
	case "waiting":
		return waiting
	case "reject":
		return reject
	default:
		return base
	}
}

// statusDisplayText maps status values to display text
func statusDisplayText(status string, iaType string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "waiting":
		switch strings.ToLower(strings.TrimSpace(iaType)) {
		case "leader":
			return "Waiting Leader Approve"
		case "pj":
			return "Waiting PJ Approve"
		default:
			return "Waiting Approve"
		}
	case "done":
		return "Done"
	case "delay":
		if strings.EqualFold(strings.TrimSpace(iaType), "pj") {
			return "Waiting PJ Approve"
		}
		return "Delay"
	case "inprogress":
		return "Inprogress"
	case "reject":
		return "Reject"
	default:
		return strings.Title(strings.ToLower(status))
	}
}

// writeProjectBlock writes a complete project block with merges, styling, and data table
func writeProjectBlock(
	f *excelize.File,
	sheet string,
	startRow int,
	projectCode, partNo, partName, model, template string,
	rows []MailData,
	baseStyle, headerStyle, projectHeaderStyle, stDone, stDelay, stInprog, stWaiting, stReject int,
) (nextRow int) {

	r := startRow

	// Create merged cells for project info header
	_ = f.MergeCell(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("C%d", r+1))
	_ = f.MergeCell(sheet, fmt.Sprintf("D%d", r), fmt.Sprintf("F%d", r+1))
	_ = f.MergeCell(sheet, fmt.Sprintf("G%d", r), fmt.Sprintf("H%d", r+1))

	// Create merged cells for model info
	_ = f.MergeCell(sheet, fmt.Sprintf("B%d", r+2), fmt.Sprintf("C%d", r+3))
	_ = f.MergeCell(sheet, fmt.Sprintf("D%d", r+2), fmt.Sprintf("H%d", r+3))

	// Separator row
	_ = f.MergeCell(sheet, fmt.Sprintf("B%d", r+4), fmt.Sprintf("H%d", r+4))

	// Set header values
	_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", r), "Project Code : "+projectCode)
	_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", r), "Part Number : "+partNo)
	_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", r), "Part Name : "+partName)
	_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", r+2), "Model : "+model)
	_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", r+2), "Template : "+template)

	// Apply style to project header block
	_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", r), fmt.Sprintf("H%d", r+4), projectHeaderStyle)

	// Table header row
	th := r + 5
	_ = f.SetRowHeight(sheet, th, 24)

	headers := []string{"Phase", "Item Name", "Item Type", "Start Date", "End Date", "Status", "Pic"}
	headerCols := []string{"B", "C", "D", "E", "F", "G", "H"}

	for cidx, h := range headers {
		cell := fmt.Sprintf("%s%d", headerCols[cidx], th)
		_ = f.SetCellValue(sheet, cell, h)
	}

	_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", th), fmt.Sprintf("H%d", th), headerStyle)

	// Data rows
	dr := th + 1
	for _, it := range rows {
		_ = f.SetRowHeight(sheet, dr, 20.5)

		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", dr), getMmpDisplayValue(it.MmpID))
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", dr), getStringValue(it.ItemName))
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", dr), getStringValue(it.ItemType))
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", dr), getDateValue(it.StartDate))
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", dr), getDateValue(it.EndDate))
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", dr), statusDisplayText(getStringValue(it.IpidStatus), getStringValue(it.IaType)))
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", dr), getStringValue(it.OwnerNames))

		// Apply base style to non-status columns
		_ = f.SetCellStyle(sheet, fmt.Sprintf("B%d", dr), fmt.Sprintf("F%d", dr), baseStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("H%d", dr), fmt.Sprintf("H%d", dr), baseStyle)

		// Apply status style to Status column
		st := statusStyle(getStringValue(it.IpidStatus), stDone, stDelay, stInprog, stWaiting, stReject, baseStyle)
		_ = f.SetCellStyle(sheet, fmt.Sprintf("G%d", dr), fmt.Sprintf("G%d", dr), st)

		dr++
	}

	// Leave 2 blank rows before next block
	return dr + 2
}

// ============================================================================
// SendMailAuto Function
// ============================================================================

// RunSendMailAuto contains the core send-mail logic so it can be called by
// both the HTTP handler and the background scheduler.
func RunSendMailAuto(db *sqlx.DB) error {
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
		log.Println("[sendMail] SendMailAuto: no data to send")
		return nil
	}

	ownerIDs := make([]string, 0)
	ownerIDSet := make(map[string]struct{})
	for _, row := range rows {
		if !row.OwnerSuIDs.Valid {
			continue
		}
		for _, rawID := range strings.Split(row.OwnerSuIDs.String, "/") {
			ownerID := strings.TrimSpace(rawID)
			if ownerID == "" || ownerID == "0" || ownerID == "-" {
				continue
			}
			if _, exists := ownerIDSet[ownerID]; exists {
				continue
			}
			ownerIDSet[ownerID] = struct{}{}
			ownerIDs = append(ownerIDs, ownerID)
		}
	}
	if len(ownerIDs) == 0 {
		log.Println("[sendMail] SendMailAuto: no owner_su_id values found for department email lookup")
		return nil
	}

	type DepartmentEmail struct {
		Email string `db:"sd_email"`
	}
	deptEmailQuery, deptEmailArgs, err := sqlx.In(`
		SELECT DISTINCT sd.sd_email
		FROM sys_user su
		JOIN sys_department sd ON sd.sd_id = su.sd_id
		WHERE su.su_id IN (?)
			AND su.su_status = 'active'
			AND sd.sd_status = 'active'
			AND sd.sd_email IS NOT NULL
			AND sd.sd_email <> ''
		ORDER BY sd.sd_email
	`, ownerIDs)
	if err != nil {
		return fmt.Errorf("failed to build department email query: %w", err)
	}
	deptEmailQuery = db.Rebind(deptEmailQuery)

	var deptEmails []DepartmentEmail
	if err := db.Select(&deptEmails, deptEmailQuery, deptEmailArgs...); err != nil {
		return fmt.Errorf("failed to query department emails: %w", err)
	}
	if len(deptEmails) == 0 {
		log.Println("[sendMail] SendMailAuto: no department email found for owner_su_id values")
		return nil
	}

	emailList := make([]string, 0, len(deptEmails))
	for _, deptEmail := range deptEmails {
		emailList = append(emailList, deptEmail.Email)
	}

	// prepare sheet name sanitizer
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

	// Build workbook with all items
	f := excelize.NewFile()

	// Build styles upfront
	baseStyle, titleStyle, headerStyle, projectHeaderStyle, stDone, stDelay, stInprog, stWaiting, stReject, err := buildStyles(f)
	if err != nil {
		return fmt.Errorf("failed to build styles: %w", err)
	}

	// Group items by model -> ip_code
	modelGroups := map[string]map[string][]MailData{}
	for _, it := range rows {
		m := getStringValue(it.IpModel)
		ipc := getStringValue(it.ProjectCode)
		if _, ok := modelGroups[m]; !ok {
			modelGroups[m] = map[string][]MailData{}
		}
		modelGroups[m][ipc] = append(modelGroups[m][ipc], it)
	}

	// deterministic ordering
	models := make([]string, 0, len(modelGroups))
	for m := range modelGroups {
		models = append(models, m)
	}
	sort.Strings(models)

	// Create sheets and populate with template layout
	firstSheet := true
	for _, model := range models {
		sheetName := sanitize(model)

		// Use default Sheet1 for first model, create new sheets for others
		if firstSheet {
			_ = f.SetSheetName("Sheet1", sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		// Apply sheet setup (column widths)
		applySheetSetup(f, sheetName)

		// Build title block
		buildTitleBlock(f, sheetName, titleStyle)

		currentRow := 5 // Start after title block

		// ip_codes sorted
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

			// Placeholder for template name - fetch from DB if available
			templateName := "-"

			// Use helper function to write complete project block
			currentRow = writeProjectBlock(
				f, sheetName, currentRow,
				getStringValue(first.ProjectCode),
				getStringValue(first.PartNo),
				getStringValue(first.PartName),
				getStringValue(first.IpModel),
				templateName,
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
	// Send to owner departments
	subject := "TBKK Project Management Tracking Status"
	attachName := "TrackingProjects.xlsx"
	body := `<html>
<body style="margin:0;padding:0;font-family:Arial,sans-serif;background-color:#f5f7fa;color:#1f2937;">
	<div style="max-width:600px;margin:0 auto;padding:32px 24px;">
		<div style="background-color:#ffffff;border-radius:12px;padding:32px;box-shadow:0 4px 16px rgba(15,23,42,0.08);">
			<p style="margin:0 0 16px 0;font-size:16px;line-height:24px;">Please review the attached file and kindly proceed with the assigned action items.</p>
			<div style="margin-top:24px;">
				<a href="http://192.168.161.205:4009/login" style="display:inline-block;padding:12px 24px;background-color:#0d6efd;color:#ffffff;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;">Open Project Tracking</a>
			</div>
		</div>
	</div>
</body>
</html>`
	if err := SendMailWithAttachment(emailList, subject, body, "text/html; charset=utf-8", attachName, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[sendMail] SendMailAuto: email sent to %d department recipient(s)", len(emailList))
	return nil
}

// SendMailAuto is the HTTP handler wrapper for RunSendMailAuto.
func SendMailAuto(c *fiber.Ctx, db *sqlx.DB) error {
	if err := RunSendMailAuto(db); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"success": true, "message": "Email sent to owner departments"})
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "-"
}

func getDateValue(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02")
	}
	return "-"
}

func getMmpDisplayValue(mmpID sql.NullInt64) string {
	if !mmpID.Valid || mmpID.Int64 == 6 {
		return ""
	}
	return fmt.Sprintf("%d", mmpID.Int64)
}

func getPicFirstName(first sql.NullString) string {
	if first.Valid && strings.TrimSpace(first.String) != "" {
		return "K. " + strings.TrimSpace(first.String)
	}
	return "-"
}

func getPicValue(emp, first sql.NullString) string {
	parts := []string{}
	if emp.Valid && strings.TrimSpace(emp.String) != "" {
		parts = append(parts, emp.String)
	}
	if first.Valid && strings.TrimSpace(first.String) != "" {
		parts = append(parts, first.String)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}
