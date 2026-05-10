package setup

import (
	"time"

	"github.com/PunMung-66/ApartmentSys/model"
)

// CreateTestBill helper creates a test bill for a room
func CreateTestBill(userID string, roomID string) *model.Bill {
	// For simplicity, create a contract first using the real userID
	contract, _ := CreateTestContract(userID, roomID, time.Now().Format("2006-01-02"), time.Now().AddDate(0, 6, 0).Format("2006-01-02"), "Active")

	// Create the bill using exactly the 8 required arguments
	bill := model.NewBill(
		contract.ID,                  // Contract ID
		"test-rate-id",               // Rate ID
		time.Now(),                   // Record Date
		1000,                         // Rent Fee
		100,                          // Water Fee
		100,                          // Electricity Fee
		50,                           // Common Fee
		time.Now().AddDate(0, 0, 30), // Due Date
		0, 0,                         // Old/New Water Units
		0, 0, // Old/New Electric Units
		0, // Water Rate
		0, // Electric Rate
	)

	// We no longer assign bill.RoomID here because we removed it from the model!

	TestDB.Create(&bill)
	return bill
}
