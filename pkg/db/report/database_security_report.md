# Отчёт по безопасности и настройке СУБД

Проект: NeXuS
Команда: КЛАЦ-КЛАЦ
Дата: 17.05.2026

---

## 1. Работа с БД через сервисную учётную запись

### Принцип

Приложение **никогда** не работает с БД под административной учётной записью. Для каждого микросервиса создаётся отдельная сервисная учётная запись с минимально необходимыми правами (principle of least privilege).

### Скрипт создания сервисных пользователей

Скрипт находится в `deployments/nexus/scripts/init_service_users.sql:1`. Выбор расположения аргументирован:
- директория `deployments/nexus/` содержит всю инфраструктурную конфигурацию (Helm, ConfigMap, Secrets)
- директория `scripts/` — общепринятое место для SQL-скриптов инициализации
- скрипт монтируется в Kubernetes Job миграции как ConfigMap и выполняется при каждом `helm install/upgrade`

Скрипт (`deployments/nexus/scripts/init_service_users.sql:1`) параметризован переменными `psql`:

```sql
-- Параметры:
--   service_user      — имя сервисного пользователя
--   service_password  — пароль
--   admin_user        — владелец БД (для ALTER DEFAULT PRIVILEGES)
--   permissions       — набор прав (по умолчанию SELECT,INSERT,UPDATE,DELETE)

CREATE ROLE :service_user WITH LOGIN PASSWORD :service_password
GRANT CONNECT ON DATABASE ... TO :service_user
GRANT USAGE ON SCHEMA public TO :service_user
GRANT :permissions ON ALL TABLES IN SCHEMA public TO :service_user
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO :service_user
ALTER DEFAULT PRIVILEGES FOR ROLE :admin_user ... GRANT :permissions ON TABLES TO :service_user
ALTER DEFAULT PRIVILEGES FOR ROLE :admin_user ... GRANT USAGE, SELECT ON SEQUENCES TO :service_user
```

### Дифференциация прав по сервисам

В Kubernetes migration-job (`deployments/nexus/templates/migration-job.yaml:37`) для разных сервисов задаются разные права:

| Сервис | Права | Обоснование |
|---|---|---|
| `board_service` | `SELECT, INSERT, UPDATE, DELETE` | Полный CRUD над досками |
| `appeal_service` | `SELECT, INSERT, UPDATE, DELETE` | Полный CRUD над обращениями |
| `user_service` | `SELECT, INSERT, UPDATE` (без DELETE) | Профили пользователей не удаляются физически |

### Подключение к БД

DSN формируется из конфигурационных параметров (`pkg/postgres/config.go:10`), подгружаемых из переменных окружения через `viper`:

```yaml
# config.yaml (user/config.yaml:15)
database:
  user: ""              # DATABASE_SERVICE_USER
  password: ""          # DATABASE_SERVICE_PASSWORD
  host: ""
  port: ""
  name: ""
```

Креды хранятся в Kubernetes Secrets и не попадают в репозиторий.

---

## 2. Защита от SQL-инъекций

### Используемый механизм — параметризованные запросы (Prepared Statements)

Библиотека `jackc/pgx/v5` нативно поддерживает механизм Extended Query Protocol PostgreSQL. **Все** SQL-запросы в проекте построены через позиционные плейсхолдеры `$1, $2, ...` — аргументы передаются отдельно от текста SQL на уровне wire-протокола, что полностью исключает возможность SQL-инъекций.

Пример из `user/internal/user/repository/repository.go:37`:

```go
addUserQuery := `
    INSERT INTO "user" (link, display_name, password_hash, email)
    VALUES ($1, $2, $3, $4)
`
_, err := r.pool.Exec(ctx, addUserQuery,
    user.Link,        // uuid.UUID
    user.DisplayName, // string
    user.PasswordHash,// string
    user.Email,       // string
)
```

**Ни в одном файле проекта нет**:
- строковой интерполяции SQL через `fmt.Sprintf` с пользовательским вводом
- ручной конкатенации строк для построения запросов
- использования `db.Query()` с подстановкой значений в строку запроса

### Дополнительные уровни защиты

