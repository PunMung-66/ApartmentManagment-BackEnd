package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/PunMung-66/ApartmentSys/model"
	"github.com/PunMung-66/ApartmentSys/repository"
	"github.com/PunMung-66/ApartmentSys/tests/Integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var billRepo repository.BillRepository

func init() {
	billRepo = repository.NewBillRepository(setup.TestDB)
}

func createTestBill(t *testing.T) *model.Bill {
	t.Helper()
	email := fmt.Sprintf("billtenant-%d@test.com", time.Now().UnixNano())
	user := model.NewUser("Bill Tenant", "9999999999", email, "password123", "TENANT")
	createdUser, err := setup.UserRepo.CreateUser(user)
	require.NoError(t, err)

	room := setup.CreateTestRoom("B101", 1, "Occupied")

	contract, err := setup.CreateTestContract(createdUser.ID, room.ID, time.Now().Format("2006-01-02"), time.Now().AddDate(0, 6, 0).Format("2006-01-02"), "Active")
	require.NoError(t, err)

	rate := &model.UtilityRate{
		Period:       time.Now().Format("2006-01"),
		WaterRate:    50.0,
		ElectricRate: 70.0,
		CommonFee:    500.0,
	}
	result := setup.TestDB.Create(rate)
	require.NoError(t, result.Error)

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
	err = billRepo.Create(bill)
	require.NoError(t, err)
	require.NotEmpty(t, bill.ID)

	return bill
}

func TestBillRepository_Create_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	bill := createTestBill(t)

	assert.NotEmpty(t, bill.ID)
	assert.Equal(t, 3000.0, bill.RentFee)
	assert.Equal(t, 250.0, bill.WaterFee)
	assert.Equal(t, 350.0, bill.ElectricityFee)
	assert.Equal(t, 500.0, bill.CommonFee)
	assert.Equal(t, "Unpaid", bill.Status)
}

func TestBillRepository_FindByID_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBill(t)

	found, err := billRepo.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.TotalAmount, found.TotalAmount)
}

func TestBillRepository_FindByID_NotFound(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	_, err := billRepo.FindByID("nonexistent-id")
	assert.Error(t, err)
}

func TestBillRepository_FindAll_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	bills, err := billRepo.FindAll()
	require.NoError(t, err)
	initialCount := len(bills)

	createTestBill(t)
	createTestBill(t)
	createTestBill(t)

	bills, err = billRepo.FindAll()
	require.NoError(t, err)
	assert.Equal(t, initialCount+3, len(bills))
}

func TestBillRepository_FindAll_Empty(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	bills, err := billRepo.FindAll()
	require.NoError(t, err)
	assert.Equal(t, 0, len(bills))
}

func TestBillRepository_Update_Status(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBill(t)
	assert.Equal(t, "Unpaid", created.Status)

	created.Status = "Paid"
	err := billRepo.Update(created)
	require.NoError(t, err)

	updated, err := billRepo.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Paid", updated.Status)
}

func TestBillRepository_Update_MultipleFields(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBill(t)

	created.Status = "WaitingApproval"
	err := billRepo.Update(created)
	require.NoError(t, err)

	updated, err := billRepo.FindByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "WaitingApproval", updated.Status)
}

func TestBillRepository_Delete_Success(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	created := createTestBill(t)

	err := billRepo.Delete(created.ID)
	require.NoError(t, err)

	_, err = billRepo.FindByID(created.ID)
	assert.Error(t, err)
}

func TestBillRepository_Delete_NotFound(t *testing.T) {
	setup.ResetTestDB()
	defer setup.ResetTestDB()

	err := billRepo.Delete("nonexistent-id")
	assert.NoError(t, err)
}
