package demo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"Atropaten/internal/domain"
	"Atropaten/internal/platform"
	"Atropaten/internal/storage/sqlite"
)

const DefaultSeed int64 = 6006
const DefaultReferenceDate = "2026-03-21"

type Options struct {
	Root          string
	Seed          int64
	ReferenceDate string
}

type Summary struct {
	Root       string         `json:"root"`
	Seed       int64          `json:"seed"`
	Reference  string         `json:"reference_date"`
	Counts     map[string]int `json:"counts"`
	DataDigest string         `json:"data_digest"`
}

type generator struct {
	store *sqlite.Store
	root  string
	now   time.Time
	seed  int64
	ids   map[string]string
}

func Generate(ctx context.Context, options Options) (Summary, error) {
	root, err := NormalizeRoot(options.Root)
	if err != nil {
		return Summary{}, err
	}
	if err := AssertIsolatedRoot(root); err != nil {
		return Summary{}, err
	}
	if options.Seed == 0 {
		options.Seed = DefaultSeed
	}
	if options.ReferenceDate == "" {
		options.ReferenceDate = DefaultReferenceDate
	}
	date, err := time.Parse("2006-01-02", options.ReferenceDate)
	if err != nil {
		return Summary{}, fmt.Errorf("reference date must be YYYY-MM-DD: %w", err)
	}
	if _, err := os.Stat(filepath.Join(root, MarkerFile)); err == nil {
		return Summary{}, errors.New("demo root already exists; use reset explicitly before regenerating")
	} else if !os.IsNotExist(err) {
		return Summary{}, err
	}
	paths := platform.DataPaths{Root: root, Database: filepath.Join(root, "atropaten.db"), Attachments: filepath.Join(root, "attachments"), Backups: filepath.Join(root, "backups")}
	if err := paths.Ensure(); err != nil {
		return Summary{}, fmt.Errorf("create isolated demo root: %w", err)
	}
	if err := WriteMarker(root, Marker{Seed: options.Seed, ReferenceDate: options.ReferenceDate}); err != nil {
		return Summary{}, err
	}
	store, err := sqlite.Open(paths.Database)
	if err != nil {
		return Summary{}, fmt.Errorf("open demo database: %w", err)
	}
	g := &generator{store: store, root: root, now: date.UTC().Add(10 * time.Hour), seed: options.Seed, ids: map[string]string{}}
	defer store.Close()
	if err := g.populate(ctx, paths); err != nil {
		return Summary{}, fmt.Errorf("populate demo data: %w", err)
	}
	return g.summary(ctx, options.ReferenceDate)
}

func (g *generator) populate(ctx context.Context, paths platform.DataPaths) error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"settings", func() error { return g.settings(ctx) }},
		{"parties", func() error { return g.parties(ctx) }},
		{"catalog", func() error { return g.catalog(ctx) }},
		{"purchases", func() error { return g.purchases(ctx) }},
		{"orders", func() error { return g.orders(ctx) }},
		{"production", func() error { return g.production(ctx) }},
		{"finance", func() error { return g.finance(ctx) }},
		{"owners-and-periods", func() error { return g.ownersAndPeriods(ctx) }},
		{"metadata", func() error { return g.metadata(ctx, paths) }},
	}
	for _, step := range steps {
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	return nil
}

func (g *generator) settings(ctx context.Context) error {
	return g.store.SaveShopSettings(ctx, domain.ShopSettings{
		ShopName:     "آسمان چاپ نمونه",
		ShopSubtitle: "محیط توسعه نمایشی — داده‌های ساختگی",
		Phone:        "+98 21 5555 6006", Address: "تهران، خیابان نمونه، پلاک ۶۰۶",
		Email: "demo@example.invalid", Website: "https://demo.example.invalid",
		RegistrationID: "DEMO-6006", TaxID: "DEMO-TAX-1405",
		DocumentFooter: "این سند در محیط توسعه و با داده ساختگی تولید شده است.",
		DocumentNotes:  "برای بررسی چاپ، فیلترها، وضعیت‌ها و گزارش‌ها.",
	})
}