| Уровень | Механизм | Пример |
|---|---|---|
| БД (DDL) | CHECK-ограничения на длину полей | `CONSTRAINT check_display_name CHECK (char_length(display_name) <= 128)` — `pkg/db/migrations/000001_init_schema.up.sql` |
| Приложение | Валидация длины и формата полей | `max_len_name_user: 128`, `max_len_password: 128`, `min_len_password: 8` — `user/config.yaml` |
| Приложение | Типизация идентификаторов через `uuid.UUID` | Все link-поля — строгий тип `uuid.UUID`, исключающий произвольные строки |
| Приложение | Валидация MIME-типов загружаемых файлов | `valid_extensions: image/png, image/jpg, image/jpeg, image/webp` |

### Интерфейс DBEngine

Все репозитории реализуют единый интерфейс для работы с БД через `pgxpool`:

```go
// user/internal/user/repository/repository.go:19
type DBEngine interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

Это гарантирует, что любой новый код будет использовать параметризованные запросы без возможности передать сырой SQL.

### Почему именно `pgx`?

- Нативная поддержка PostgreSQL (не через `database/sql`, что даёт лучшую производительность)
- Полная поддержка Extended Query Protocol (Prepared Statements на уровне драйвера)
- Встроенный connection pool (`pgxpool`) с настраиваемыми параметрами
- Поддержка `pgconn.PgError` для типобезопасной обработки ошибок PostgreSQL (`pgerrcode.UniqueViolation`, `pgerrcode.NotNullViolation`)
- Нативная поддержка сканирования в Go-типы (включая `uuid.UUID`)
- Поддержка tracer'ов для Prometheus-метрик

---

## 3. Пул соединений и параметры соединений

### Реализация Connection Pool

Пул соединений реализован через `pgxpool` (библиотека `jackc/pgx/v5/pgxpool`). Код инициализации расположен в двух местах:
- `pkg/postgres/postgres.go:24` — основная версия с передачей `*Config`
- `pkg/db/postgres.go:40` — версия с передачей `Config` по значению

```go
func NewPoolPostgres(dsn string, conf *Config, logger *zerolog.Logger) (*pgxpool.Pool, error) {
    poolConfig, err := pgxpool.ParseConfig(dsn)
    poolConfig.ConnConfig.Tracer = &tracer.PrometheusTracer{}
    poolConfig.MinConns = conf.MinConnections
    poolConfig.MaxConns = conf.MaxConnections
    poolConfig.MaxConnLifetime = conf.MaxConnectionLifetime
    poolConfig.HealthCheckPeriod = conf.MaxHealthCheckPeriod

    for i := 1; i <= conf.MaxRetries; i++ {
        contextWithTimeOut, cancel := context.WithTimeout(context.Background(), conf.TimeOut)
        if pool, err := pgxpool.NewWithConfig(contextWithTimeOut, poolConfig); err == nil {
            pingErr := pool.Ping(contextWithTimeOut)
            if pingErr == nil {
                return pool, nil
            }
            pool.Close()
        }
        time.Sleep(conf.PingSleepTime)
        cancel()
    }
    return nil, ErrorConnectPosgress
}
```

Ключевые особенности:
- **Retry-логика**: до 5 попыток с паузой 2 секунды между попытками
- **Context timeout**: каждая попытка подключения имеет таймаут 5 секунд
- **Health check**: каждые 30 секунд pgxpool проверяет здоровье соединений
- **Max connection lifetime**: 1 час — после этого соединение принудительно пересоздаётся для предотвращения утечек и stale-соединений

### Балансировка `max_connections` (PostgreSQL) и `MaxConns` (пул)

| Параметр | Значение | Обоснование |
|---|---|---|
| `database.max_connections` (пул приложения) | **10** | Рассчитано на 3 микросервиса × 10 = 30 соединений суммарно с запасом под пиковую нагрузку |
| `max_connections` (PostgreSQL) | **100** (стандартное значение) | При 3 сервисах и 10 соединениях каждый — макс. 30 одновременных. Запас более 3× гарантирует стабильность |
| `database.min_connections` (пул) | **2** | Держим 2 "горячих" соединения для быстрой обработки первых запросов |
| `database.max_connection_lifetime` | **1h** | Предотвращает использование stale-соединений после перезапуска БД или обрыва сети |
| `database.max_health_check_period` | **30s** | Частая проверка здоровья; при обнаружении мёртвого соединения pgxpool заменяет его новым |

**Аргументация**: `MaxConns = 10` выбрано с учётом того, что каждое соединение PostgreSQL — это отдельный процесс на сервере, потребляющий ~5-10 MB RAM. При трёх микросервисах и 10 соединениях получаем ≤ 30 соединений = ≤ 300 MB, что безопасно для типового инстанса с 1-2 GB RAM. Увеличение `MaxConns` сверх 30 на сервис не имеет смысла — возросший параллелизм упрётся в CPU БД, а не в количество доступных соединений.

### `listen_addresses`

Настроено в `deployments/nexus/configs/postgresql.conf:1`:

```ini
listen_addresses = '*'
```

**Аргументация**: в кластере Kubernetes все микросервисы и БД находятся во внутренней сети кластера. Значение `'*'` разрешает подключения с любого сетевого интерфейса внутри пода, но сам порт PostgreSQL (5432) **не exposed** наружу — он доступен только через внутренний Service типа ClusterIP. Таким образом, безопасность обеспечивается на уровне Kubernetes NetworkPolicy, а не на уровне `listen_addresses`. Использование `localhost` было бы избыточным и сломало бы междеплойментное взаимодействие.

---

## 4. Настройка параметров сервера и клиента

### Таймауты на стороне приложения

В проекте таймауты реализованы **на уровне приложения**, а не на уровне PostgreSQL. Это сделано осознанно: контекст с таймаутом пробрасывается от входящего gRPC-запроса через все слои до вызова БД.

| Параметр | Значение | Файл |
|---|---|---|
| `database.time_out` (таймаут соединения) | **5s** | `pkg/postgres/config.go:35` |
| `database.ping_sleep_time` (пауза между retry) | **2s** | `pkg/postgres/config.go:34` |
| `database.max_retries` | **5** | `pkg/postgres/config.go:36` |
| gRPC-таймауты сервисов | user: 3s, board: 5s, appeal: 5s | `facade/config.yaml` |

**Аргументация значений**:
- `time_out = 5s`: установка TCP-соединения с PostgreSQL в локальной сети K8s обычно занимает < 100ms. Таймаут 5s даёт запас на временную перегрузку сети, но не позволяет запросу зависнуть навсегда.
- Ограничения уровня `statement_timeout` (например, 1 минута) **не имеют смысла для UX** — пользователь не будет ждать минуту. 5-секундный таймаут на уровне gRPC + пагинация + кэширование (Redis) обеспечивают быстрый отклик.
- С точки зрения **DOS-защиты**: комбинация gRPC-таймаутов + rate-limiter'а (свой микросервис) + лимитов пула соединений (MaxConns=10) не позволяет одному клиенту исчерпать ресурсы БД.

### Таймауты на стороне PostgreSQL

На уровне БД `statement_timeout` и `lock_timeout` **не устанавливаются**, так как управление таймаутами полностью делегировано приложению через `context.Context`. Это даёт более гибкий контроль: разные запросы могут иметь разные таймауты (логин — быстрее, загрузка файла в S3 — дольше), что невозможно при едином `statement_timeout`.

`idle_in_transaction_session_timeout` не установлен — при использовании `pgxpool` транзакции управляются явно через `pool.Begin(ctx)`, и зависших транзакций не возникает.

---

## 5. Логгирование и протоколирование медленных запросов

### Конфигурация логгирования PostgreSQL

Файл: `deployments/nexus/configs/postgresql.conf:2`

```ini
log_destination = 'stderr'
logging_collector = on
log_directory = 'pg_log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_rotation_age = 1440          # 24 часа
log_rotation_size = 102400       # 100 MB
log_truncate_on_rotation = on

