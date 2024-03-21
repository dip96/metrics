package driver

import _ "github.com/golang-migrate/migrate/v4/database/postgres"

func InitPostgres() {
	//TODO при импорте выше, при компиляции автоматически регистрируется нужный драйвер
	//TODO можно импорт перенести в файл выше по директории, но пока так оставлю
}