func (g *generator) parties(ctx context.Context) error {
	customers := []domain.Customer{
		{ID: "CUS-DEMO-01", Name: "نگارستان رنگین‌کمان بازرگانی شرق", Phone: "021-44006006", Email: "negar@example.invalid", Address: "تهران، شهرک غرب، بلوار آزمایشی", Notes: "مشتری نمونه با نام بلند برای تست جدول و inspector. تماس ترجیحی بعد از ساعت ۱۴.", Active: true},
		{ID: "CUS-DEMO-02", Name: "شرکت بسته‌بندی دانه‌طلایی", Phone: "021-22334455", Email: "daneh@example.invalid", Address: "کرج، شهر صنعتی شماره ۲", Notes: "سفارش‌های دوره‌ای؛ یادداشت چندخطی برای تست جزئیات.", Active: true},
		{ID: "CUS-DEMO-03", Name: "استودیو طرحِ فردا", Phone: "09120006006", Email: "studio@example.invalid", Address: "اصفهان، خیابان هنر", Notes: "مشتری با سفارش فوری و پرداخت ناقص.", Active: true},
		{ID: "CUS-DEMO-04", Name: "مشتری آرشیوی نمایشی", Phone: "", Email: "", Address: "", Notes: "رکورد غیرفعال برای تست فیلتر آرشیو.", Active: false},
	}
	for _, c := range customers {
		c.CreatedAt, c.UpdatedAt = g.now, g.now
		if err := g.store.SaveCustomer(ctx, c); err != nil {
			return err
		}
	}
	suppliers := []domain.Supplier{
		{ID: "SUP-DEMO-01", Name: "تأمین کاغذ و ورق آفتاب", Code: "SUP-AFT", Phone: "021-66006006", Email: "paper@example.invalid", Address: "تهران، انبار غرب", Notes: "تأمین‌کننده اصلی کاغذ و ورق.", Active: true},
		{ID: "SUP-DEMO-02", Name: "مواد مصرفی مرکب و روکش پارس", Code: "SUP-PARS", Phone: "021-77007700", Email: "ink@example.invalid", Address: "تهران، بازار چاپ", Notes: "برای سناریوی برون‌سپاری و خرید مرکب.", Active: true},
		{ID: "SUP-DEMO-03", Name: "خدمات برش و صحافی نمونه", Code: "SUP-FIN", Phone: "", Email: "finish@example.invalid", Address: "", Notes: "تأمین‌کننده با اطلاعات اختیاری ناقص.", Active: true},
	}
	for _, s := range suppliers {
		s.CreatedAt, s.UpdatedAt = g.now, g.now
		if err := g.store.SaveSupplier(ctx, s); err != nil {
			return err
		}
	}
	g.ids["customer"] = customers[0].ID
	g.ids["supplier"] = suppliers[0].ID
	g.ids["supplier_outsource"] = suppliers[2].ID
	return nil
}

func (g *generator) catalog(ctx context.Context) error {
	mat := []domain.Material{
		{ID: "MAT-DEMO-01", Name: "کاغذ گلاسه ۳۰۰ گرم — A3+", SKU: "PAP-A3-300", Category: "کاغذ", PurchaseUnit: "sheet", ConsumptionUnit: "sheet", ConversionFactor: domain.QuantityScale, PhysicalStock: 0, ReorderLevel: domain.Quantity(120 * domain.QuantityScale), AverageUnitCostRial: 0, PreferredSupplier: "SUP-DEMO-01", Notes: "موجودی با مقدار اعشاری در چند حرکت و آستانه سفارش بالا."},
		{ID: "MAT-DEMO-02", Name: "مرکب چهاررنگ استاندارد پروفایل‌دار", SKU: "INK-CMYK-PRO", Category: "مرکب", PurchaseUnit: "liter", ConsumptionUnit: "liter", ConversionFactor: domain.QuantityScale, ReorderLevel: domain.Quantity(6 * domain.QuantityScale), PreferredSupplier: "SUP-DEMO-02", Notes: "برای تست مقدار اعشاری و هشدار موجودی کم."},
		{ID: "MAT-DEMO-03", Name: "فیلم لمینت مات ضدخش رول ویژه", SKU: "LAM-MAT-ROLL", Category: "روکش", PurchaseUnit: "roll", ConsumptionUnit: "meter", ConversionFactor: domain.Quantity(25 * domain.QuantityScale), ReorderLevel: domain.Quantity(20 * domain.QuantityScale), PreferredSupplier: "SUP-DEMO-02", Notes: "تبدیل رول به متر؛ مقدار باقیمانده برای تست رزرو."},
		{ID: "MAT-DEMO-04", Name: "ماده نمونه صفرموجودی برای توجه داشبورد", SKU: "ZERO-ATTN", Category: "مصرفی", PurchaseUnit: "pack", ConsumptionUnit: "piece", ConversionFactor: domain.Quantity(10 * domain.QuantityScale), ReorderLevel: domain.Quantity(5 * domain.QuantityScale), PreferredSupplier: "SUP-DEMO-01", Notes: "عمداً صفر نگه داشته شده تا attention و empty/low-stock دیده شود."},
	}
	for _, m := range mat {
		m.CreatedAt, m.UpdatedAt = g.now, g.now
		m.Active = true
		if err := g.store.Create(ctx, m); err != nil {
			return err
		}
	}
	machine := []domain.Machine{
		{ID: "MAC-DEMO-01", Name: "چاپ چهاررنگ آفتاب ۷۰×۱۰۰", Code: "PRESS-4C", Category: "چاپ", RateBasis: domain.RatePerHour, RateRial: 18500000, SetupCostRial: 42000000, Notes: "ماشین فعال برای jobهای فوری."},
		{ID: "MAC-DEMO-02", Name: "برش اتوماتیک دقیق با میز بلند", Code: "CUT-AUTO", Category: "پس‌پردازش", RateBasis: domain.RatePerMinute, RateRial: 920000, SetupCostRial: 8000000, Notes: "یادداشت طولانی برای تست inspector."},
		{ID: "MAC-DEMO-03", Name: "ماشین آرشیوی نمایشی", Code: "OLD-DEMO", Category: "آرشیو", RateBasis: domain.RatePerUnit, RateRial: 0, SetupCostRial: 0, Notes: "ماشین غیرفعال."},
	}
	for i := range machine {
		machine[i].CreatedAt, machine[i].UpdatedAt = g.now, g.now
		machine[i].Active = i != 2
		if err := g.store.SaveMachine(ctx, machine[i]); err != nil {
			return err
		}
	}
	service := []domain.Service{
		g.service("SVC-DEMO-01", "چاپ پوستر رنگی تیراژ متغیر", "POSTER-X", "چاپ", "سرویسی با پارامترهای قطع و تیراژ برای تست inspector و محاسبه قیمت.", "MAT-DEMO-01", "MAC-DEMO-01"),
		g.service("SVC-DEMO-02", "کارت ویزیت لمینت مات با گوشه‌گرد", "CARD-MAT", "کارت", "سفارش ترکیبی با مصرف فیلم و خدمات تکمیلی.", "MAT-DEMO-01", "MAC-DEMO-02"),
		g.service("SVC-DEMO-03", "خدمت دستی بدون قیمت‌گذاری خودکار", "MANUAL-OPS", "خدمات", "سرویس با قیمت‌گذاری دستی برای تست وضعیت هشدار.", "MAT-DEMO-02", "MAC-DEMO-01"),
	}
	for _, s := range service {
		if err := g.store.SaveServiceDefinition(ctx, s); err != nil {
			return err
		}
	}
	g.ids["material_paper"] = mat[0].ID
	g.ids["material_ink"] = mat[1].ID
	g.ids["material_lam"] = mat[2].ID
	g.ids["service_print"] = service[0].ID
	g.ids["service_card"] = service[1].ID
	g.ids["machine"] = machine[0].ID
	return nil
}

