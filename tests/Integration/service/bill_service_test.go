package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/PunMung-66/ApartmentSys/model"
	"github.com/PunMung-66/ApartmentSys/tests/Integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestBillForService(t *testing.T) *model.Bill {
	t.Helper()

	email := fmt.Sprintf("billsvc-%d@test.com", time.Now().UnixNano())
	user := model.NewUser("Bill Svc", "8888888888", email, "password123", "TENANT")
	createdUser, err := setup.UserRepo.CreateUser(user)
	require.NoError(t, err)

	room := setup.CreateTestRoom("SVC101", 1, "Occupied")

	contract, err := setup.CreateTestContract(createdUser.ID, room.ID, time.Now().Format("2006-01-02"), time.Now().AddDate(0, 6, 0).Format("2006-01-02"), "Active")
	require.NoError(t, err)

	rate := &model.UtilityRate{
		Period:       time.Now().Format("2006-01"),
		WaterRate:    50.0,
		ElectricRate: 70.0,
		CommonFee:    500.0,
	}
	r := setup.TestDB.Create(rate)
	require.NoError(t, r.Error)

	bill := model.NewBill(
		contract.ID,
		rate.ID,
		time.Now(),
		3000.0,
		250.0,
		350.0,
		500.0,
		time.Now().AddDate(0, 0, 30),
		10, 15, // Old/New Water Units
		20, 25, // Old/New Electric Units
		rate.WaterRate,
		rate.ElectricRate,
	)
	err = setup.BillRepo.Create(bill)
	require.NoError(t, err)
	require.NotEmpty(t, bill.ID)

	return bill
}

func TestBillService_GetAllBills_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	bills, err := setup.BillService.GetAllBills()
	require.NoError(t, err)
	initialCount := len(bills)

	createTestBillForService(t)
	createTestBillForService(t)

	bills, err = setup.BillService.GetAllBills()
	require.NoError(t, err)
	assert.Equal(t, initialCount+2, len(bills))
}

func TestBillService_GetAllBills_Empty(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	bills, err := setup.BillService.GetAllBills()
	require.NoError(t, err)
	assert.Equal(t, 0, len(bills))
}

func TestBillService_GetBillByID_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBillForService(t)

	found, err := setup.BillService.GetBillByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.InDelta(t, 4100.0, found.TotalAmount, 0.01)
}

func TestBillService_GetBillByID_NotFound(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	_, err := setup.BillService.GetBillByID("nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bill not found")
}

func TestBillService_UpdateBill_StatusToPaid(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBillForService(t)
	assert.Equal(t, "Unpaid", created.Status)

	created.Status = "Paid"
	updated, err := setup.BillService.UpdateBill(created)
	require.NoError(t, err)
	assert.Equal(t, "Paid", updated.Status)
}

func TestBillService_UpdateBill_StatusToWaitingApproval(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBillForService(t)

	created.Status = "WaitingApproval"
	updated, err := setup.BillService.UpdateBill(created)
	require.NoError(t, err)
	assert.Equal(t, "WaitingApproval", updated.Status)
}

func TestBillService_UpdateBill_NotFound(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	bill := &model.Bill{ID: "nonexistent-id", Status: "Paid"}
	_, err := setup.BillService.UpdateBill(bill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bill not found")
}

func TestBillService_DeleteBill_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBillForService(t)

	err := setup.BillService.DeleteBill(created.ID)
	require.NoError(t, err)

	_, err = setup.BillService.GetBillByID(created.ID)
	assert.Error(t, err)
}

func TestBillService_DeleteBill_NotFound(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	err := setup.BillService.DeleteBill("nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bill not found")
}

func TestBillService_UpdateBill_StatusRejected(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBillForService(t)

	created.Status = "Rejected"
	updated, err := setup.BillService.UpdateBill(created)
	require.NoError(t, err)
	assert.Equal(t, "Rejected", updated.Status)
}

func TestBillService_MultipleStatusUpdates(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBillForService(t)

	created.Status = "WaitingApproval"
	_, err := setup.BillService.UpdateBill(created)
	require.NoError(t, err)

	found, err := setup.BillService.GetBillByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "WaitingApproval", found.Status)

	found.Status = "Paid"
	_, err = setup.BillService.UpdateBill(found)
	require.NoError(t, err)

	found, err = setup.BillService.GetBillByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Paid", found.Status)
}
