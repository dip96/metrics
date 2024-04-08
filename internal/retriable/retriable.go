package retriable

import (
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
)

func IsConnectionException(err error) bool {
	var pgErr pgx.PgError
	if errors.As(err, &pgErr) {
		// Проверяем код ошибки на тайм-аут
		if pgErr.Code == pgerrcode.IdleSessionTimeout ||
			pgErr.Code == pgerrcode.IdleInTransactionSessionTimeout {
			return true
		}
	}

	return false
}
