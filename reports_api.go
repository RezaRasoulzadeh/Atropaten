package main

import (
	"Atropaten/internal/application"
	"Atropaten/internal/domain"
	"fmt"
)

type ReportSummaryDTO struct {
	Key                 string `json:"key"`
	Label               string `json:"label"`
	AmountRial          int64  `json:"amountRial"`
	SecondaryAmountRial int64  `json:"secondaryAmountRial"`
	Count               int64  `json:"count"`
}
type ReportRowDTO struct {
	ID                     string `json:"id"`
	ReferenceID            string `json:"referenceId"`
	Name                   string `json:"name"`
	SecondaryName          string `json:"secondaryName"`
	Category               string `json:"category"`
	Status                 string `json:"status"`
	Date                   string `json:"date"`
	AmountRial             int64  `json:"amountRial"`
	SecondaryAmountRial    int64  `json:"secondaryAmountRial"`
	TertiaryAmountRial     int64  `json:"tertiaryAmountRial"`
	QuantityUnits          int64  `json:"quantityUnits"`
	SecondaryQuantityUnits int64  `json:"secondaryQuantityUnits"`
}
type ReportDTO struct {
	Kind      string             `json:"kind"`
	StartDate string             `json:"startDate"`
	EndDate   string             `json:"endDate"`
	Summaries []ReportSummaryDTO `json:"summaries"`
	Rows      []ReportRowDTO     `json:"rows"`
}
type DashboardAttentionDTO struct {
	Label      string `json:"label"`
	Detail     string `json:"detail"`
	Date       string `json:"date"`
	Kind       string `json:"kind"`
	AmountRial int64  `json:"amountRial"`
}
type DashboardLowStockDTO struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Unit              string `json:"unit"`
	AvailableUnits    int64  `json:"availableUnits"`
	ReorderLevelUnits int64  `json:"reorderLevelUnits"`
	AverageCostRial   int64  `json:"averageCostRial"`
	ValueRial         int64  `json:"valueRial"`
}
type DashboardProductionDTO struct {
	ID          string `json:"id"`
	OrderID     string `json:"orderId"`
	OrderNumber string `json:"orderNumber"`
	Customer    string `json:"customer"`
	Service     string `json:"service"`
	Status      string `json:"status"`
	DueDate     string `json:"dueDate"`
}
type DashboardActivityDTO struct {
	ID         string `json:"id"`
	Date       string `json:"date"`
	Label      string `json:"label"`
	Detail     string `json:"detail"`
	Direction  string `json:"direction"`
	AmountRial int64  `json:"amountRial"`
}
type DashboardDTO struct {
	StartDate         string                   `json:"startDate"`
	EndDate           string                   `json:"endDate"`
	RevenueRial       int64                    `json:"revenueRial"`
	GrossProfitRial   int64                    `json:"grossProfitRial"`
	CashRial          int64                    `json:"cashRial"`
	BankRial          int64                    `json:"bankRial"`
	ReceivableRial    int64                    `json:"receivableRial"`
	PayableRial       int64                    `json:"payableRial"`
	OpenInvoiceCount  int                      `json:"openInvoiceCount"`
	DueOrderCount     int                      `json:"dueOrderCount"`
	OverdueOrderCount int                      `json:"overdueOrderCount"`
	InProductionCount int                      `json:"inProductionCount"`
	ReadyOrderCount   int                      `json:"readyOrderCount"`
	Attention         []DashboardAttentionDTO  `json:"attention"`
	LowStock          []DashboardLowStockDTO   `json:"lowStock"`
	Production        []DashboardProductionDTO `json:"production"`
	RecentActivity    []DashboardActivityDTO   `json:"recentActivity"`
}
type ShopSettingsDTO struct {
	ShopName       string `json:"shopName"`
	ShopSubtitle   string `json:"shopSubtitle"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	Email          string `json:"email"`
	Website        string `json:"website"`
	RegistrationID string `json:"registrationId"`
	TaxID          string `json:"taxId"`
	LogoPath       string `json:"logoPath"`
	DocumentFooter string `json:"documentFooter"`
	DocumentNotes  string `json:"documentNotes"`
}
type PrintLineDTO struct {
	Description   string `json:"description"`
	Unit          string `json:"unit"`
	QuantityUnits int64  `json:"quantityUnits"`
	UnitPriceRial int64  `json:"unitPriceRial"`
	LineTotalRial int64  `json:"lineTotalRial"`
}
type StatementLineDTO struct {
	Date        string `json:"date"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
	DebitRial   int64  `json:"debitRial"`
	CreditRial  int64  `json:"creditRial"`
	BalanceRial int64  `json:"balanceRial"`
}
type PrintAllocationDTO struct {
	Reference  string `json:"reference"`
	TargetType string `json:"targetType"`
	AmountRial int64  `json:"amountRial"`
}
type PrintDocumentDTO struct {
	Kind            string               `json:"kind"`
	Number          string               `json:"number"`
	Date            string               `json:"date"`
	DueDate         string               `json:"dueDate"`
	Status          string               `json:"status"`
	CustomerName    string               `json:"customerName"`
	CustomerContact string               `json:"customerContact"`
	SupplierName    string               `json:"supplierName"`
	Reference       string               `json:"reference"`
	Method          string               `json:"method"`
	AccountName     string               `json:"accountName"`
	PaymentStatus   string               `json:"paymentStatus"`
	Notes           string               `json:"notes"`
	SubtotalRial    int64                `json:"subtotalRial"`
	DiscountRial    int64                `json:"discountRial"`
	TotalRial       int64                `json:"totalRial"`
	PaidRial        int64                `json:"paidRial"`
	RemainingRial   int64                `json:"remainingRial"`
	AmountRial      int64                `json:"amountRial"`
	Shop            ShopSettingsDTO      `json:"shop"`
	Lines           []PrintLineDTO       `json:"lines"`
	StatementLines  []StatementLineDTO   `json:"statementLines"`
	Allocations     []PrintAllocationDTO `json:"allocations"`
}

