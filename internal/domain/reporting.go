package domain

import "time"

type ReportSummary struct {
	Key, Label                      string
	AmountRial, SecondaryAmountRial int64
	Count                           int64
}

type ReportRow struct {
	ID, ReferenceID, Name, SecondaryName, Category, Status, Date string
	AmountRial, SecondaryAmountRial, TertiaryAmountRial          int64
	QuantityUnits, SecondaryQuantityUnits                        int64
}

type Report struct {
	Kind, StartDate, EndDate string
	Summaries                []ReportSummary
	Rows                     []ReportRow
}

type DashboardAttention struct {
	Label, Detail, Date, Kind string
	AmountRial                int64
}
type DashboardLowStock struct {
	ID, Name, Unit                    string
	AvailableUnits, ReorderLevelUnits int64
	AverageCostRial, ValueRial        int64
}
type DashboardProduction struct{ ID, OrderID, OrderNumber, Customer, Service, Status, DueDate string }
type DashboardActivity struct {
	ID, Date, Label, Detail, Direction string
	AmountRial                         int64
}

type Dashboard struct {
	StartDate, EndDate                                                                     string
	RevenueRial, GrossProfitRial, CashRial, BankRial, ReceivableRial, PayableRial          int64
	OpenInvoiceCount, DueOrderCount, OverdueOrderCount, InProductionCount, ReadyOrderCount int
	Attention                                                                              []DashboardAttention
	LowStock                                                                               []DashboardLowStock
	Production                                                                             []DashboardProduction
	RecentActivity                                                                         []DashboardActivity
}

type ShopSettings struct {
	ShopName, ShopSubtitle, Phone, Address, Email, Website         string
	RegistrationID, TaxID, LogoPath, DocumentFooter, DocumentNotes string
}

type PrintLine struct {
	Description, Unit            string
	QuantityUnits                int64
	UnitPriceRial, LineTotalRial int64
}
type StatementLine struct {
	Date, Reference, Description       string
	DebitRial, CreditRial, BalanceRial int64
}
type PrintAllocation struct {
	Reference, TargetType string
	AmountRial            int64
}
type PrintDocument struct {
	Kind, Number, Date, DueDate, Status, CustomerName, CustomerContact, SupplierName string
	Reference, Method, AccountName, PaymentStatus, Notes                             string
	SubtotalRial, DiscountRial, TotalRial, PaidRial, RemainingRial, AmountRial       int64
	Shop                                                                             ShopSettings
	Lines                                                                            []PrintLine
	StatementLines                                                                   []StatementLine
	Allocations                                                                      []PrintAllocation
}

func ReportRange(start, end time.Time) (time.Time, time.Time) {
	return start.UTC(), end.UTC().Add(24 * time.Hour)
}