func (g *generator) service(id, name, code, category, description, materialID, machineID string) domain.Service {
	min, max := domain.Quantity(1*domain.QuantityScale), domain.Quantity(10000*domain.QuantityScale)
	return domain.Service{ID: id, Name: name, Code: code, Category: category, Description: description, Active: true, CreatedAt: g.now, UpdatedAt: g.now,
		Parameters:  []domain.ServiceParameter{{ID: id + "-P1", ServiceID: id, Key: "run_size", Label: "تیراژ", Type: domain.ParameterInteger, Required: true, Position: 0, DefaultValue: "100", MinValue: &min, MaxValue: &max, Unit: "عدد", Active: true, CreatedAt: g.now, UpdatedAt: g.now}},
		Components:  []domain.ServiceCostComponent{{ID: id + "-C1", ServiceID: id, Name: "کاغذ/مواد اصلی", Type: domain.CostMaterial, ReferenceID: materialID, UsageMode: domain.UsageParameter, ParameterKey: "run_size", UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true, Position: 0, Notes: "مصرف وابسته به تیراژ", CreatedAt: g.now, UpdatedAt: g.now}, {ID: id + "-C2", ServiceID: id, Name: "ماشین و آماده‌سازی", Type: domain.CostMachine, ReferenceID: machineID, UsageMode: domain.UsageFixed, UsageQuantity: domain.QuantityScale, Multiplier: domain.QuantityScale, Enabled: true, Position: 1, Notes: "هزینه setup", CreatedAt: g.now, UpdatedAt: g.now}, {ID: id + "-C3", ServiceID: id, Name: "پرت و سربار", Type: domain.CostOverhead, UsageMode: domain.UsageFixed, Percentage: domain.Quantity(7 * domain.QuantityScale), Multiplier: domain.QuantityScale, Enabled: true, Position: 2, Notes: "درصد نمایشی", CreatedAt: g.now, UpdatedAt: g.now}},
		PricingRule: &domain.ServicePricingRule{ID: id + "-R1", ServiceID: id, Type: domain.PricingMarkup, MarkupPercentage: domain.Quantity(35 * domain.QuantityScale), CreatedAt: g.now, UpdatedAt: g.now}}
}

