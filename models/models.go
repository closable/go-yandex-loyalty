package models

type (
	OrdersDB struct {
		OrderNumber string
		Status      string
		Accrual     float64
		UploadAt    string
	}
	Orders struct {
		Number   string  `json:"number"`
		Status   string  `json:"status"`
		Accrual  float64 `json:"accrual"`
		UploadAt string  `json:"upload_at"`
	}
	WithdrawDB struct {
		Current   float64
		Withdrawn float64
	}
	WithdrawGet struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	Withdraw struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
	WithdrawGetDB struct {
		Order       string
		Sum         float64
		ProcessedAt string
	}
	AccrualGet struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}
)
