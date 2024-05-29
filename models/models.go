// Пакет консолидации моделей приложения
package models

type (
	// Структура заказ
	OrdersDB struct {
		// Заказ
		OrderNumber string
		// Статус
		Status string
		// Кол-во баллов
		Accrual float32
		// Загружено
		UploadAt string
	}
	// Структура запроса заказа
	Orders struct {
		// Заказ
		Number string `json:"number"`
		// Статус
		Status string `json:"status"`
		// Кол-во баллов
		Accrual float32 `json:"accrual"`
		// Загоужено
		UploadAt string `json:"upload_at"`
	}
	// Структра баланса
	WithdrawDB struct {
		// Текущий
		Current float32
		// Всего баллов
		Withdrawn float32
	}
	//Структура запроса списания
	WithdrawGet struct {
		// Заказ
		Order string `json:"order"`
		// Сумма списания
		Sum float32 `json:"sum"`
	}
	// Структура едиицы хранения списаний
	Withdraw struct {
		//Заказ
		Order string `json:"order"`
		// Сумма
		Sum float32 `json:"sum"`
		// Обработано
		ProcessedAt string `json:"processed_at"`
	}
	WithdrawGetDB struct {
		//Заказ
		Order string
		//Суммаа
		Sum float32
		//Обработано
		ProcessedAt string
	}
	// Запрос состояния
	AccrualGet struct {
		// Заказ
		Order string `json:"order"`
		// Статус
		Status string `json:"status"`
		// Сумма
		Accrual float32 `json:"accrual"`
	}
)