func (a *App) reportingService() (*application.ReportingService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.reporting == nil {
		return nil, fmt.Errorf("reporting service is not initialized")
	}
	return a.reporting, nil
}
func (a *App) GetReport(kind, startDate, endDate string) (ReportDTO, error) {
	s, e := a.reportingService()
	if e != nil {
		return ReportDTO{}, e
	}
	v, e := s.Report(a.materialContext(), kind, startDate, endDate)
	if e != nil {
		return ReportDTO{}, e
	}
	return reportDTO(v), nil
}
func (a *App) GetDashboard(startDate, endDate string) (DashboardDTO, error) {
	s, e := a.reportingService()
	if e != nil {
		return DashboardDTO{}, e
	}
	v, e := s.Dashboard(a.materialContext(), startDate, endDate)
	if e != nil {
		return DashboardDTO{}, e
	}
	return dashboardDTO(v), nil
}
func (a *App) GetPrintDocument(kind, id, startDate, endDate, partyID string) (PrintDocumentDTO, error) {
	s, e := a.reportingService()
	if e != nil {
		return PrintDocumentDTO{}, e
	}
	v, e := s.PrintDocument(a.materialContext(), kind, id, startDate, endDate, partyID)
	if e != nil {
		return PrintDocumentDTO{}, e
	}
	return printDocumentDTO(v), nil
}
func (a *App) GetShopSettings() (ShopSettingsDTO, error) {
	s, e := a.reportingService()
	if e != nil {
		return ShopSettingsDTO{}, e
	}
	v, e := s.ShopSettings(a.materialContext())
	return ShopSettingsDTO{v.ShopName, v.ShopSubtitle, v.Phone, v.Address, v.Email, v.Website, v.RegistrationID, v.TaxID, v.LogoPath, v.DocumentFooter, v.DocumentNotes}, e
}
func (a *App) SaveShopSettings(v ShopSettingsDTO) error {
	s, e := a.reportingService()
	if e != nil {
		return e
	}
	return s.SaveShopSettings(a.materialContext(), domainShopSettings(v))
}
func domainShopSettings(v ShopSettingsDTO) domain.ShopSettings {
	return domain.ShopSettings{ShopName: v.ShopName, ShopSubtitle: v.ShopSubtitle, Phone: v.Phone, Address: v.Address, Email: v.Email, Website: v.Website, RegistrationID: v.RegistrationID, TaxID: v.TaxID, LogoPath: v.LogoPath, DocumentFooter: v.DocumentFooter, DocumentNotes: v.DocumentNotes}
}

