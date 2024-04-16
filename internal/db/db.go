package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	errors_api "github.com/closable/go-yandex-loyalty/internal/errors"
	"github.com/closable/go-yandex-loyalty/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	DB *sql.DB
}

func NewDB(connstring string) (*Store, error) {
	db, err := sql.Open("pgx", connstring)
	if err != nil {
		return nil, err
	}
	return &Store{
		DB: db,
	}, nil
}

func (s *Store) GetConn() (*sql.Conn, error) {
	ctx := context.Background()
	conn, err := s.DB.Conn(ctx)

	return conn, err
}

func (s *Store) ValidateRegisterInfo(login, pass string) error {
	var errMessage string
	// invaid registerinformation
	if len(login) == 0 || len(pass) == 0 {
		errMessage = "part of register information is empty"
		return errors_api.NewAPIError(errors.New("login or pass empty"), errMessage, http.StatusBadRequest)
	}

	// user is present
	sql := "select count(*) cnt from ya.users where user_name=$1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	var rows int
	err = stmt.QueryRowContext(ctx, login).Scan(&rows)
	if err != nil {
		return errors_api.NewAPIError(err, "error during executing", http.StatusInternalServerError)
	}

	if rows > 0 {
		userPresentErr := errors.New("user is present")
		return errors_api.NewAPIError(userPresentErr, "login already present", http.StatusConflict)
	}

	return nil
}

func (s *Store) AddUser(login, pass string) error {
	sql := "insert into ya.users (user_name, user_passw, status) values ($1, sha256($2)::text, true)"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors_api.NewAPIError(err, "error during begin tx", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return errors_api.NewAPIError(err, "error during prepare execution", http.StatusInternalServerError)
	}

	_, err = stmt.ExecContext(ctx, login, pass)
	if err != nil {
		return errors_api.NewAPIError(err, "error during add pocess", http.StatusInternalServerError)
	}

	if err = tx.Commit(); err != nil {
		return errors_api.NewAPIError(err, "error commit during add user", http.StatusInternalServerError)
	}

	return nil
}

func (s *Store) Login(login, pass string) (int, error) {
	sql := `
	select user_id  
		from ya.users u 
	where u.user_name = $1 and u.user_passw = sha256($2)::text and status`

	// invaid registerinformation
	if len(login) == 0 || len(pass) == 0 {
		errMessage := "part of register information is empty"
		return 0, errors_api.NewAPIError(errors.New("login or pass empty"), errMessage, http.StatusBadRequest)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return 0, errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	var userID int
	err = stmt.QueryRowContext(ctx, login, pass).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return 0, errors_api.NewAPIError(err, "user id not found", http.StatusUnauthorized)

		} else {
			return 0, errors_api.NewAPIError(err, "error during executing", http.StatusInternalServerError)
		}

	}

	return userID, nil
}

func (s *Store) GetOrders(userID int) ([]models.OrdersDB, error) {
	sql := `
	select order_number, status, accrual, uploaded_at
		from ya.orders 
		where user_id=$1 order by uploaded_at desc`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res := make([]models.OrdersDB, 0)
	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return res, errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil || rows.Err() != nil {
		return res, errors_api.NewAPIError(err, "error during query", http.StatusInternalServerError)
	}

	for rows.Next() {
		item := models.OrdersDB{}
		err = rows.Scan(&item.OrderNumber, &item.Status, &item.Accrual, &item.UploadAt)
		if err != nil {
			return res, errors_api.NewAPIError(err, "error during scan", http.StatusInternalServerError)
		}
		res = append(res, item)

	}
	return res, nil
}

func (s *Store) Balance(userID int) (models.WithdrawDB, error) {
	sql := `
	select coalesce(sum(o.accrual), 0) accrual, coalesce(sum(w.sum),0) withdraw
		from ya.orders o
		left join ya.withdrawals w on w.order_number = o.order_number
	where o.user_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res := &models.WithdrawDB{}
	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return *res, errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	err = stmt.QueryRowContext(ctx, userID).Scan(&res.Current, &res.Withdrawn)
	if err != nil {
		return *res, errors_api.NewAPIError(err, "error during query", http.StatusInternalServerError)
	}
	fmt.Println("баланс ", res)
	return *res, nil
}

func (s *Store) AddOrder(userID int, orderNumber, accStatus string, accrual float64) error {
	sql := `
	select 
		case when o.user_id = $1 then true else false end is_owner,
		case when o.status = 'PROCESSING' then 202
			else 200
			end status	
		from ya.orders o 
		where order_number = $2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors_api.NewAPIError(err, "error during begin tx", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	var isOwner bool
	var status int

	err = stmt.QueryRowContext(ctx, userID, orderNumber).Scan(&isOwner, &status)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return errors_api.NewAPIError(err, "error during execute", http.StatusInternalServerError)
	}

	//fmt.Println(isOwner, status, err)

	if status != 0 {
		err := errors.New("the data is already there")
		if !isOwner {
			return errors_api.NewAPIError(err, "", http.StatusConflict)
		} else {
			return errors_api.NewAPIError(err, "", status) // only 202, 202
		}

	}

	sqlAdd := `
	insert into ya.orders 
		(user_id, order_number, status, accrual, uploaded_at)
	values 
		($1, $2, $3, $4, now())`

	stmt, err = tx.PrepareContext(ctx, sqlAdd)
	if err != nil {
		return errors_api.NewAPIError(err, "error during prepare insert", http.StatusInternalServerError)
	}

	_, err = stmt.ExecContext(ctx, userID, orderNumber, accStatus, accrual)
	if err != nil {
		return errors_api.NewAPIError(err, "error during executing insert order", http.StatusInternalServerError)
	}

	if err = tx.Commit(); err != nil {
		return errors_api.NewAPIError(err, "error commit during add order", http.StatusInternalServerError)
	}

	return nil
}