log_min_duration_statement = 200 # медленный запрос > 200ms
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on
log_temp_files = 0               # логировать все временные файлы
log_error_verbosity = default

log_line_prefix = '%m [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

**Формат логов совместим с PGBadger**: `stderr` + `log_line_prefix` с временной меткой `%m`, PID процесса `%p`, номером строки `%l`, пользователем `%u`, базой `%d`, приложением `%a`, клиентом `%h` — это стандартный формат, распознаваемый PGBadger'ом "из коробки" без дополнительных настроек.

При запуске PGBadger:
```bash
pgbadger /var/lib/postgresql/data/pg_log/postgresql-*.log -o report.html
```

### `log_min_duration_statement = 200ms`

Значение 200ms выбрано исходя из бизнес-требований:
- типовой gRPC-запрос к микросервису должен отрабатывать за 50-150ms (с учётом сетевых задержек)
- запросы > 200ms уже считаются медленными и требуют анализа (отсутствие индекса, неоптимальный план выполнения)
- при 200ms в лог попадают действительно проблемные запросы, а не весь трафик

Ротация логов: файлы логов ротируются каждые 24 часа (1440 минут) или при достижении 100 MB, что даёт разумный размер файлов.

### `pg_stat_statements` и `auto_explain`

Файл: `deployments/nexus/configs/postgresql.conf:23`

