> [!WARNING]
> WORK IN PROGRESS

## Описание
> [!NOTE]
> Проект написан в рамках курса "Продвинутый Go-разработчик" от [Яндекс.Практикум](https://practicum.yandex.ru).

**gometheus** - Сервис для сбора метрик на Go. Состоит из сервера и клиента (агента).

## Makefile
Для упрощения локальной разработки БД, сервер и агент запускаются в docker контейнерах.
Некоторые основные команды для работы с ними вынесены в Makefile:

| Команда           | Описание                                          |
|-------------------|---------------------------------------------------|
| build             | Собрать образы                                    |
| up                | Поднять контейнеры                                |
| down              | Убить контейнеры                                  |
| logs              | Вывести логи                                      |
| logsf             | Вывести и продолжать выводить логи                |
| vet-server        | Запустить go vet для сервера                      |
| test-server       | Запустить go test для сервера                     |
| test-race-server  | Запустить go test -race для сервера               |
| test-cover-server | Вычислить code coverage сервера                   |
| vet-agent         | Запустить go vet для агента                       |
| test-agent        | Запустить go test для агента                      |
| test-race-agent   | Запустить go test -race для агента                |
| test-cover-agent  | Вычислить code coverage агента                    |
| pprof-cpu-server  | Записать профиль использования CPU для сервера    |
| pprof-mem-server  | Записать профиль использования памяти для сервера |
| pprof-cpu-agent   | Записать профиль использования CPU для агента     |
| pprof-mem-agent   | Записать профиль использования памяти для агента  |

## Кофигурация
Сервер и агент поддерживают конфигурацию через переменные окружения и флаги процесса. В приоритете параметры переданные через флаг.

### Сервер
| Переменная окружения   | Флаг                     | Описание                                                                        | По-умолчанию         |
|------------------------|--------------------------|---------------------------------------------------------------------------------|----------------------|
| ADDRESS                | -a / --address           | Адрес сервера                                                                   | localhost:8080       |
| LOG_LEVEL              | -l / --log-level         | Уровень логирования                                                             | info                 |
| STORE_INTERVAL         | -i / --store-interval    | Интервал дампа текущего состояния в файл (в секундах)                           | 300                  |
| FILE_STORAGE_PATH      | -f / --file-storage-path | Файл для сохранения дампа                                                       | /tmp/metrics-db.json |
| RESTORE                | -r / --restore           | Восстановить из дампа при запуске                                               | true                 |
| DATABASE_DRIVER        | --database-driver        | Драйвер БД                                                                      | pgx                  |
| DATABASE_DSN           | -d / --database-dsn      | DSN базы данных                                                                 |                      |
| KEY                    | -k / --key               | Секретный ключ для HMAC подписи/валидации                                       |                      |
| CPU_PROFILE_FILE       | --cpu-profile-file       | Файл для записи профиля использования CPU                                       | ./cpu.pprof          |
| CPU_PROFILE_DURATION   | --cpu-profile-duration   | Время записи профиля использования CPU                                          | 30s                  |
| MEM_PROFILE_FILE       | --mem-profile-file       | Файл для записи профиля использования памяти                                    | ./mem.pprof          |

### Агент
| Переменная окружения | Флаг                   | Описание                                             | По-умолчанию   |
|----------------------|------------------------|------------------------------------------------------|----------------|
| ADDRESS              | -a / --address         | Адрес сервера                                        | localhost:8080 |
| LOG_LEVEL            | -l / --log-level       | Уровень логирования                                  | info           |
| POLL_INTERVAL        | -p / --poll-interval   | Интервал сбора метрик (в секундах)                   | 2              |
| REPORT_INTERVAL      | -r / --report-interval | Интервал отправки метрик (в секундах)                | 10             |
| BATCH_SIZE           | -b / --batch-size      | Кол-во отправляемых метрик в одном запросе           | 200            |
| KEY                  | -k / --key             | Секретный ключ для HMAC подписи/валидации            |                |
| RATE_LIMIT           | -m / --rate-limit      | Максимальное кол-во одновременных запросов к серверу | 10             |
| CPU_PROFILE_FILE     | --cpu-profile-file     | Файл для записи профиля использования CPU            | ./cpu.pprof    |
| CPU_PROFILE_DURATION | --cpu-profile-duration | Время записи профиля использования CPU               | 30s            |
| MEM_PROFILE_FILE     | --mem-profile-file     | Файл для записи профиля использования памяти         | ./mem.pprof    |

## Структура проекта

|      Директория | Субдиректория | Содержимое                                                                                    |
|----------------:|---------------|-----------------------------------------------------------------------------------------------|
|         .docker |               |                                                                                               |
|               - | agent         | Dockerfile для агента                                                                         |
|               - | server        | Dockerfile для сервера                                                                        |
|         .github |               | CI/CD конфиги                                                                                 |
|             cmd |               |                                                                                               |
|               - | agent         | Код входной точки агента                                                                      |
|               - | server        | Код входной точки сервера                                                                     |
|  internal/agent |               | Внутренние пакеты агента                                                                      |
|               - | app           | DI и запуск/остановка основных горутин агента                                                 |
|               - | collector     | Интерфейс коллектора метрик и его реализации                                                  |
|               - | config        | Обработка переменных окружения и флагов процесса                                              |
| internal/server |               | Внутренние пакеты сервера                                                                     |
|               - | app           | DI и запуск/остановка основных горутин агента                                                 |
|               - | api           | Хендлеры HTTP-запросов, работа с JSON                                                         |
|               - | config        | Обработка переменных окружения и флагов процесса                                              |
|               - | manager       | Фасад для работы с хранилищем                                                                 |
|               - | middleware    | HTTP-Middleware (HMAC, recover)                                                               | 
|               - | router        | Конфигурирование endpointов, прокидывание middleware                                          |
|               - | storage       | Интерфейс хранилища и его реализации (in-memory, pgsql, dump)                                 |
|               - | templates     | Шаблоны страниц и фасад для работы с ними                                                     |
| internal/common |               | Общие внутренние пакеты приложения                                                            | |
|               - | logger        | Логирование                                                                                   |
|               - | metric        | Интерфейс метрики, его реализации, фабрика и трансформеры                                     |
|               - | pprof         | Слушатель сигналов USR1/USR2 запускающий профилировщик                                        |
|             pkg |               | Доступные к переиспользованию пакеты                                                          |
|               - | client        | Go-клиент для HTTP-интерфейса сервера                                                         |
|               - | errors        | Функции для работы с ошибками                                                                 |
|               - | generator     | Реализация паттерна генератор                                                                 |
|               - | middleware    | HTTP-Middleware (комрессия, декомпрессия, интеграция с [zap](https://github.com/uber-go/zap)) |
|               - | mutex         | Реализация именованого мьютекса                                                               |
|               - | pprof         | Фасад для записи профилей pprof                                                               |
|               - | queue         | Реализация структуры очередь                                                                  |
|               - | request       | Модели запросов к сервису                                                                     |
|               - | response      | Модели ответов сервиса                                                                        |
|               - | retry         | Реализация retry логики                                                                       |
|               - | semaphore     | Реализация примитива синхронизации семафор                                                    |
|               - | slice         | Функции для работы со слайсами                                                                |

## Используемые сторонние пакеты

| Пакет                                                               | Описание                       |
|---------------------------------------------------------------------|--------------------------------|
| [caarlos0/env](https://github.com/caarlos0/env)                     | Обработка переменных окружения |
| [spf13/pflag](https://github.com/spf13/pflag)                       | Обработка флагов процесса      |
| [uber-go/zap](https://github.com/uber-go/zap)                       | Логирование                    |
| [go-chi/chi](https://github.com/go-chi/chi)                         | HTTP-роутинг                   |
| [asaskevich/govalidator](https://github.com/asaskevich/govalidator) | Валидация данных               |
| [go-resty/resty](https://github.com/go-resty/resty)                 | HTTP-клиент                    |
| [jackc/pgx](https://github.com/jackc/pgx)                           | Драйвер pgsql                  |
| [pressly/goose](https://github.com/pressly/goose)                   | Миграции БД                    |
| [shirou/gopsutil](https://github.com/shirou/gopsutil)               | Коллектор метрик CPU, RAM      |
| [stretchr/testify](https://github.com/stretchr/testify)             | Автотесты                      |
| [ory/dockertest](https://github.com/ory/dockertest)                 | Автотесты БД (если недоступна) |