func reportDTO(v domain.Report) ReportDTO {
	out := ReportDTO{Kind: v.Kind, StartDate: v.StartDate, EndDate: v.EndDate}
	for _, x := range v.Summaries {
		out.Summaries = append(out.Summaries, ReportSummaryDTO{x.Key, x.Label, x.AmountRial, x.SecondaryAmountRial, x.Count})
	}
	for _, x := range v.Rows {
		out.Rows = append(out.Rows, ReportRowDTO{x.ID, x.ReferenceID, x.Name, x.SecondaryName, x.Category, x.Status, x.Date, x.AmountRial, x.SecondaryAmountRial, x.TertiaryAmountRial, x.QuantityUnits, x.SecondaryQuantityUnits})
	}
	return out
}
func dashboardDTO(v domain.Dashboard) DashboardDTO {
	out := DashboardDTO{StartDate: v.StartDate, EndDate: v.EndDate, RevenueRial: v.RevenueRial, GrossProfitRial: v.GrossProfitRial, CashRial: v.CashRial, BankRial: v.BankRial, ReceivableRial: v.ReceivableRial, PayableRial: v.PayableRial, OpenInvoiceCount: v.OpenInvoiceCount, DueOrderCount: v.DueOrderCount, OverdueOrderCount: v.OverdueOrderCount, InProductionCount: v.InProductionCount, ReadyOrderCount: v.ReadyOrderCount}
	for _, x := range v.Attention {
		out.Attention = append(out.Attention, DashboardAttentionDTO{x.Label, x.Detail, x.Date, x.Kind, x.AmountRial})
	}
	for _, x := range v.LowStock {
		out.LowStock = append(out.LowStock, DashboardLowStockDTO{x.ID, x.Name, x.Unit, x.AvailableUnits, x.ReorderLevelUnits, x.AverageCostRial, x.ValueRial})
	}
	for _, x := range v.Production {
		out.Production = append(out.Production, DashboardProductionDTO{x.ID, x.OrderID, x.OrderNumber, x.Customer, x.Service, x.Status, x.DueDate})
	}
	for _, x := range v.RecentActivity {
		out.RecentActivity = append(out.RecentActivity, DashboardActivityDTO{x.ID, x.Date, x.Label, x.Detail, x.Direction, x.AmountRial})
	}
	return out
}
func printDocumentDTO(v domain.PrintDocument) PrintDocumentDTO {
	out := PrintDocumentDTO{Kind: v.Kind, Number: v.Number, Date: v.Date, DueDate: v.DueDate, Status: v.Status, CustomerName: v.CustomerName, CustomerContact: v.CustomerContact, SupplierName: v.SupplierName, Reference: v.Reference, Method: v.Method, AccountName: v.AccountName, PaymentStatus: v.PaymentStatus, Notes: v.Notes, SubtotalRial: v.SubtotalRial, DiscountRial: v.DiscountRial, TotalRial: v.TotalRial, PaidRial: v.PaidRial, RemainingRial: v.RemainingRial, AmountRial: v.AmountRial, Shop: ShopSettingsDTO{v.Shop.ShopName, v.Shop.ShopSubtitle, v.Shop.Phone, v.Shop.Address, v.Shop.Email, v.Shop.Website, v.Shop.RegistrationID, v.Shop.TaxID, v.Shop.LogoPath, v.Shop.DocumentFooter, v.Shop.DocumentNotes}}
	for _, x := range v.Lines {
		out.Lines = append(out.Lines, PrintLineDTO{x.Description, x.Unit, x.QuantityUnits, x.UnitPriceRial, x.LineTotalRial})
	}
	for _, x := range v.StatementLines {
		out.StatementLines = append(out.StatementLines, StatementLineDTO{x.Date, x.Reference, x.Description, x.DebitRial, x.CreditRial, x.BalanceRial})
	}
	for _, x := range v.Allocations {
		out.Allocations = append(out.Allocations, PrintAllocationDTO{x.Reference, x.TargetType, x.AmountRial})
	}
	return out
}