```ini
shared_preload_libraries = 'pg_stat_statements,auto_explain'
pg_stat_statements.track = all
pg_stat_statements.track_utility = on
pg_stat_statements.max = 10000

auto_explain.log_min_duration = 500
auto_explain.log_analyze = on
auto_explain.log_buffers = on
auto_explain.log_timing = on
auto_explain.log_triggers = on
auto_explain.log_verbose = on
auto_explain.log_nested_statements = on
auto_explain.log_level = 'LOG'
```

Расширения также включаются в миграциях каждого сервиса:
```sql
-- user/internal/db/migrations/000002_enable_extensions.up.sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();
```

**`pg_stat_statements`**:
- `track = all` — отслеживает все запросы (включая вложенные)
- `max = 10000` — хранит статистику по 10 000 уникальным нормализованным запросам
- Даёт возможность находить самые частые и самые медленные запросы через `SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC`

**`auto_explain`**:
- `log_min_duration = 500ms` — запросы дольше 500ms автоматически получают план выполнения в лог
- `log_analyze = on` — план включает фактическое время выполнения (не только оценки)
- `log_buffers = on` — показывает использование буферов
- `log_nested_statements = on` — планы для запросов внутри функций

---

## 6. Мониторинг нагрузки

### Архитектура мониторинга

Используется стек Prometheus + Grafana, развёрнутый через Docker Compose и Kubernetes:

```
Docker Compose                          Kubernetes
┌──────────────┐                       ┌────────────────┐
│  Prometheus   │◄──── scrape ────────►│  k3s services   │
│  :9090        │                      │  (ClusterIP)    │
└──────┬───────┘                       └────────────────┘
       │
┌──────┴───────┐
│   Grafana    │
│   :3000      │
└──────────────┘
```

### Метрики на стороне приложения

Каждый микросервис экспортирует метрики на порту `:9091` через `promhttp.Handler()`. Prometheus собирает их по HTTP:

`deployments/monitoring/prometheus.yml:5`:
```yaml
scrape_configs:
  - job_name: 'user_service'
    static_configs:
      - targets: ['nexus-user:9091']
  - job_name: 'board_service'
    static_configs:
      - targets: ['nexus-board:9091']
  - job_name: 'appeal_service'
    static_configs:
      - targets: ['nexus-appeal:9091']
  # + facade, authorization, mail_sender, rate_limiter
```

### Prometheus-метрики БД

Файл: `pkg/metrics/postgres/metrics.go:9`

```go
var (
    DbQueriesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{Name: "db_queries_total"},
        []string{"status"},  // "success" | "error"
    )
    DbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
        },
        []string{"method"},
    )
)
```

Метрики собираются через pgx Tracer — `pkg/db/tracer/tracer.go:15`:

```go
type PrometheusTracer struct{}

func (pt *PrometheusTracer) TraceQueryStart(ctx context.Context, ...) context.Context {
    return context.WithValue(ctx, startTimeKey, time.Now())
}

func (pt *PrometheusTracer) TraceQueryEnd(ctx context.Context, ...) {
    duration := time.Since(startTime).Seconds()
    dbMetric.DbQueryDuration.WithLabelValues("sql_query").Observe(duration)
    dbMetric.DbQueriesTotal.WithLabelValues(status).Inc()
}
```

Это позволяет видеть в Grafana:
- количество запросов к БД (суммарно и по статусам)
- распределение длительности запросов (гистограмма с бакетами от 1ms до 1s)
- находить аномалии (всплески ошибок, запросы > 1s)

### Дополнительные метрики

