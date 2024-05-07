package backgrounds

import (
	"fmt"
	"sync"

	"github.com/closable/go-yandex-loyalty/internal/db"
	"github.com/closable/go-yandex-loyalty/internal/handlers"
	"go.uber.org/zap"
)

func SyncAccruals(db *db.Store, acc string, sugar *zap.SugaredLogger, orders ...string) {
	var wg sync.WaitGroup

	for _, order := range orders {
		wg.Add(1)
		go func() {

			defer wg.Done()
			res, status := handlers.AccrualActions(order, sugar, acc)
			if status < 204 {
				err := db.UpdateNotProcessedOrders(res.Order, res.Status, res.Accrual)
				if err != nil {
					sugar.Infoln(fmt.Sprintf("background sync order %s operation failed %s", order, err))
				}
				sugar.Infoln("background sync order complete", order)
			}
		}()
		wg.Wait()

	}

}
