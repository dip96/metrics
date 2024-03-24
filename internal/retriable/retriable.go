package retriable

import (
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"log"
	"net"
)

func IsRetriableError(err error) bool {
	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		log.Printf("Postgres error: %s", err.Error())
		return true
	}
	return false
}

func IsConnectError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		log.Printf("Connect error: %s", err.Error())
		return true
	}
	return false
}

func CheckError(err error) {
	IsRetriableError(err)
	IsConnectError(err)
}

func IsConnectionException(err error) bool {
	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}
