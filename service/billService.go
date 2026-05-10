package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/PunMung-66/ApartmentSys/model"
	"github.com/PunMung-66/ApartmentSys/repository"
)

type BillService interface {
	GenerateMonthlyBill(roomID string, contractID string, recordDate time.Time, dueDate time.Time) (*model.Bill, error)
	GetAllBills() ([]model.Bill, error)
	GetBillByID(id string) (*model.Bill, error)
	GetBillsByUserID(userID string) ([]model.Bill, error)
	UpdateBill(bill *model.Bill) (*model.Bill, error)
	DeleteBill(billID string) error
}

type billService struct {
	billRepo     repository.BillRepository
	billSlipRepo *repository.BillSlipRepository
	contractRepo *repository.ContractRepository
	roomRepo     *repository.RoomRepository
	usageRepo    *repository.UtilityUsageRepository
	rateRepo     *repository.UtilityRateRepository
}

func NewBillService(br repository.BillRepository, bsr *repository.BillSlipRepository, cr *repository.ContractRepository, rr *repository.RoomRepository, ur *repository.UtilityUsageRepository, rate *repository.UtilityRateRepository) BillService {
	return &billService{
		billRepo:     br,
		billSlipRepo: bsr,
		contractRepo: cr,
		roomRepo:     rr,
		usageRepo:    ur,
		rateRepo:     rate,
	}
}

func (s *billService) GetAllBills() ([]model.Bill, error) {
	return s.billRepo.FindAll()
}

func (s *billService) GetBillByID(id string) (*model.Bill, error) {
	bill, err := s.billRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}
	return bill, nil
}

func (s *billService) GetBillsByUserID(userID string) ([]model.Bill, error) {
	contracts, err := s.contractRepo.FindContractsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("no contracts found for user: %w", err)
	}
	var contractIDs []string
	for _, c := range contracts {
		contractIDs = append(contractIDs, c.ID)
	}
	if len(contractIDs) == 0 {
		return []model.Bill{}, nil
	}
	return s.billRepo.FindByContractIDs(contractIDs)
}

func (s *billService) UpdateBill(bill *model.Bill) (*model.Bill, error) {
	existing, err := s.billRepo.FindByID(bill.ID)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	if bill.Status != "" {
		existing.Status = bill.Status
	}

	if err := s.billRepo.Update(existing); err != nil {
		return nil, fmt.Errorf("failed to update bill: %w", err)
	}

	return existing, nil
}

func (s *billService) DeleteBill(billID string) error {
	_, err := s.billRepo.FindByID(billID)
	if err != nil {
		return fmt.Errorf("bill not found: %w", err)
	}

	// Delete associated bill slip first to avoid FK constraint violation
	if err := s.billSlipRepo.DeleteByBillID(billID); err != nil {
		return fmt.Errorf("failed to delete bill slip: %w", err)
	}

	return s.billRepo.Delete(billID)
}

func (s *billService) GenerateMonthlyBill(roomID string, contractID string, recordDate time.Time, dueDate time.Time) (*model.Bill, error) {
	// 1. Room Expert (BR-02 Validation)
	room, err := s.roomRepo.FindRoomByID(roomID)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}
	if room.Status == "Available" {
		return nil, errors.New("BR-02 Violation: Cannot generate bill for an AVAILABLE room")
	}

	// 2. Get Utility Rate
	rate, err := s.rateRepo.FindLatestRate()
	if err != nil {
		return nil, errors.New("failed to retrieve active utility rates")
	}

	// 3. Get Utility Usage
	usage, err := s.usageRepo.FindLatestByContract(contractID)
	if err != nil || usage == nil {
		return nil, fmt.Errorf("utility usage data not found for contract %s", contractID)
	}

	// 4. Calculate Usage (BR-12 Validation)
	waterUnits, err := usage.CalculateWaterUsage()
	if err != nil {
		return nil, err
	}
	electricUnits, err := usage.CalculateElectricUsage()
	if err != nil {
		return nil, err
	}

	waterFee := float64(waterUnits) * rate.WaterRate
	electricFee := float64(electricUnits) * rate.ElectricRate

	// 5. Creator: Construct the Bill with usage and rate details
	newBill := model.NewBill(
		contractID,
		rate.ID,
		recordDate,
		3000, // Rent fee (Update later if needed)
		waterFee,
		electricFee,
		rate.CommonFee,
		dueDate,
		usage.OldWaterUnit,
		usage.NewWaterUnit,
		usage.OldElectricUnit,
		usage.NewElectricUnit,
		rate.WaterRate,
		rate.ElectricRate,
	)

	// 6. Save to Database
	if err := s.billRepo.Create(newBill); err != nil {
		return nil, fmt.Errorf("failed to save bill: %w", err)
	}

	return newBill, nil
}
