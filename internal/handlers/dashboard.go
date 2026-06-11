package handlers

import (
	"time"

	"apiTrackingSystem/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// DashboardProject represents the row returned by the dashboard query
type DashboardProject struct {
	IpCode         utils.NullString `db:"ip_code" json:"ip_code"`
	IpModel        utils.NullString `db:"ip_model" json:"ip_model"`
	IpPartNo       utils.NullString `db:"ip_part_no" json:"ip_part_no"`
	IpPartName     utils.NullString `db:"ip_part_name" json:"ip_part_name"`
	IpSopYearMonth utils.NullString `db:"ip_sop_year_month" json:"ip_sop_year_month"`
	IpCustomerName utils.NullString `db:"ip_customer_name" json:"ip_customer_name"`
	IpmpName       utils.NullString `db:"ipmp_name" json:"ipmp_name"`
	IpmpStartDate  *time.Time       `db:"ipmp_start_date" json:"ipmp_start_date"`
	IpmpEndDate    *time.Time       `db:"ipmp_end_date" json:"ipmp_end_date"`
	IpmpStatus     utils.NullString `db:"ipmp_status" json:"ipmp_status"`
}

// MasterPlanSummary represents aggregated master plan statuses per project
type MasterPlanSummary struct {
	IpID                              int64            `db:"ip_id" json:"ip_id"`
	IpCode                            utils.NullString `db:"ip_code" json:"ip_code"`
	IpCustomerName                    utils.NullString `db:"ip_customer_name" json:"ip_customer_name"`
	IpModel                           utils.NullString `db:"ip_model" json:"ip_model"`
	IpPartNo                          utils.NullString `db:"ip_part_no" json:"ip_part_no"`
	IpPartName                        utils.NullString `db:"ip_part_name" json:"ip_part_name"`
	IpSopDate                         *time.Time       `db:"ip_sop_date" json:"ip_sop_date"`
	KickOff                           utils.NullString `db:"Kick_Off" json:"kick_off"`
	KickOffStartDate                  *time.Time       `db:"Kick_Off_start_date" json:"kick_off_start_date"`
	KickOffEndDate                    *time.Time       `db:"Kick_Off_end_date" json:"kick_off_end_date"`
	SupplierKickOff                   utils.NullString `db:"Supplier_Kick_Off" json:"supplier_kick_off"`
	SupplierKickOffStartDate          *time.Time       `db:"Supplier_Kick_Off_start_date" json:"supplier_kick_off_start_date"`
	SupplierKickOffEndDate            *time.Time       `db:"Supplier_Kick_Off_end_date" json:"supplier_kick_off_end_date"`
	MoldAndMCToolingReview            utils.NullString `db:"Mold_And_MC_Tooling_Review" json:"mold_and_mc_tooling_review"`
	MoldAndMCToolingReviewStartDate   *time.Time       `db:"Mold_And_MC_Tooling_Review_start_date" json:"mold_and_mc_tooling_review_start_date"`
	MoldAndMCToolingReviewEndDate     *time.Time       `db:"Mold_And_MC_Tooling_Review_end_date" json:"mold_and_mc_tooling_review_end_date"`
	MoldPO                            utils.NullString `db:"Mold_PO" json:"mold_po"`
	MoldPOStartDate                   *time.Time       `db:"Mold_PO_start_date" json:"mold_po_start_date"`
	MoldPOEndDate                     *time.Time       `db:"Mold_PO_end_date" json:"mold_po_end_date"`
	ToolingPO                         utils.NullString `db:"Tooling_PO" json:"tooling_po"`
	ToolingPOStartDate                *time.Time       `db:"Tooling_PO_start_date" json:"tooling_po_start_date"`
	ToolingPOEndDate                  *time.Time       `db:"Tooling_PO_end_date" json:"tooling_po_end_date"`
	OTSOffToolsSample                 utils.NullString `db:"OTS_Off_Tools_Sample" json:"ots_off_tools_sample"`
	OTSOffToolsSampleStartDate        *time.Time       `db:"OTS_Off_Tools_Sample_start_date" json:"ots_off_tools_sample_start_date"`
	OTSOffToolsSampleEndDate          *time.Time       `db:"OTS_Off_Tools_Sample_end_date" json:"ots_off_tools_sample_end_date"`
	InitialPpk                        utils.NullString `db:"Initial_Ppk" json:"initial_ppk"`
	InitialPpkStartDate               *time.Time       `db:"Initial_Ppk_start_date" json:"initial_ppk_start_date"`
	InitialPpkEndDate                 *time.Time       `db:"Initial_Ppk_end_date" json:"initial_ppk_end_date"`
	OPSOffProcessSample               utils.NullString `db:"OPS_Off_Process_Sample" json:"ops_off_process_sample"`
	OPSOffProcessSampleStartDate      *time.Time       `db:"OPS_Off_Process_Sample_start_date" json:"ops_off_process_sample_start_date"`
	OPSOffProcessSampleEndDate        *time.Time       `db:"OPS_Off_Process_Sample_end_date" json:"ops_off_process_sample_end_date"`
	ResultPpkPass                     utils.NullString `db:"Result_Ppk_Pass" json:"result_ppk_pass"`
	ResultPpkPassStartDate            *time.Time       `db:"Result_Ppk_Pass_start_date" json:"result_ppk_pass_start_date"`
	ResultPpkPassEndDate              *time.Time       `db:"Result_Ppk_Pass_end_date" json:"result_ppk_pass_end_date"`
	PreRAR                            utils.NullString `db:"Pre_R_A_R" json:"pre_ra_r"`
	PreRARStartDate                   *time.Time       `db:"Pre_R_A_R_start_date" json:"pre_r_a_r_start_date"`
	PreRAREndDate                     *time.Time       `db:"Pre_R_A_R_end_date" json:"pre_r_a_r_end_date"`
	RAR                               utils.NullString `db:"R_A_R" json:"r_a_r"`
	RARStartDate                      *time.Time       `db:"R_A_R_start_date" json:"r_a_r_start_date"`
	RAREndDate                        *time.Time       `db:"R_A_R_end_date" json:"r_a_r_end_date"`
	InternalAuditIATFSafety           utils.NullString `db:"Internal_Audit_IATF_Safety" json:"internal_audit_iatf_safety"`
	InternalAuditIATFSafetyStartDate  *time.Time       `db:"Internal_Audit_IATF_Safety_start_date" json:"internal_audit_iatf_safety_start_date"`
	InternalAuditIATFSafetyEndDate    *time.Time       `db:"Internal_Audit_IATF_Safety_end_date" json:"internal_audit_iatf_safety_end_date"`
	TBKKPPAPSubmitt                   utils.NullString `db:"TBKK_PPAP_Submitt" json:"tbkk_ppap_submitt"`
	TBKKPPAPSubmittStartDate          *time.Time       `db:"TBKK_PPAP_Submitt_start_date" json:"tbkk_ppap_submitt_start_date"`
	TBKKPPAPSubmittEndDate            *time.Time       `db:"TBKK_PPAP_Submitt_end_date" json:"tbkk_ppap_submitt_end_date"`
	CustomerAuditPpap                 utils.NullString `db:"Customer_Audit_ppap" json:"customer_audit_ppap"`
	CustomerAuditPpapStartDate        *time.Time       `db:"Customer_Audit_ppap_start_date" json:"customer_audit_ppap_start_date"`
	CustomerAuditPpapEndDate          *time.Time       `db:"Customer_Audit_ppap_end_date" json:"customer_audit_ppap_end_date"`
	CustomerPPAPApproved              utils.NullString `db:"Customer_PPAP_Approved" json:"customer_ppap_approved"`
	CustomerPPAPApprovedStartDate     *time.Time       `db:"Customer_PPAP_Approved_start_date" json:"customer_ppap_approved_start_date"`
	CustomerPPAPApprovedEndDate       *time.Time       `db:"Customer_PPAP_Approved_end_date" json:"customer_ppap_approved_end_date"`
	AssessmentProjectSignOff          utils.NullString `db:"Assessment_Project_Sign_off" json:"assessment_project_sign_off"`
	AssessmentProjectSignOffStartDate *time.Time       `db:"Assessment_Project_Sign_off_start_date" json:"assessment_project_sign_off_start_date"`
	AssessmentProjectSignOffEndDate   *time.Time       `db:"Assessment_Project_Sign_off_end_date" json:"assessment_project_sign_off_end_date"`
	PPPreProduct                      utils.NullString `db:"PP_Pre_Product" json:"pp_pre_product"`
	PPPreProductStartDate             *time.Time       `db:"PP_Pre_Product_start_date" json:"pp_pre_product_start_date"`
	PPPreProductEndDate               *time.Time       `db:"PP_Pre_Product_end_date" json:"pp_pre_product_end_date"`
	TBKKSOPStartOfProduction          utils.NullString `db:"TBKK_SOP_Start_Of_Production" json:"tbkk_sop_start_of_production"`
	TBKKSOPStartOfProductionStartDate *time.Time       `db:"TBKK_SOP_Start_Of_Production_start_date" json:"tbkk_sop_start_of_production_start_date"`
	TBKKSOPStartOfProductionEndDate   *time.Time       `db:"TBKK_SOP_Start_Of_Production_end_date" json:"tbkk_sop_start_of_production_end_date"`
	InitialControl3Month              utils.NullString `db:"Initial_Control_3_Month" json:"initial_control_3_month"`
	InitialControl3MonthStartDate     *time.Time       `db:"Initial_Control_3_Month_start_date" json:"initial_control_3_month_start_date"`
	InitialControl3MonthEndDate       *time.Time       `db:"Initial_Control_3_Month_end_date" json:"initial_control_3_month_end_date"`
}

// ListInprogressProjects returns projects in progress with their master plan info
func ListInprogressProjects(c *fiber.Ctx, db *sqlx.DB) error {
	query := `SELECT
  ip.ip_code,
  ip.ip_model,
  ip.ip_part_no,
  ip.ip_part_name,
  DATE_FORMAT(ip.ip_sop_date, '%Y-%m') AS ip_sop_year_month,
  ip.ip_customer_name,
  ipmp.ipmp_name,
  ipmp.ipmp_start_date ,
  ipmp.ipmp_end_date ,
  ipmp.ipmp_status
FROM info_project AS ip
LEFT JOIN info_project_master_plan AS ipmp
  ON ip.ip_id = ipmp.ip_id
WHERE ip.ip_status = 'inprogress'
GROUP BY
  ip.ip_code,
  ip.ip_model,
  ip.ip_part_no,
  ip.ip_part_name,
  DATE_FORMAT(ip.ip_sop_date, '%Y-%m'),
  ip.ip_customer_name,
  ipmp.ipmp_name;`

	var rows []DashboardProject
	if err := db.Select(&rows, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(rows)
}

// ListMasterPlanSummary returns a pivoted view of master plan statuses per project
func ListMasterPlanSummary(c *fiber.Ctx, db *sqlx.DB) error {
	query := `SELECT
					ipmp.ip_id,
					ip.ip_code,
					ip.ip_customer_name,
					ip.ip_model,
					ip.ip_part_name,
					ip.ip_sop_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Kick Off' THEN ipmp.ipmp_status END) AS Kick_Off,
					MAX(CASE WHEN ipmp.ipmp_name = 'Kick Off' THEN ipmp.ipmp_start_date END) AS Kick_Off_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Kick Off' THEN ipmp.ipmp_end_date END) AS Kick_Off_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Supplier Kick Off' THEN ipmp.ipmp_status END) AS Supplier_Kick_Off,
					MAX(CASE WHEN ipmp.ipmp_name = 'Supplier Kick Off' THEN ipmp.ipmp_start_date END) AS Supplier_Kick_Off_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Supplier Kick Off' THEN ipmp.ipmp_end_date END) AS Supplier_Kick_Off_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Mold & M/C Tooling Review' THEN ipmp.ipmp_status END) AS Mold_And_MC_Tooling_Review,
					MAX(CASE WHEN ipmp.ipmp_name = 'Mold & M/C Tooling Review' THEN ipmp.ipmp_start_date END) AS Mold_And_MC_Tooling_Review_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Mold & M/C Tooling Review' THEN ipmp.ipmp_end_date END) AS Mold_And_MC_Tooling_Review_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Mold PO' THEN ipmp.ipmp_status END) AS Mold_PO,
					MAX(CASE WHEN ipmp.ipmp_name = 'Mold PO' THEN ipmp.ipmp_start_date END) AS Mold_PO_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Mold PO' THEN ipmp.ipmp_end_date END) AS Mold_PO_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Tooling PO' THEN ipmp.ipmp_status END) AS Tooling_PO,
					MAX(CASE WHEN ipmp.ipmp_name = 'Tooling PO' THEN ipmp.ipmp_start_date END) AS Tooling_PO_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Tooling PO' THEN ipmp.ipmp_end_date END) AS Tooling_PO_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'OTS : Off Tools Sample' THEN ipmp.ipmp_status END) AS OTS_Off_Tools_Sample,
					MAX(CASE WHEN ipmp.ipmp_name = 'OTS : Off Tools Sample' THEN ipmp.ipmp_start_date END) AS OTS_Off_Tools_Sample_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'OTS : Off Tools Sample' THEN ipmp.ipmp_end_date END) AS OTS_Off_Tools_Sample_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Initial Ppk' THEN ipmp.ipmp_status END) AS Initial_Ppk,
					MAX(CASE WHEN ipmp.ipmp_name = 'Initial Ppk' THEN ipmp.ipmp_start_date END) AS Initial_Ppk_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Initial Ppk' THEN ipmp.ipmp_end_date END) AS Initial_Ppk_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'OPS : Off Process Sample' THEN ipmp.ipmp_status END) AS OPS_Off_Process_Sample,
					MAX(CASE WHEN ipmp.ipmp_name = 'OPS : Off Process Sample' THEN ipmp.ipmp_start_date END) AS OPS_Off_Process_Sample_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'OPS : Off Process Sample' THEN ipmp.ipmp_end_date END) AS OPS_Off_Process_Sample_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Result Ppk Pass' THEN ipmp.ipmp_status END) AS Result_Ppk_Pass,
					MAX(CASE WHEN ipmp.ipmp_name = 'Result Ppk Pass' THEN ipmp.ipmp_start_date END) AS Result_Ppk_Pass_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Result Ppk Pass' THEN ipmp.ipmp_end_date END) AS Result_Ppk_Pass_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Pre-R@R' THEN ipmp.ipmp_status END) AS Pre_R_A_R,
					MAX(CASE WHEN ipmp.ipmp_name = 'Pre-R@R' THEN ipmp.ipmp_start_date END) AS Pre_R_A_R_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Pre-R@R' THEN ipmp.ipmp_end_date END) AS Pre_R_A_R_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'R@R' THEN ipmp.ipmp_status END) AS R_A_R,
					MAX(CASE WHEN ipmp.ipmp_name = 'R@R' THEN ipmp.ipmp_start_date END) AS R_A_R_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'R@R' THEN ipmp.ipmp_end_date END) AS R_A_R_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Internal Audit IATF & Safety' THEN ipmp.ipmp_status END) AS Internal_Audit_IATF_Safety,
					MAX(CASE WHEN ipmp.ipmp_name = 'Internal Audit IATF & Safety' THEN ipmp.ipmp_start_date END) AS Internal_Audit_IATF_Safety_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Internal Audit IATF & Safety' THEN ipmp.ipmp_end_date END) AS Internal_Audit_IATF_Safety_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'TBKK PPAP Submitt' THEN ipmp.ipmp_status END) AS TBKK_PPAP_Submitt,
					MAX(CASE WHEN ipmp.ipmp_name = 'TBKK PPAP Submitt' THEN ipmp.ipmp_start_date END) AS TBKK_PPAP_Submitt_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'TBKK PPAP Submitt' THEN ipmp.ipmp_end_date END) AS TBKK_PPAP_Submitt_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Customer Audit ppap' THEN ipmp.ipmp_status END) AS Customer_Audit_ppap,
					MAX(CASE WHEN ipmp.ipmp_name = 'Customer Audit ppap' THEN ipmp.ipmp_start_date END) AS Customer_Audit_ppap_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Customer Audit ppap' THEN ipmp.ipmp_end_date END) AS Customer_Audit_ppap_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Customer PPAP Approved' THEN ipmp.ipmp_status END) AS Customer_PPAP_Approved,
					MAX(CASE WHEN ipmp.ipmp_name = 'Customer PPAP Approved' THEN ipmp.ipmp_start_date END) AS Customer_PPAP_Approved_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Customer PPAP Approved' THEN ipmp.ipmp_end_date END) AS Customer_PPAP_Approved_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Assessment (Project  Sign-off)' THEN ipmp.ipmp_status END) AS Assessment_Project_Sign_off,
					MAX(CASE WHEN ipmp.ipmp_name = 'Assessment (Project  Sign-off)' THEN ipmp.ipmp_start_date END) AS Assessment_Project_Sign_off_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Assessment (Project  Sign-off)' THEN ipmp.ipmp_end_date END) AS Assessment_Project_Sign_off_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'PP : Pre Product' THEN ipmp.ipmp_status END) AS PP_Pre_Product,
					MAX(CASE WHEN ipmp.ipmp_name = 'PP : Pre Product' THEN ipmp.ipmp_start_date END) AS PP_Pre_Product_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'PP : Pre Product' THEN ipmp.ipmp_end_date END) AS PP_Pre_Product_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'TBKK SOP : Start Of Production' THEN ipmp.ipmp_status END) AS TBKK_SOP_Start_Of_Production,
					MAX(CASE WHEN ipmp.ipmp_name = 'TBKK SOP : Start Of Production' THEN ipmp.ipmp_start_date END) AS TBKK_SOP_Start_Of_Production_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'TBKK SOP : Start Of Production' THEN ipmp.ipmp_end_date END) AS TBKK_SOP_Start_Of_Production_end_date,

					MAX(CASE WHEN ipmp.ipmp_name = 'Initial Control 3 Month' THEN ipmp.ipmp_status END) AS Initial_Control_3_Month,
					MAX(CASE WHEN ipmp.ipmp_name = 'Initial Control 3 Month' THEN ipmp.ipmp_start_date END) AS Initial_Control_3_Month_start_date,
					MAX(CASE WHEN ipmp.ipmp_name = 'Initial Control 3 Month' THEN ipmp.ipmp_end_date END) AS Initial_Control_3_Month_end_date

				FROM info_project_master_plan ipmp
				LEFT JOIN info_project ip ON ip.ip_id = ipmp.ip_id
				GROUP BY
					ipmp.ip_id,
					ip.ip_part_no;`

	var rows []MasterPlanSummary
	if err := db.Select(&rows, query); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error", "detail": err.Error()})
	}
	return c.Status(200).JSON(rows)
}
