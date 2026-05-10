package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillStatus string

const (
	BillUnpaid          BillStatus = "Unpaid"
	BillWaitingApproval BillStatus = "WaitingApproval"
	BillPaid            BillStatus = "Paid"
	BillRejected        BillStatus = "Rejected"
)

type Bill struct {
	ID             string    `gorm:"type:char(36);primaryKey" json:"bill_id"`
	ContractID     string    `json:"contract_id" gorm:"not null"`
	RateID         string    `json:"rate_id" gorm:"not null"`
	RecordDate     time.Time `json:"record_date" gorm:"type:date;not null"`
	RentFee        float64   `json:"rent_fee" gorm:"type:decimal(10,2);not null"`
	WaterFee       float64   `json:"water_fee" gorm:"type:decimal(10,2);not null"`
	ElectricityFee float64   `json:"electricity_fee" gorm:"type:decimal(10,2);not null"`
	CommonFee      float64   `json:"common_fee" gorm:"type:decimal(10,2);not null"`
	TotalAmount    float64   `json:"total_amount" gorm:"type:decimal(10,2);not null"`

	// Usage details
	OldWaterUnit    int `json:"old_water_unit" gorm:"not null;default:0"`
	NewWaterUnit    int `json:"new_water_unit" gorm:"not null;default:0"`
	OldElectricUnit int `json:"old_electric_unit" gorm:"not null;default:0"`
	NewElectricUnit int `json:"new_electric_unit" gorm:"not null;default:0"`

	// Rate details at time of billing
	WaterRate    float64 `json:"water_rate" gorm:"type:decimal(10,2);not null;default:0"`
	ElectricRate float64 `json:"electric_rate" gorm:"type:decimal(10,2);not null;default:0"`

	// Status with BR-07 default
	Status string `json:"status" gorm:"not null;default:'Unpaid';check:status IN ('Unpaid','WaitingApproval','Paid','Rejected')"`

	DueDate time.Time `json:"due_date" gorm:"type:date;not null"`

	// ✅ THE FIX: We map GORM's automatic CreatedAt to your database's "created_date" column
	CreatedAt time.Time `json:"created_at" gorm:"column:created_date;not null"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Contract    *Contract    `gorm:"foreignKey:ContractID" json:"-"`
	UtilityRate *UtilityRate `gorm:"foreignKey:RateID" json:"-"`
	BillSlip    *BillSlip    `gorm:"foreignKey:BillID" json:"bill_slip,omitempty"`
}

func (Bill) TableName() string {
	return "bills"
}

// BeforeCreate hooks into GORM to set the ID and Timestamp automatically
func (b *Bill) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	// Safety check: Ensure CreatedAt is set if GORM hasn't filled it yet
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	return
}

// CalculateTotal handles the math for the bill (Information Expert)
func (b *Bill) CalculateTotal() {
	b.TotalAmount = b.RentFee + b.WaterFee + b.ElectricityFee + b.CommonFee
}

// NewBill Constructor
func NewBill(contractID, rateID string, recordDate time.Time, rentFee, waterFee, electricityFee, commonFee float64, dueDate time.Time, oldWater, newWater, oldElectric, newElectric int, waterRate, electricRate float64) *Bill {
	bill := &Bill{
		ContractID:      contractID,
		RateID:          rateID,
		RecordDate:      recordDate,
		RentFee:         rentFee,
		WaterFee:        waterFee,
		ElectricityFee:  electricityFee,
		CommonFee:       commonFee,
		DueDate:         dueDate,
		OldWaterUnit:    oldWater,
		NewWaterUnit:    newWater,
		OldElectricUnit: oldElectric,
		NewElectricUnit: newElectric,
		WaterRate:       waterRate,
		ElectricRate:    electricRate,
	}

	// Calculate the total immediately upon creation
	bill.CalculateTotal()

	return bill
}