func (g *generator) purchases(ctx context.Context) error {
	purchases := []struct {
		id, supplier, invoice string
		date                  time.Time
		items                 []domain.PurchaseItem
		note                  string
	}{
		{"PUR-DEMO-01", "SUP-DEMO-01", "FAKE-1405-001", g.now.AddDate(0, -2, -4), []domain.PurchaseItem{{ID: "PIT-DEMO-01", MaterialID: "MAT-DEMO-01", MaterialNameSnapshot: "کاغذ گلاسه ۳۰۰ گرم — A3+", PurchaseUnitSnapshot: "sheet", ConsumptionUnitSnapshot: "sheet", PurchaseQuantity: domain.Quantity(420 * domain.QuantityScale), ConversionFactorSnapshot: domain.QuantityScale, ConsumptionQuantity: domain.Quantity(420 * domain.QuantityScale), UnitAcquisitionCostRial: 1850000, LineTotalRial: 777000000}}, "خرید ثبت‌شده برای موجودی اولیه و تست بهای تمام‌شده."},
		{"PUR-DEMO-02", "SUP-DEMO-02", "FAKE-1405-014", g.now.AddDate(0, -1, 8), []domain.PurchaseItem{{ID: "PIT-DEMO-02", MaterialID: "MAT-DEMO-02", MaterialNameSnapshot: "مرکب چهاررنگ استاندارد پروفایل‌دار", PurchaseUnitSnapshot: "liter", ConsumptionUnitSnapshot: "liter", PurchaseQuantity: domain.Quantity(18*domain.QuantityScale + 500000), ConversionFactorSnapshot: domain.QuantityScale, ConsumptionQuantity: domain.Quantity(18*domain.QuantityScale + 500000), UnitAcquisitionCostRial: 12800000, LineTotalRial: 236800000}}, "خرید مرکب با مقدار ۱۸.۵ لیتر برای تست اعشار."},
		{"PUR-DEMO-03", "SUP-DEMO-02", "FAKE-1405-027", g.now.AddDate(0, -1, 18), []domain.PurchaseItem{{ID: "PIT-DEMO-03", MaterialID: "MAT-DEMO-03", MaterialNameSnapshot: "فیلم لمینت مات ضدخش رول ویژه", PurchaseUnitSnapshot: "roll", ConsumptionUnitSnapshot: "meter", PurchaseQuantity: domain.Quantity(3 * domain.QuantityScale), ConversionFactorSnapshot: domain.Quantity(25 * domain.QuantityScale), ConsumptionQuantity: domain.Quantity(75 * domain.QuantityScale), UnitAcquisitionCostRial: 95000000, LineTotalRial: 285000000}}, "خرید رول لمینت با تبدیل واحد."},
	}
	for _, x := range purchases {
		p := domain.Purchase{ID: x.id, SupplierID: x.supplier, SupplierNameSnapshot: x.supplier, SupplierCodeSnapshot: x.supplier, SupplierInvoiceNumber: x.invoice, PurchaseDate: x.date, Status: domain.PurchaseDraft, Notes: x.note, CreatedAt: g.now, UpdatedAt: g.now, Items: x.items}
		for i := range p.Items {
			p.Items[i].PurchaseID = p.ID
			p.Items[i].Position = i
			p.Items[i].AllocatedAdditionalCostRial = 0
			p.Items[i].LandedUnitCostRial = p.Items[i].UnitAcquisitionCostRial
		}
		p.SubtotalRial = x.items[0].LineTotalRial
		p.TotalRial = p.SubtotalRial
		if err := g.store.SavePurchase(ctx, p); err != nil {
			return err
		}
		if err := g.store.PostPurchase(ctx, p.ID); err != nil {
			return err
		}
	}
	if err := g.store.AdjustInventory(ctx, "MAT-DEMO-01", domain.Quantity(-37*domain.QuantityScale), 0, "مصرف دستی نمونه برای ایجاد مقدار باقی‌مانده و audit trail"); err != nil {
		return err
	}
	return nil
}

func demoJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

func (g *generator) item(id, orderID, serviceID, name, code string, qty domain.Quantity, cost, price int64, note string) domain.OrderItem {
	return domain.OrderItem{ID: id, OrderID: orderID, Position: 0, ServiceID: serviceID, ServiceNameSnapshot: name, ServiceCodeSnapshot: code, Quantity: qty, QuantityUnit: "piece", ResolvedParametersJSON: demoJSON(map[string]string{"run_size": qty.String()}), CostBreakdownJSON: demoJSON([]map[string]any{{"name": "مواد و ماشین", "amount_rial": cost}}), PricingSnapshotJSON: demoJSON(map[string]any{"seed": g.seed, "price_rial": price}), EstimatedCostRial: cost, SuggestedPriceRial: price, SellingPriceRial: price, Notes: note}
}

func (g *generator) orders(ctx context.Context) error {
	orders := []struct {
		id, customer, status, fulfillment, priority, note string
		promised                                          time.Time
		service, name, code                               string
		qty                                               domain.Quantity
		cost, price, discount                             int64
	}{
		{"ORD-DEMO-01", "CUS-DEMO-01", "", string(domain.CommercialConfirmed), string(domain.FulfillmentPending), string(domain.PriorityUrgent), "سفارش فوری با job در حال تولید و رزرو فعال.", g.now.AddDate(0, 0, -2), "SVC-DEMO-01", "چاپ پوستر رنگی تیراژ متغیر", "POSTER-X", domain.Quantity(120 * domain.QuantityScale), 180000000, 420000000, 10000000},
		{"ORD-DEMO-02", "CUS-DEMO-02", "", string(domain.CommercialConfirmed), string(domain.FulfillmentPending), string(domain.PriorityHigh), "سفارش آماده تحویل پس از اتمام تولید.", g.now.AddDate(0, 0, -8), "SVC-DEMO-02", "کارت ویزیت لمینت مات با گوشه‌گرد", "CARD-MAT", domain.Quantity(2500 * domain.QuantityScale), 310000000, 690000000, 0},
		{"ORD-DEMO-03", "CUS-DEMO-03", string(domain.CommercialClosed), string(domain.FulfillmentDelivered), string(domain.PriorityNormal), "سفارش تحویل‌شده با فاکتور پرداخت‌شده.", g.now.AddDate(0, -1, -6), "SVC-DEMO-01", "چاپ پوستر رنگی تیراژ متغیر", "POSTER-X", domain.Quantity(500 * domain.QuantityScale), 260000000, 1280000000, 80000000},
		{"ORD-DEMO-04", "CUS-DEMO-01", "", string(domain.CommercialDraft), string(domain.FulfillmentPending), string(domain.PriorityLow), "پیش‌نویس با یادداشت طولانی برای تست صفحه‌بندی.", g.now.AddDate(0, 0, 25), "SVC-DEMO-03", "خدمت دستی بدون قیمت‌گذاری خودکار", "MANUAL-OPS", domain.Quantity(3 * domain.QuantityScale), 90000000, 150000000, 0},
	}
	for _, x := range orders {
		o := domain.NewOrder(x.id, x.customer, g.now.AddDate(0, 0, -10))
		o.CustomerNameSnapshot = map[string]string{"CUS-DEMO-01": "نگارستان رنگین‌کمان بازرگانی شرق", "CUS-DEMO-02": "شرکت بسته‌بندی دانه‌طلایی", "CUS-DEMO-03": "استودیو طرحِ فردا"}[x.customer]
		o.CustomerPhoneSnapshot = map[string]string{"CUS-DEMO-01": "021-44006006", "CUS-DEMO-02": "021-22334455", "CUS-DEMO-03": "09120006006"}[x.customer]
		o.Notes = x.note
		o.Priority = domain.Priority(x.priority)
		o.PromisedAt = &x.promised
		o.Items = []domain.OrderItem{g.item(x.id+"-ITEM", x.id, x.service, x.name, x.code, x.qty, x.cost, x.price, "اقلام نمونه برای تست inspector")}
		o.DiscountRial = x.discount
		if err := o.RecalculateTotals(); err != nil {
			return err
		}
		if err := g.store.CreateOrder(ctx, o); err != nil {
			return err
		}
		saved, err := g.store.GetOrder(ctx, x.id)
		if err != nil {
			return err
		}
		saved.CommercialStatus = domain.CommercialStatus(x.status)
		saved.FulfillmentStatus = domain.FulfillmentStatus(x.fulfillment)
		saved.UpdatedAt = g.now
		if err := g.store.SaveOrder(ctx, saved); err != nil {
			return err
		}
	}
	g.ids["order_active"] = "ORD-DEMO-01"
	g.ids["order_ready"] = "ORD-DEMO-02"
	g.ids["order_paid"] = "ORD-DEMO-03"
	return nil
}

