package main

import (
	"fmt"
	"github.com/dip96/metrics/cmd/staticlint"
	"github.com/dorfire/go-analyzers/src/onlyany"
	gokartAnalyzers "github.com/praetorian-inc/gokart/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main является точкой входа для инструмента статического анализа.
// Эта функция выполняет следующие шаги:
//
// 1. Инициализирует пустой срез analyzers для хранения всех анализаторов.
//
// 2. Добавляет стандартные анализаторы из пакета golang.org/x/tools/go/analysis/passes:
//   - asmdecl.Analyzer: Проверяет соответствие ассемблерных деклараций.
//   - assign.Analyzer: Проверяет корректность операций присваивания.
//   - atomic.Analyzer: Проверяет правильное использование пакета sync/atomic.
//
// 3. Добавляет анализаторы из пакета honnef.co/go/tools:
//   - staticcheck.Analyzers (класс SA): Анализаторы для поиска ошибок, оптимизации
//     производительности и проверки корректности использования API.
//   - stylecheck.Analyzers (класс ST): Анализаторы для проверки стилистических
//     соглашений и согласованности кода.
//   - quickfix.Analyzers (класс QF1): Анализаторы для быстрого исправления
//     простых проблем.
//   - simple.Analyzers (класс S1): Анализаторы для упрощения кода и устранения
//     сложности.
//
// 4. Добавляет анализаторы из пакета github.com/praetorian-inc/gokart/analyzers:
//   - Набор анализаторов для выявления потенциальных уязвимостей безопасности.
//
// 5. Добавляет специализированные анализаторы:
//
//   - onlyany.Analyzer: Проверяет использование `any` в аргументах функций.
//
//   - staticlint.Analyzer: Проверяет, что os.Exit не вызывается напрямую в функции
//     main пакета main.
//
//     6. Вызывает multichecker.Main с набором всех анализаторов. multichecker
//     обеспечивает запуск анализаторов в правильном порядке (учитывая их зависимости)
//     и агрегирует их результаты.
//
// При добавлении каждого анализатора, функция выводит его имя в стандартный вывод.
// Это помогает пользователю понять, какие именно анализаторы будут выполнены.
//
// Использование:
//
//	go run ./cmd/multichecker ./...
//
// Эта команда запустит все указанные анализаторы для всех пакетов в текущем
// проекте. Результаты анализа (ошибки, предупреждения, советы) будут выведены
// в стандартный вывод ошибок.
//
// Преимущества этого подхода:
//   - Комплексный анализ: Объединение анализаторов из разных источников позволяет
//     охватить широкий спектр потенциальных проблем.
//   - Гибкость: Легко добавлять или удалять анализаторы, настраивая инструмент
//     под конкретные нужды проекта.
//   - Эффективность: multichecker оптимизирует выполнение анализаторов, избегая
//     повторного анализа одних и тех же частей кода.
//   - Интеграция: Этот инструмент может быть легко интегрирован в процессы
//     разработки, CI/CD пайплайны и редакторы кода.
func main() {
	var analyzers []*analysis.Analyzer

	// Стандартные статические анализаторы
	analyzers = append(analyzers, asmdecl.Analyzer)
	analyzers = append(analyzers, assign.Analyzer)
	analyzers = append(analyzers, atomic.Analyzer)

	// Все анализаторы класса SA из staticcheck.io
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
		fmt.Printf("Добавлен анализатор: %s\n", a.Analyzer.Name)
	}

	// Все анализаторы класса ST из staticcheck.io
	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
		fmt.Printf("Добавлен анализатор: %s\n", a.Analyzer.Name)
	}

	// Все анализаторы класса QF1 из staticcheck.io
	for _, a := range quickfix.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
		fmt.Printf("Добавлен анализатор: %s\n", a.Analyzer.Name)
	}

	// Все анализаторы класса S1 из staticcheck.io
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
		fmt.Printf("Добавлен анализатор: %s\n", a.Analyzer.Name)
	}

	for _, a := range gokartAnalyzers.Analyzers {
		analyzers = append(analyzers, a)
		fmt.Printf("Добавлен анализатор: %s\n", a.Name)
	}

	analyzers = append(analyzers, onlyany.Analyzer)
	fmt.Printf("Добавлен анализатор: %s\n", onlyany.Analyzer.Name)

	analyzers = append(analyzers, staticlint.Analyzer)
	fmt.Printf("Добавлен анализатор: %s\n", staticlint.Analyzer.Name)

	multichecker.Main(analyzers...)
}
