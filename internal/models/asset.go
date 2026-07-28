package models

import "time"

// AssetCategory represents a category of assets (KIB classification)
type AssetCategory struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssetFundingSource represents the source of funding for assets
type AssetFundingSource struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssetLocation represents a location or room where assets are stored
type AssetLocation struct {
	ID         int        `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Capacity   int        `json:"capacity"`
	PICStaffID *int       `json:"pic_staff_id"`
	PICName    string     `json:"pic_name,omitempty"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

// Asset represents a school asset/inventory item
type Asset struct {
	ID              int        `json:"id"`
	InventoryCode   string     `json:"inventory_code"`
	QRCode          string     `json:"qr_code"`
	CategoryID      int        `json:"category_id"`
	CategoryName    string     `json:"category_name,omitempty"`
	Name            string     `json:"name"`
	Specification   string     `json:"specification"`
	Brand           string     `json:"brand"`
	SerialNumber    string     `json:"serial_number"`
	AcquisitionYear int        `json:"acquisition_year"`
	AcquisitionDate *time.Time `json:"acquisition_date"`
	PurchasePrice   float64    `json:"purchase_price"`
	FundingSourceID *int       `json:"funding_source_id"`
	FundingSource   string     `json:"funding_source,omitempty"`
	LocationID      *int       `json:"location_id"`
	LocationName    string     `json:"location_name,omitempty"`
	PICStaffID      *int       `json:"pic_staff_id"`
	PICName         string     `json:"pic_name,omitempty"`
	Condition       string     `json:"condition"`
	Quantity        int        `json:"quantity"`
	Unit            string     `json:"unit"`
	Notes           string     `json:"notes"`
	PhotoURL        string     `json:"photo_url"`
	IsBorrowable    bool       `json:"is_borrowable"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	CreatedBy       *int       `json:"created_by"`
}

// AssetBorrowing represents a borrowing transaction
type AssetBorrowing struct {
	ID              int        `json:"id"`
	AssetID         int        `json:"asset_id"`
	AssetName       string     `json:"asset_name,omitempty"`
	BorrowerType    string     `json:"borrower_type"`
	BorrowerID      int        `json:"borrower_id"`
	BorrowerName    string     `json:"borrower_name,omitempty"`
	Purpose         string     `json:"purpose"`
	BorrowDate      time.Time  `json:"borrow_date"`
	DueDate         time.Time  `json:"due_date"`
	ReturnDate      *time.Time `json:"return_date"`
	Status          string     `json:"status"`
	ApprovedBy      *int       `json:"approved_by"`
	ApprovedByName  string     `json:"approved_by_name,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at"`
	RejectionReason string     `json:"rejection_reason"`
	ReturnCondition string     `json:"return_condition"`
	ReturnNotes     string     `json:"return_notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

// AssetMaintenanceTicket represents a maintenance or damage report ticket
type AssetMaintenanceTicket struct {
	ID              int        `json:"id"`
	TicketNumber    string     `json:"ticket_number"`
	AssetID         *int       `json:"asset_id"`
	AssetName       string     `json:"asset_name,omitempty"`
	LocationID      *int       `json:"location_id"`
	LocationName    string     `json:"location_name,omitempty"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	ReportedBy      int        `json:"reported_by"`
	ReporterName    string     `json:"reporter_name,omitempty"`
	ReportedAt      time.Time  `json:"reported_at"`
	AssignedTo      *int       `json:"assigned_to"`
	AssignedToName  string     `json:"assigned_to_name,omitempty"`
	PhotoURLs       string     `json:"photo_urls"`
	ResolutionNotes string     `json:"resolution_notes"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	ClosedAt        *time.Time `json:"closed_at"`
	Cost            float64    `json:"cost"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

// AssetMovement represents a movement of asset from one location to another
type AssetMovement struct {
	ID             int       `json:"id"`
	AssetID        int       `json:"asset_id"`
	AssetName      string    `json:"asset_name,omitempty"`
	FromLocationID *int      `json:"from_location_id"`
	FromLocation   string    `json:"from_location,omitempty"`
	ToLocationID   int       `json:"to_location_id"`
	ToLocation     string    `json:"to_location,omitempty"`
	MovedBy        int       `json:"moved_by"`
	MovedByName    string    `json:"moved_by_name,omitempty"`
	MovedAt        time.Time `json:"moved_at"`
	Reason         string    `json:"reason"`
	Notes          string    `json:"notes"`
}

// AssetDisposal represents the disposal/removal of an asset from inventory
type AssetDisposal struct {
	ID            int        `json:"id"`
	AssetID       int        `json:"asset_id"`
	AssetName     string     `json:"asset_name,omitempty"`
	DisposalDate  time.Time  `json:"disposal_date"`
	Reason        string     `json:"reason"`
	Description   string     `json:"description"`
	BookValue     float64    `json:"book_value"`
	DisposalValue float64    `json:"disposal_value"`
	ApprovedBy    *int       `json:"approved_by"`
	ApprovedByName string    `json:"approved_by_name,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at"`
	DocumentURL   string     `json:"document_url"`
	CreatedBy     int        `json:"created_by"`
	CreatedByName string     `json:"created_by_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AssetLabelTemplate represents a template for printing asset labels
type AssetLabelTemplate struct {
	ID                    int        `json:"id"`
	Name                  string     `json:"name"`
	PaperSize             string     `json:"paper_size"`
	Layout                string     `json:"layout"`
	IncludeLogo           bool       `json:"include_logo"`
	IncludeSchoolName     bool       `json:"include_school_name"`
	IncludeQRCode         bool       `json:"include_qr_code"`
	IncludeInventoryCode  bool       `json:"include_inventory_code"`
	IncludeAssetName      bool       `json:"include_asset_name"`
	IncludeFundingSource  bool       `json:"include_funding_source"`
	QRSize                int        `json:"qr_size"`
	FontSize              int        `json:"font_size"`
	IsDefault             bool       `json:"is_default"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at"`
}

// AssetStockOpname represents a stock taking/inventory audit session
type AssetStockOpname struct {
	ID          int        `json:"id"`
	SessionName string     `json:"session_name"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string     `json:"status"`
	CreatedBy   int        `json:"created_by"`
	CreatedByName string   `json:"created_by_name,omitempty"`
	Notes       string     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// AssetStockOpnameItem represents individual items in a stock opname session
type AssetStockOpnameItem struct {
	ID                 int        `json:"id"`
	OpnameID           int        `json:"opname_id"`
	AssetID            int        `json:"asset_id"`
	AssetName          string     `json:"asset_name,omitempty"`
	ExpectedLocationID *int       `json:"expected_location_id"`
	ExpectedLocation   string     `json:"expected_location,omitempty"`
	ActualLocationID   *int       `json:"actual_location_id"`
	ActualLocation     string     `json:"actual_location,omitempty"`
	ExpectedCondition  string     `json:"expected_condition"`
	ActualCondition    string     `json:"actual_condition"`
	Status             string     `json:"status"`
	ScannedBy          *int       `json:"scanned_by"`
	ScannedByName      string     `json:"scanned_by_name,omitempty"`
	ScannedAt          *time.Time `json:"scanned_at"`
	Notes              string     `json:"notes"`
}