func (g *generator) production(ctx context.Context) error {
	err := g.store.CreateProductionJob(ctx, domain.ProductionJob{ID: "JOB-DEMO-01", OrderID: "ORD-DEMO-01", OrderItemID: "ORD-DEMO-01-ITEM", Quantity: domain.Quantity(120 * domain.QuantityScale), QuantityUnit: "piece", AssignedMachineID: "MAC-DEMO-01", Status: domain.ProductionPending, Priority: string(domain.PriorityUrgent), Notes: "در حال تولید؛ یک رزرو فعال برای تست داشبورد.", CreatedAt: g.now, UpdatedAt: g.now, PlannedAt: func() *time.Time { v := g.now.Add(-24 * time.Hour); return &v }()})
	if err != nil {
		return err
	}
	if err := g.store.TransitionProductionJob(ctx, "JOB-DEMO-01", domain.ProductionInProgress); err != nil {
		return err
	}
	if err := g.store.CreateReservation(ctx, domain.InventoryReservation{ID: "RES-DEMO-01", MaterialID: "MAT-DEMO-01", OrderID: "ORD-DEMO-01", OrderItemID: "ORD-DEMO-01-ITEM", ProductionJobID: "JOB-DEMO-01", Quantity: domain.Quantity(70 * domain.QuantityScale), Status: domain.ReservationActive, CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	if _, err := g.store.RecordProductionConsumption(ctx, "JOB-DEMO-01", "MAT-DEMO-01", "CON-DEMO-01", domain.Quantity(35*domain.QuantityScale), domain.Quantity(1*domain.QuantityScale/2), "مصرف بخشی با پرت اعشاری برای تست cost breakdown"); err != nil {
		return err
	}
	err = g.store.CreateProductionJob(ctx, domain.ProductionJob{ID: "JOB-DEMO-02", OrderID: "ORD-DEMO-02", OrderItemID: "ORD-DEMO-02-ITEM", Quantity: domain.Quantity(2500 * domain.QuantityScale), QuantityUnit: "piece", AssignedMachineID: "MAC-DEMO-02", Status: domain.ProductionPending, Priority: string(domain.PriorityHigh), Notes: "تکمیل‌شده و آماده تحویل.", CreatedAt: g.now, UpdatedAt: g.now})
	if err != nil {
		return err
	}
	if err := g.store.TransitionProductionJob(ctx, "JOB-DEMO-02", domain.ProductionInProgress); err != nil {
		return err
	}
	if err := g.store.CreateReservation(ctx, domain.InventoryReservation{ID: "RES-DEMO-02", MaterialID: "MAT-DEMO-01", OrderID: "ORD-DEMO-02", OrderItemID: "ORD-DEMO-02-ITEM", ProductionJobID: "JOB-DEMO-02", Quantity: domain.Quantity(90 * domain.QuantityScale), Status: domain.ReservationActive, CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	if _, err := g.store.RecordProductionConsumption(ctx, "JOB-DEMO-02", "MAT-DEMO-01", "CON-DEMO-02", domain.Quantity(60*domain.QuantityScale), domain.Quantity(1*domain.QuantityScale), "پرت تکمیل‌شده"); err != nil {
		return err
	}
	if err := g.store.TransitionProductionJob(ctx, "JOB-DEMO-02", domain.ProductionCompleted); err != nil {
		return err
	}
	return nil
}

func (g *generator) finance(ctx context.Context) error {
	for _, x := range []struct {
		id, order, customer string
		amount              int64
		paid                time.Time
	}{
		{"INV-DEMO-01", "ORD-DEMO-03", "CUS-DEMO-03", 1200000000, g.now.AddDate(0, -1, -5)},
		{"INV-DEMO-02", "ORD-DEMO-01", "CUS-DEMO-01", 410000000, g.now.AddDate(0, 0, -1)},
	} {
		o, err := g.store.GetOrder(ctx, x.order)
		if err != nil {
			return err
		}
		inv := domain.Invoice{ID: x.id, CustomerID: x.customer, CustomerNameSnapshot: o.CustomerNameSnapshot, CustomerPhoneSnapshot: o.CustomerPhoneSnapshot, OrderID: o.ID, IssueDate: x.paid, DueDate: func() *time.Time { v := x.paid.AddDate(0, 0, 14); return &v }(), Status: domain.InvoiceDraft, Notes: "فاکتور نمایشی با تاریخ سررسید برای print preview.", SubtotalRial: x.amount, TotalRial: x.amount, CreatedAt: x.paid, UpdatedAt: x.paid, Items: []domain.InvoiceItem{{ID: x.id + "-ITEM", InvoiceID: x.id, Position: 0, OrderItemID: o.Items[0].ID, DescriptionSnapshot: o.Items[0].ServiceNameSnapshot, ServiceID: o.Items[0].ServiceID, QuantityUnits: int64(o.Items[0].Quantity), QuantityUnit: o.Items[0].QuantityUnit, UnitPriceRial: x.amount, LineTotalRial: x.amount, Notes: "خط فاکتور نمایشی"}}}
		if err := g.store.SaveInvoice(ctx, inv); err != nil {
			return err
		}
		if err := g.store.PostInvoice(ctx, x.id); err != nil {
			return err
		}
	}
	if _, err := g.store.CreatePayment(ctx, domain.Payment{ID: "PAY-DEMO-01", Direction: domain.PaymentIncoming, Method: domain.PaymentBankTransfer, FinancialAccountID: "FIN-BANK", CustomerID: "CUS-DEMO-03", AmountRial: 1200000000, PostedAt: g.now.AddDate(0, -1, -4), Reference: "DEMO-SETTLED", Notes: "پرداخت کامل فاکتور نمونه", Status: domain.PaymentPosted, IdempotencyKey: "PAY-DEMO-01", CreatedAt: g.now, Allocations: []domain.PaymentAllocation{{ID: "PAL-DEMO-01", PaymentID: "PAY-DEMO-01", Position: 0, TargetType: "invoice", TargetID: "INV-DEMO-01", AmountRial: 1200000000}}}); err != nil {
		return err
	}
	if _, err := g.store.CreatePayment(ctx, domain.Payment{ID: "PAY-DEMO-02", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", CustomerID: "CUS-DEMO-01", AmountRial: 150000000, PostedAt: g.now.AddDate(0, 0, -1), Reference: "DEMO-PARTIAL", Notes: "پرداخت بخشی برای badge و summary", Status: domain.PaymentPosted, IdempotencyKey: "PAY-DEMO-02", CreatedAt: g.now, Allocations: []domain.PaymentAllocation{{ID: "PAL-DEMO-02", PaymentID: "PAY-DEMO-02", Position: 0, TargetType: "invoice", TargetID: "INV-DEMO-02", AmountRial: 150000000}}}); err != nil {
		return err
	}
	if _, err := g.store.CreateExpense(ctx, domain.Expense{ID: "EXP-DEMO-01", ExpenseDate: g.now.AddDate(0, 0, -7), CategoryAccountID: "ACC-EXP-UTILITIES", Payee: "شرکت خدمات شهری نمونه", Description: "قبض برق و نگهداری کارگاه نمایشی", PaymentMethod: string(domain.PaymentBankTransfer), FinancialAccountID: "FIN-BANK", Notes: "هزینه ساختگی برای گزارش هزینه‌ها.", AmountRial: 48500000, Status: "Posted", IdempotencyKey: "EXP-DEMO-01", CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	if _, err := g.store.CreateTransfer(ctx, domain.FinancialTransfer{ID: "TRF-DEMO-01", SourceFinancialAccountID: "FIN-BANK", DestinationFinancialAccountID: "FIN-CASH", AmountRial: 75000000, TransferDate: g.now.AddDate(0, 0, -3), Reference: "DEMO-CASH-FLOAT", Notes: "انتقال نمونه بین حساب‌ها", Status: "Posted", IdempotencyKey: "TRF-DEMO-01", CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	return g.checksLoans(ctx)
}

func (g *generator) checksLoans(ctx context.Context) error {
	if _, err := g.store.CreateCheck(ctx, domain.Check{ID: "CHK-DEMO-01", CheckNumber: "CHK-DEMO-001", Direction: domain.CheckIncoming, Bank: "بانک نمونه", Branch: "شعبه مرکزی", AccountDescriptor: "IR-DEMO-6006", AmountRial: 325000000, IssueDate: g.now.AddDate(0, 0, -12), DueDate: g.now.AddDate(0, 0, -2), PayerPayee: "نگارستان رنگین‌کمان بازرگانی شرق", CustomerID: "CUS-DEMO-01", FinancialAccountID: "FIN-BANK", Notes: "چک سررسیدشده برای attention داشبورد.", Status: domain.CheckDraft, CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	if _, err := g.store.ChangeCheckStatus(ctx, "CHK-DEMO-01", domain.CheckReceived, "تحویل از مشتری نمونه", "CHK-DEMO-01-RECEIVED"); err != nil {
		return err
	}
	if _, err := g.store.ChangeCheckStatus(ctx, "CHK-DEMO-01", domain.CheckDeposited, "در انتظار تسویه", "CHK-DEMO-01-DEPOSITED"); err != nil {
		return err
	}
	if _, err := g.store.CreateCheck(ctx, domain.Check{ID: "CHK-DEMO-02", CheckNumber: "CHK-DEMO-002", Direction: domain.CheckOutgoing, Bank: "بانک نمونه", Branch: "شعبه تأمین‌کننده", AccountDescriptor: "IR-DEMO-6006", AmountRial: 98000000, IssueDate: g.now.AddDate(0, 0, -4), DueDate: g.now.AddDate(0, 0, 10), PayerPayee: "تأمین کاغذ و ورق آفتاب", SupplierID: "SUP-DEMO-01", FinancialAccountID: "FIN-BANK", Notes: "چک پرداختنی سررسید آینده.", Status: domain.CheckDraft, CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	if _, err := g.store.ChangeCheckStatus(ctx, "CHK-DEMO-02", domain.CheckIssued, "صادرشده برای خرید کاغذ", "CHK-DEMO-02-ISSUED"); err != nil {
		return err
	}
	loanDate := g.now.AddDate(0, -2, -10)
	if _, err := g.store.CreateLoan(ctx, domain.Loan{ID: "LOAN-DEMO-01", Direction: domain.LoanPayable, CounterpartyName: "بانک توسعه ساختگی", SupplierID: "SUP-DEMO-02", PrincipalRial: 2400000000, InterestFeeRial: 180000000, StartDate: loanDate, EndDate: func() *time.Time { v := g.now.AddDate(0, 1, 20); return &v }(), Status: domain.LoanActive, Notes: "وام پرداختنی نمایشی با قسط عقب‌افتاده و قسط آینده.", FinancialAccountID: "FIN-BANK", IdempotencyKey: "LOAN-DEMO-01", CreatedAt: g.now, UpdatedAt: g.now, Installments: []domain.LoanInstallment{{ID: "LOAN-DEMO-01-I1", LoanID: "LOAN-DEMO-01", Position: 0, DueDate: g.now.AddDate(0, 0, -20), PrincipalRial: 800000000, InterestFeeRial: 60000000, TotalDueRial: 860000000}, {ID: "LOAN-DEMO-01-I2", LoanID: "LOAN-DEMO-01", Position: 1, DueDate: g.now.AddDate(0, 0, 40), PrincipalRial: 800000000, InterestFeeRial: 60000000, TotalDueRial: 860000000}, {ID: "LOAN-DEMO-01-I3", LoanID: "LOAN-DEMO-01", Position: 2, DueDate: g.now.AddDate(0, 2, 10), PrincipalRial: 800000000, InterestFeeRial: 60000000, TotalDueRial: 860000000}}}); err != nil {
		return err
	}
	return nil
}

func (g *generator) ownersAndPeriods(ctx context.Context) error {
	owners := []domain.Owner{{ID: "OWN-DEMO-01", Name: "سارا نیک‌فر", Phone: "09121112233", Email: "sara@example.invalid", Notes: "مالک اصلی نمونه.", Active: true, OwnershipBPS: 6000, ProfitSharingBPS: 6000, CreatedAt: g.now, UpdatedAt: g.now}, {ID: "OWN-DEMO-02", Name: "کامران دادخواه", Phone: "09124445566", Email: "kamran@example.invalid", Notes: "مالک دوم نمونه.", Active: true, OwnershipBPS: 4000, ProfitSharingBPS: 4000, CreatedAt: g.now, UpdatedAt: g.now}}
	for _, o := range owners {
		if _, err := g.store.CreateOwner(ctx, o); err != nil {
			return err
		}
	}
	if _, err := g.store.CreateOwnerTransaction(ctx, domain.OwnerTransaction{ID: "OTX-DEMO-01", OwnerID: "OWN-DEMO-01", Type: domain.OwnerTxCapitalContribution, FinancialAccountID: "FIN-BANK", AmountRial: 3500000000, OccurredAt: g.now.AddDate(0, -2, 0), Description: "آورده اولیه مالک", Notes: "سرمایه ساختگی برای تراز آزمایشی.", Status: domain.OwnerTxPosted, IdempotencyKey: "OTX-DEMO-01", CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	if _, err := g.store.CreateOwnerTransaction(ctx, domain.OwnerTransaction{ID: "OTX-DEMO-02", OwnerID: "OWN-DEMO-02", Type: domain.OwnerTxDrawing, FinancialAccountID: "FIN-CASH", AmountRial: 120000000, OccurredAt: g.now.AddDate(0, 0, -12), Description: "برداشت مالک", Notes: "برای تنوع گردش حساب.", Status: domain.OwnerTxPosted, IdempotencyKey: "OTX-DEMO-02", CreatedAt: g.now, UpdatedAt: g.now}); err != nil {
		return err
	}
	_, err := g.store.CreateFiscalPeriod(ctx, domain.FiscalPeriod{ID: "FP-DEMO-1405", Name: "سال مالی نمایشی ۱۴۰۵", StartDate: time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2027, 3, 20, 0, 0, 0, 0, time.UTC), Notes: "دوره باز برای تست گزارش، پیش‌نمایش تخصیص و بستن دوره.", Status: domain.FiscalPeriodOpen, IdempotencyKey: "FP-DEMO-1405", CreatedAt: g.now, UpdatedAt: g.now})
	return err
}

func (g *generator) metadata(ctx context.Context, paths platform.DataPaths) error {
	attachmentDir := filepath.Join(paths.Attachments, "demo")
	if err := os.MkdirAll(attachmentDir, 0o700); err != nil {
		return err
	}
	file := filepath.Join(attachmentDir, "artwork-demo.txt")
	contents := []byte("Fictional Atropaten development demo artwork placeholder.\n")
	if err := os.WriteFile(file, contents, 0o600); err != nil {
		return err
	}
	sum := sha256.Sum256(contents)
	size := int64(len(contents))
	if err := g.store.SaveAttachment(ctx, domain.Attachment{ID: "ATT-DEMO-01", OwnerType: domain.AttachmentOrder, OwnerID: "ORD-DEMO-01", FileName: "artwork-demo.txt", Path: file, MIMEType: "text/plain", SizeBytes: &size, Checksum: hex.EncodeToString(sum[:]), Category: domain.AttachmentArtwork, Notes: "فایل کوچک و ساختگی برای تست managed attachment و proof.", CreatedAt: g.now}); err != nil {
		return err
	}
	return g.store.SaveProof(ctx, domain.Proof{ID: "PRF-DEMO-01", OwnerType: domain.AttachmentOrder, OwnerID: "ORD-DEMO-01", AttachmentID: "ATT-DEMO-01", Status: domain.ProofWaitingApproval, VersionLabel: "v2 — انتظار تأیید", PreparedAt: func() *time.Time { v := g.now.AddDate(0, 0, -1); return &v }(), ApproverNote: "منتظر تأیید مشتری برای رنگ نهایی.", InternalNote: "proof ساختگی برای تست tab و badge.", CreatedAt: g.now})
}

func (g *generator) summary(ctx context.Context, reference string) (Summary, error) {
	counts := map[string]int{}
	readers := []struct {
		name string
		read func() (int, error)
	}{
		{"customers", func() (int, error) { v, e := g.store.ListCustomers(ctx, true); return len(v), e }}, {"suppliers", func() (int, error) { v, e := g.store.ListSuppliers(ctx, true); return len(v), e }}, {"materials", func() (int, error) { v, e := g.store.List(ctx, true); return len(v), e }}, {"services", func() (int, error) { v, e := g.store.ListServices(ctx, true); return len(v), e }}, {"machines", func() (int, error) { v, e := g.store.ListMachines(ctx, true); return len(v), e }}, {"purchases", func() (int, error) { v, e := g.store.ListPurchases(ctx); return len(v), e }}, {"orders", func() (int, error) { v, e := g.store.ListOrders(ctx); return len(v), e }}, {"production", func() (int, error) { v, e := g.store.ListProductionJobs(ctx, ""); return len(v), e }}, {"invoices", func() (int, error) { v, e := g.store.ListInvoices(ctx); return len(v), e }}, {"payments", func() (int, error) { v, e := g.store.ListPayments(ctx); return len(v), e }}, {"expenses", func() (int, error) { v, e := g.store.ListExpenses(ctx); return len(v), e }}, {"transfers", func() (int, error) { v, e := g.store.ListTransfers(ctx); return len(v), e }}, {"checks", func() (int, error) { v, e := g.store.ListChecks(ctx, "", ""); return len(v), e }}, {"loans", func() (int, error) { v, e := g.store.ListLoans(ctx, "", ""); return len(v), e }}, {"owners", func() (int, error) { v, e := g.store.ListOwners(ctx, true); return len(v), e }}, {"periods", func() (int, error) { v, e := g.store.ListFiscalPeriods(ctx); return len(v), e }},
	}
	for _, r := range readers {
		n, e := r.read()
		if e != nil {
			return Summary{}, e
		}
		counts[r.name] = n
	}
	start := g.now.AddDate(0, -3, 0)
	end := g.now.AddDate(0, 1, 0)
	for _, kind := range []string{"profit_loss", "cash_bank", "receivables", "payables", "expenses", "inventory", "sales_by_service", "customer_sales", "material_usage", "production"} {
		if _, err := g.store.Report(ctx, kind, start, end); err != nil {
			return Summary{}, fmt.Errorf("build %s report: %w", kind, err)
		}
	}
	if _, err := g.store.Dashboard(ctx, start, end); err != nil {
		return Summary{}, fmt.Errorf("build dashboard: %w", err)
	}
	if _, err := g.store.PrintDocument(ctx, "invoice", "INV-DEMO-01", "", "", ""); err != nil {
		return Summary{}, fmt.Errorf("build invoice print document: %w", err)
	}
	encoded, _ := json.Marshal(counts)
	digest := sha256.Sum256(append([]byte(fmt.Sprintf("%d:%s:", g.seed, reference)), encoded...))
	return Summary{Root: g.root, Seed: g.seed, Reference: reference, Counts: counts, DataDigest: hex.EncodeToString(digest[:])}, nil
}
