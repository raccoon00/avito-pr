# Сервис назначения Ревьюверов для Pull Request'ов

Микросервис для автоматического назначения ревьюверов на Pull Request'ы и управления командами разработки.

## Сборка и запуск

### Быстрый запуск с Makefile

```bash
# Сборка и запуск всех сервисов
# (Нужен установленный go и docker)
make run

# Просмотр логов
make logs

# Запуск интеграционных тестов
make test

# Очистка и перезапуск
make auto_test

# Остановка сервисов
make clean_down
```

### Ручная сборка и запуск

#### Сборка приложения
```bash
GOOS=linux CGO_ENABLED=0 go build -o bin/restapi ./cmd/restapi
```

#### Запуск с Docker Compose
```bash
# Используя стандартный .env файл
docker compose up --build -d

# Используя другой .env файл
docker compose --env-file .env.production up --build -d
```

#### Интеграционные тесты
```bash
# Запуск сервиса (требуется для тестов)
docker compose up -d

# Запуск тестов
go test -count=1 ./tests

# Запуск конкретного теста
go test -count=1 ./tests -run TestTeamAdd
```

## Описание структуры проекта

```
avito-pr/
├── cmd/restapi/           # Точка входа приложения
├── internal/
│   ├── adapter/           # Адаптеры для внешних систем
│   │   ├── http/          # HTTP хендлеры и роутинг (Gin)
│   │   └── postgres/      # Репозитории для PostgreSQL (pgx)
│   ├── app/               # Инициализация приложения, подвязывание адаптеров
│   ├── config/            # Конфигурация (через переменные среды)
│   ├── domain/            # Доменные модели и ошибки
│   └── service/           # Бизнес-логика и интерфейсы
├── migrations/            # Миграции базы данных
├── tests/                 # Интеграционные тесты
├── api/                   # API спецификация (OpenAPI) + тех задание
├── bin/                   # Скомпилированные бинарники
├── Makefile               # Скрипты сборки и запуска
├── docker-compose.yml     # Конфигурация Docker Compose
├── Dockerfile             # Конфигурация Docker для сервиса
└── go.mod                 # Зависимости Go
```

### Основные компоненты

- **HTTP API**: RESTful API реализованное с помощью `Gin`
- **PostgreSQL**: Запущена отдельным сервисом, взаимодействие с помощью `pgx`
- **Domain**: Структуры для бизнес-логики
- **Service**: Реализация бизнес-логика в отрыве от конкретных адаптеров
- **Интеграционные сервисы**: Полные тесты всех эндпоинтов

### Доступные эндпоинты

Были реализованы все обязательные эндпоинты:
- `POST /team/add` - Создание команды с участниками
- `GET /team/get` - Получение команды по имени
- `POST /users/setIsActive` - Установка активности пользователя
- `POST /pullRequest/create` - Создание PR с автоматическим назначением ревьюверов
- `POST /pullRequest/reassign` - Переназначение ревьювера
- `POST /pullRequest/merge` - Мердж PR (идемпотентная операция)
- `GET /users/getReview` - Получение PR, назначенных пользователю для ревью


## Дополнительные задания

### Нагрузочное тестирование

Для нагрузочного тестирования используется отдельный пакет `stress/stress.go`. Тестирование проводилось с параметрами:
- Длительность теста: 30 секунд
- Конкурентность: 10 воркеров
- Целевой RPS: 5 (требование задания)

#### Результаты нагрузочного тестирования

```
Stress Test Results:
Test                     Requests    Success   Errors      RPS  Avg Latency  P95 Latency  P99 Latency
--------------------------------------------------------------------------------------------
Create PR                    3887     100.0%        0    775.7          4ms          3ms          3ms
Reassign Reviewer             940     100.0%        0    186.8          3ms          3ms          4ms
Get Team                     7420     100.0%        0   1482.0          2ms          2ms          1ms
2025/11/16 22:50:35
Requirements Check
2025/11/16 22:50:35 OK: Create PR: Average latency 4ms meets 300ms requirement
2025/11/16 22:50:35 OK: Create PR: Success rate 100.00% meets 99.9% requirement
2025/11/16 22:50:35 OK: Reassign Reviewer: Average latency 3ms meets 300ms requirement
2025/11/16 22:50:35 OK: Reassign Reviewer: Success rate 100.00% meets 99.9% requirement
2025/11/16 22:50:35 OK: Get Team: Average latency 2ms meets 300ms requirement
2025/11/16 22:50:35 OK: Get Team: Success rate 100.00% meets 99.9% requirement
```

#### Запуск нагрузочного тестирования

```bash
# Сборка и запуск теста
go build -o bin/stress ./stress
./bin/stress

# Или автоматически
make auto_stress
```

Все эндпоинты сервиса успешно проходят нагрузочное тестирование и соответствуют требованиям SLI:
- Время ответа ≤300ms
- Успешность ≥99.9%