func (s *Store) AddWithdraw(userID int, orderNumber string, sum float64) error {
	// sql := `select count(*) from ya.withdrawals where order_number=$1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// tx, err := s.DB.BeginTx(ctx, nil)
	// if err != nil {
	// 	return errors_api.NewAPIError(err, "error during begin tx", http.StatusInternalServerError)
	// }
	// defer tx.Rollback()

	// stmt, err := tx.PrepareContext(ctx, sql)
	// if err != nil {
	// 	return errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	// }

	// var foundOrder int
	// err = stmt.QueryRowContext(ctx, orderNumber).Scan(&foundOrder)
	// if err != nil {
	// 	return errors_api.NewAPIError(err, "error during check order", http.StatusInternalServerError)
	// }
	// // order not found
	// if foundOrder > 0 {
	// 	err := errors.New("withdrawals alrrady has the order")
	// 	return errors_api.NewAPIError(err, "", http.StatusUnprocessableEntity)
	// }

	sql := `insert into ya.withdrawals (order_number, sum, processed_at) values ($1, $2, now())`
	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return errors_api.NewAPIError(err, "error during insert prepare", http.StatusInternalServerError)
	}
	// add withdraw
	res, err := stmt.ExecContext(ctx, orderNumber, sum)
	if err != nil {
		return errors_api.NewAPIError(err, "error during executing insert withdraw", http.StatusInternalServerError)
	}

	// if err = tx.Commit(); err != nil {
	// 	return errors_api.NewAPIError(err, "error commit during add withdraw", http.StatusInternalServerError)
	// }

	sql = `select sum from ya.withdrawals where order_number = $1`
	var o float64
	s.DB.QueryRow(sql, orderNumber).Scan(&o)

	fmt.Println("!!! добавление withdraw", res, err, orderNumber, sum, o)
	return nil
}

func (s *Store) GetWithdrawals(userID int) ([]models.WithdrawGetDB, error) {
	sql := `
	select w.order_number, w.sum, w.processed_at
		from ya.withdrawals w 
		left join ya.orders o on o.order_number = w.order_number
		where o.user_id=$1 `

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res := make([]models.WithdrawGetDB, 0)
	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return res, errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil || rows.Err() != nil {
		return res, errors_api.NewAPIError(err, "error during query", http.StatusInternalServerError)
	}

	for rows.Next() {
		item := models.WithdrawGetDB{}
		err = rows.Scan(&item.Order, &item.Sum, &item.ProcessedAt)
		if err != nil {
			return res, errors_api.NewAPIError(err, "error during scan", http.StatusInternalServerError)
		}
		res = append(res, item)

	}
	return res, nil
}

func (s *Store) PrepareDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	pipe := make([]string, 4)
	pipe[0] = `CREATE SCHEMA IF NOT EXISTS ya AUTHORIZATION postgres`
	pipe[1] = `CREATE TABLE IF NOT EXISTS ya.users
				(
					user_id bigserial NOT NULL,
					user_name character varying(255) COLLATE pg_catalog."default" NOT NULL,
					user_passw character varying(100) COLLATE pg_catalog."default" NOT NULL,
					status boolean DEFAULT false,
					CONSTRAINT user_pkey PRIMARY KEY (user_id)
				)`
	pipe[2] = `CREATE TABLE IF NOT EXISTS ya.orders
				(
					id_order bigserial NOT NULL,
					user_id integer NOT NULL,
					order_number character varying(20) COLLATE pg_catalog."default" NOT NULL,
					status character varying(20) COLLATE pg_catalog."default" NOT NULL,
					accrual numeric(10,2) DEFAULT 0.0,
					uploaded_at timestamp with time zone,
					CONSTRAINT orders_pkey PRIMARY KEY (id_order)
				)`
	pipe[3] = `CREATE TABLE IF NOT EXISTS ya.withdrawals
				(
					id_withdraw bigserial NOT NULL,
					order_number character varying(20) COLLATE pg_catalog."default" NOT NULL,
					sum numeric(10,2) DEFAULT 0.0,
					processed_at time with time zone,
					CONSTRAINT withdrawals_pkey PRIMARY KEY (id_withdraw)
				)`

	for ind, sql := range pipe {
		_, err := s.DB.ExecContext(ctx, sql)
		if err != nil {
			fmt.Println(ind, sql, err)
			return err
		}
	}

	return nil
}

func (s *Store) NotProcessedOrders() ([]string, error) {
	sql := `select order_number from ya.orders where status not in ('INVALID', 'PROCESSED')`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res := make([]string, 0)

	rows, err := s.DB.QueryContext(ctx, sql)
	if err != nil || rows.Err() != nil {
		return res, errors_api.NewAPIError(err, "error during query", http.StatusInternalServerError)
	}

	for rows.Next() {
		order := ""
		err = rows.Scan(&order)
		if err != nil {
			return res, errors_api.NewAPIError(err, "error during scan", http.StatusInternalServerError)
		}
		res = append(res, order)

	}
	return res, nil
}

func (s *Store) UpdateNotProcessedOrders(order, status string, accrual float64) error {
	sql := `update ya.orders SET status = $2, accrual = $3 where order_number = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return errors_api.NewAPIError(err, "error during begin tx", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return errors_api.NewAPIError(err, "error during prepare", http.StatusInternalServerError)
	}

	_, err = stmt.ExecContext(ctx, order, status, accrual)
	if err != nil {
		return errors_api.NewAPIError(err, "error during execute", http.StatusInternalServerError)
	}

	if err = tx.Commit(); err != nil {
		return errors_api.NewAPIError(err, "error commit during update order", http.StatusInternalServerError)
	}

	return nil
}
