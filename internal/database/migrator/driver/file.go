package driver

import (
	_ "github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitFile() {
	//TODO при импорте выше, при компиляции автоматически регистрируется нужный драйвер
	//TODO можно импорт перенести в файл выше по директории, но пока так оставлю
}
