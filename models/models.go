package models

type (
	OrdersDB struct {
		OrderNumber string
		Status      string
		Accrual     float32
		UploadAt    string
	}
	Orders struct {
		Number   string  `json:"number"`
		Status   string  `json:"status"`
		Accrual  float32 `json:"accrual"`
		UploadAt string  `json:"upload_at"`
	}
	WithdrawDB struct {
		Current   float32
		Withdrawn float32
	}
	WithdrawGet struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
	Withdraw struct {
		Order       string  `json:"order"`
		Sum         float32 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
	WithdrawGetDB struct {
		Order       string
		Sum         float32
		ProcessedAt string
	}
	AccrualGet struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float32 `json:"accrual"`
	}
)
