package controller

import (
	"net/http"
	"time"

	"github.com/PunMung-66/ApartmentSys/internal/response"
	"github.com/PunMung-66/ApartmentSys/model"
	"github.com/PunMung-66/ApartmentSys/service"
	"github.com/gin-gonic/gin"
)

type BillController struct {
	billService service.BillService
}

func NewBillController(bs service.BillService) *BillController {
	return &BillController{billService: bs}
}

// GenerateRequest represents the expected JSON payload
type GenerateRequest struct {
	RoomID     string `json:"room_id" binding:"required"`
	ContractID string `json:"contract_id" binding:"required"`
	RecordDate string `json:"record_date" binding:"required"`
	DueDate    string `json:"due_date" binding:"required"`
}

type UpdateBillRequest struct {
	Status string `json:"status"`
}

// GenerateBill handles the POST /bills/generate route
func (ctrl *BillController) GenerateBill(c *gin.Context) {
	var req GenerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	recordDate, err := time.Parse("2006-01-02", req.RecordDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record_date format. Use YYYY-MM-DD"})
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format. Use YYYY-MM-DD"})
		return
	}

	bill, err := ctrl.billService.GenerateMonthlyBill(req.RoomID, req.ContractID, recordDate, dueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bill generated successfully",
		"data":    bill,
	})
}

func (ctrl *BillController) GetBills(c *gin.Context) {
	bills, err := ctrl.billService.GetAllBills()
	if err != nil {
		res := response.NewAppResponse(http.StatusInternalServerError, "Failed to retrieve bills", nil)
		c.JSON(res.Status, res.Response())
		return
	}

	res := response.NewAppResponse(http.StatusOK, "Bills retrieved successfully", bills)
	c.JSON(res.Status, res.Response())
}

func (ctrl *BillController) GetBillByID(c *gin.Context) {
	billID := c.Param("id")

	bill, err := ctrl.billService.GetBillByID(billID)
	if err != nil {
		res := response.NewAppResponse(http.StatusNotFound, "Bill not found", nil)
		c.JSON(res.Status, res.Response())
		return
	}

	res := response.NewAppResponse(http.StatusOK, "Bill retrieved successfully", bill)
	c.JSON(res.Status, res.Response())
}

func (ctrl *BillController) UpdateBill(c *gin.Context) {
	billID := c.Param("id")

	var req UpdateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := response.NewAppResponse(http.StatusBadRequest, "Invalid request body", nil)
		c.JSON(res.Status, res.Response())
		return
	}

	bill := &model.Bill{ID: billID}
	if req.Status != "" {
		bill.Status = req.Status
	}

	updatedBill, err := ctrl.billService.UpdateBill(bill)
	if err != nil {
		res := response.NewAppResponse(http.StatusBadRequest, err.Error(), nil)
		c.JSON(res.Status, res.Response())
		return
	}

	res := response.NewAppResponse(http.StatusOK, "Bill updated successfully", updatedBill)
	c.JSON(res.Status, res.Response())
}

func (ctrl *BillController) DeleteBill(c *gin.Context) {
	billID := c.Param("id")

	err := ctrl.billService.DeleteBill(billID)
	if err != nil {
		res := response.NewAppResponse(http.StatusNotFound, "Failed to delete bill", nil)
		c.JSON(res.Status, res.Response())
		return
	}

	res := response.NewAppResponse(http.StatusOK, "Bill deleted successfully", nil)
	c.JSON(res.Status, res.Response())
}
