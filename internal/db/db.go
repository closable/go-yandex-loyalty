package db

import (
	"context"
	"database/sql"
	"fmt"
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
	// invaid registerinformation
	if len(login) == 0 || len(pass) == 0 {
		return errors_api.ErrorRegInfo
	}

	// user is present
	sql := "select count(*) cnt from ya.users where user_name=$1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	var rows int
	err = stmt.QueryRowContext(ctx, login).Scan(&rows)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	if rows > 0 {
		return errors_api.ErrorConflict
	}

	return nil
}

func (s *Store) AddUser(login, pass string) error {
	sql := "insert into ya.users (user_name, user_passw, status) values ($1, sha256($2)::text, true)"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorBeginTx.Error(), err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	_, err = stmt.ExecContext(ctx, login, pass)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecCommit.Error(), err)
	}

	return nil
}

func (s *Store) Login(login, pass string) (int, error) {
	sqlString := `
	select user_id  
		from ya.users u 
	where u.user_name = $1 and u.user_passw = sha256($2)::text and status`

	// invaid registerinformation
	if len(login) == 0 || len(pass) == 0 {
		return 0, errors_api.ErrorRegInfo
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stmt, err := s.DB.PrepareContext(ctx, sqlString)
	if err != nil {
		return 0, fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	var userID int
	err = stmt.QueryRowContext(ctx, login, pass).Scan(&userID)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
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
		return res, fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil || rows.Err() != nil {
		return res, fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	for rows.Next() {
		item := models.OrdersDB{}
		err = rows.Scan(&item.OrderNumber, &item.Status, &item.Accrual, &item.UploadAt)
		if err != nil {
			return res, fmt.Errorf("%s %w", errors_api.ErrorScanQuery.Error(), err)
		}
		res = append(res, item)

	}
	return res, nil
}

func (s *Store) Balance(userID int) (float32, float32, error) {
	sql := `
	select coalesce(sum(o.accrual),0) current, coalesce((select sum(w.sum) from ya.withdrawals w where user_id=$1),0) withdrawn
		from ya.orders o 
		where user_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//res := &models.WithdrawDB{}
	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return 0, 0, fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	var current float32
	var withdrawn float32

	err = stmt.QueryRowContext(ctx, userID, userID).Scan(&current, &withdrawn)
	if err != nil {
		return 0, 0, fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}
	return current, withdrawn, nil
}

func (s *Store) AddOrder(userID int, orderNumber, accStatus string, accrual float32) error {
	sqlString := `
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
		return fmt.Errorf("%s %w", errors_api.ErrorBeginTx.Error(), err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sqlString)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	var isOwner bool
	var status int

	err = stmt.QueryRowContext(ctx, userID, orderNumber).Scan(&isOwner, &status)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	if status != 0 {
		if !isOwner {
			return errors_api.ErrorConflict
		} else {
			return errors_api.ErrorInfoFound
		}
	}

	sqlAdd := `
	insert into ya.orders 
		(user_id, order_number, status, accrual, uploaded_at)
	values 
		($1, $2, $3, $4, now())`

	stmt, err = tx.PrepareContext(ctx, sqlAdd)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	_, err = stmt.ExecContext(ctx, userID, orderNumber, accStatus, accrual)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecCommit.Error(), err)
	}
	return nil
}

func (s *Store) AddWithdraw(userID int, orderNumber string, sum float32) error {
	sql := `insert into ya.withdrawals (user_id, order_number, sum, processed_at) values ($1, $2, $3, now())`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorBeginTx.Error(), err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}
	// add withdraw
	_, err = stmt.ExecContext(ctx, userID, orderNumber, sum)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecCommit.Error(), err)
	}

	return nil
}

func (s *Store) GetWithdrawals(userID int) ([]models.WithdrawGetDB, error) {
	sql := `select w.order_number, w.sum, w.processed_at from ya.withdrawals w where w.user_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	res := make([]models.WithdrawGetDB, 0)

	stmt, err := s.DB.PrepareContext(ctx, sql)
	if err != nil {
		return res, fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil || rows.Err() != nil {
		return res, fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	for rows.Next() {
		item := models.WithdrawGetDB{}
		err = rows.Scan(&item.Order, &item.Sum, &item.ProcessedAt)
		if err != nil {
			return res, fmt.Errorf("%s %w", errors_api.ErrorScanQuery.Error(), err)
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
					user_id integer NOT NULL,
					order_number character varying(20) COLLATE pg_catalog."default" NOT NULL,
					sum numeric(10,2) DEFAULT 0.0,
					processed_at timestamp with time zone,
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
		return res, fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	for rows.Next() {
		order := ""
		err = rows.Scan(&order)
		if err != nil {
			return res, fmt.Errorf("%s %w", errors_api.ErrorScanQuery.Error(), err)
		}
		res = append(res, order)

	}
	return res, nil
}

func (s *Store) UpdateNotProcessedOrders(order, status string, accrual float32) error {
	sql := `update ya.orders SET status = $2, accrual = $3 where order_number = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorBeginTx.Error(), err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorPrepareQuery.Error(), err)
	}

	_, err = stmt.ExecContext(ctx, order, status, accrual)
	if err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecQuery.Error(), err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s %w", errors_api.ErrorExecCommit.Error(), err)
	}

	return nil
}