| Тип метрик | Расположение | Что отслеживается |
|---|---|---|
| gRPC-метрики | `pkg/metrics/grpc/metrics.go` | `grpc_requests_total`, `grpc_request_duration_seconds` |
| HTTP-метрики | `pkg/metrics/http/metrics.go` | `http_requests_total`, `http_request_duration_seconds` |
| gRPC-интерсептор | `pkg/interceptors/prometheus.go:1` | Автоматическая запись всех входящих gRPC-запросов |

### Мониторинг ресурсов и RPS

- **CPU / RAM БД и микросервисов**: Prometheus + Grafana + node-exporter (в K3s — через `infrastructure/monitoring/minimal/prometheus.yaml`)
- **RPS и сетевая статистика**: gRPC-интерсепторы записывают method, status, duration для каждого запроса
- **Sentry**: отслеживание ошибок в рантайме, настраивается через `pkg/logger/sentry.go:1` с конфигурацией DSN, environment, release, sample rate

---

## 7. S3-хранилище (kS3)

### Реализация

S3-клиент реализован в пакете `pkg/s3/` через AWS SDK v2. Используется для хранения:
- **аватаров пользователей** (user service): бакет `S3_AVATARS_BUCKET`
- **фонов досок** (board service): бакет `S3_BOARDS_BACKGROUNDS_BUCKET`
- **вложений карточек** (board service): бакет `S3_CARDS_ATTACHMENTS_BUCKET`
- **вложений обращений** (appeal service): бакет appeal attachments

### Ключевые файлы

| Файл | Назначение |
|---|---|
| `pkg/s3/client.go` | Интерфейсы `S3Client` и `S3Bucket` |
| `pkg/s3/aws_client.go` | `NewAWSClient()` — создание клиента с region, endpoint, accessKey, secretKey |
| `pkg/s3/aws_bucket.go` | `AWSBucket.Put()` и `.Delete()` — загрузка/удаление объектов |
| `pkg/s3/aws_client_test.go` | Unit-тесты клиента |
| `pkg/s3/aws_bucket_test.go` | Unit-тесты бакета с моком AWSClientAPI |
| `pkg/s3/mock_aws_client/AWSClientAPI.go` | Mockery-мок для тестов |

### Конфигурация (на примере board)

```yaml
s3:
  region: ""                              # S3_REGION
  endpoint: ""                            # S3_ENDPOINT
  access_key: ""                          # S3_ACCESS_KEY
  secret_key: ""                          # S3_SECRET_KEY
  boards_backgrounds_bucket: ""           # S3_BOARDS_BACKGROUNDS_BUCKET
  boards_backgrounds_prefix: ""           # S3_BOARDS_BACKGROUNDS_PREFIX
  cards_attachments_bucket: ""            # S3_CARDS_ATTACHMENTS_BUCKET
  cards_attachments_prefix: ""            # S3_CARDS_ATTACHMENTS_PREFIX
  connect_timeout: ""                     # S3_CONNECT_TIMEOUT
```

Файлы доступны публично через `ObjectCannedACLPublicRead`, URL формируется как `https://{bucket}.{endpoint}`.

---

## 8. Итоговая сводка безопасности

| Аспект | Статус | Реализация |
|---|---|---|
| Сервисная учётная запись | ✅ | `init_service_users.sql` + K8s Job, права дифференцированы по сервисам |
| SQL Injection | ✅ | Параметризованные запросы через `pgx` (`$1, $2`), без интерполяции |
| Connection Pool | ✅ | `pgxpool`, MinConns=2, MaxConns=10, health check 30s, lifetime 1h |
| Таймауты | ✅ | Контекстные (5s на connect), gRPC-таймауты, retry до 5 раз |
| Логгирование медленных запросов | ✅ | `log_min_duration_statement=200ms`, формат совместим с PGBadger |
| `pg_stat_statements` | ✅ | `track=all`, `max=10000` |
| `auto_explain` | ✅ | `log_min_duration=500ms`, `log_analyze=on` |
| Мониторинг | ✅ | Prometheus + Grafana, метрики pgx Tracer, gRPC/HTTP интерсепторы |
| S3-хранилище | ✅ | AWS SDK v2, unit-тесты с моками |
| `listen_addresses` | ✅ | `'*'` внутри кластера K3s, порт не exposed наружу |
| Sentry | ✅ | Отслеживание ошибок рантайма |
| Unit-тесты БД | ✅ | `pkg/postgres/postgres_test.go`, моки S3 |
