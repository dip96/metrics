package main

import (
	"fmt"
	"github.com/dip96/metrics/internal/storage"
	"github.com/dip96/metrics/internal/storage/mem"
	"testing"
)

func TestMain(m *testing.M) {
	// Инициализируем хранилище перед запуском тестов
	initStorage()

	// Запускаем тесты
	code := m.Run()

	// Выходим с кодом возврата
	//os.Exit(code)
	fmt.Print(code)
}

func initStorage() {
	//db, err := postgresStorage.NewDB()
	//if err != nil {
	//	panic(err)
	//}
	//defer db.Pool.Close()
	//
	//storage.Storage = db

	storage.Storage = mem.NewStorage()
}
