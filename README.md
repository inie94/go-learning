# go-learning 🚀

> Персональный репозиторий для изучения Go с нуля до уровня Senior Developer.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

---

## 📖 О репозитории

Этот репозиторий — моя structured learning path по языку Go. Здесь я фиксирую:

- 📝 Конспекты по каждой теме
- 💻 Практические примеры и упражнения
- 🧪 Результаты экспериментов с рантаймом и профилированием
- 🏗️ Мини-проекты для закрепления материала

Подробный план обучения со всеми темами, разбитыми по этапам, находится в файле [`PLAN.md`](./PLAN.md).

---

## 🛠️ Структура репозитория
```
go-learning/
├── README.md ← вы здесь
├── PLAN.md ← полный текстовый план обучения
├── stage-1-basics/
│ ├── 01-syntax/
│ ├── 02-composite-types/
│ ├── 03-functions/
│ ├── 04-interfaces/
│ └── 05-error-handling/
├── stage-2-middle/
│ ├── 01-concurrency/
│ ├── 02-tooling/
│ ├── 03-io/
│ └── 04-testing/
├── stage-3-internals/
│ ├── 01-runtime/
│ ├── 02-profiling/
│ ├── 03-architecture/
│ └── 04-generics-reflection/
├── stage-4-senior/
│ ├── 01-networking/
│ ├── 02-databases/
│ ├── 03-security/
│ └── 04-cicd-observability/
└── projects/ ← мини-проекты для закрепления
```

---

## 🚀 Быстрый старт

```bash
# Клонируй репозиторий
git clone https://github.com/your-username/go-learning.git
cd go-learning

# Запусти примеры из любого этапа
cd stage-1-basics/01-syntax
go run main.go

# Запусти тесты с проверкой на гонки данных
go test -race ./...
```