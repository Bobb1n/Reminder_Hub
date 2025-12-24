# Reminder Hub - Полная документация

## Общая информация

**Название проекта:** Reminder Hub
**Описание:** Автоматизированное приложение для создания задач из email-писем с использованием LLM. Анализирует письма, извлекает встречи и дедлайны.

## Архитектура

Reminder Hub - это микросервисное приложение на Go, состоящее из нескольких сервисов, связанных через RabbitMQ. Приложение использует PostgreSQL для хранения данных и NGINX как reverse proxy.

### Структура проекта

```
/workspace/
├── README.md
├── docker-compose.yml
├── go.mod
├── go.sum
├── .env.example
├── nginx/
│   ├── conf.d/
│   ├── logs/
│   └── nginx.conf
├── pkg/                     # Общие пакеты
│   ├── http/
│   ├── logger/
│   ├── models/
│   └── rabbitmq/
└── services/               # Микросервисы
    ├── api-gateway/        # Шлюз API
    ├── auth/              # Сервис аутентификации
    ├── core/              # Сервис работы с почтой
    ├── collector/         # Сервис сбора данных
    └── analyzer/          # Сервис анализа
```

## Компоненты системы

### 1. API Gateway (api-gateway)
- **Порт:** 8080
- **Назначение:** Единая точка входа для всех внешних запросов
- **Функции:**
  - Аутентификация запросов
  - Роутинг к соответствующим внутренним сервисам
  - Проверка JWT токенов
  - Обработка CORS и других общих задач

### 2. Auth Service (auth-service)
- **Порт:** 8081
- **Назначение:** Управление аутентификацией и пользователями
- **Технологии:** PostgreSQL, JWT
- **Функции:**
  - Регистрация и вход пользователей
  - Генерация и проверка JWT токенов
  - Управление черным списком токенов
  - Получение информации о текущем пользователе

#### API Endpoints Auth Service:
- `GET /health` - проверка статуса сервиса
- `POST /auth/register` - регистрация нового пользователя
- `POST /auth/login` - вход пользователя
- `GET /auth/me` - информация о текущем пользователе
- `POST /auth/validate` - проверка токена (внутренний endpoint)
- `POST /auth/logout` - выход из системы

### 3. Core Service (core-service)
- **Порт:** 8082
- **Назначение:** Работа с почтой по imap и отправка сообщений в Analyzer Service
- **Технологии:** PostgreSQL, RabbitMQ
- **Функции:**
   - Подключение к почтовым ящикам
    - Получение новых писем
    - Отправка писем в очередь для анализа
    - Управление настройками почтовых ящиков

### 4. Collector Service (collector-service)
- **Порт:** 8083
- **Назначение:** Сбор задач и дедлайнов 
- **Технологии:** PostgreSQL, RabbitMQ
- **Функции:**
    - Crud операции с событиями
    - Получение сообщений по rabbit

### 5. Analyzer Service (analyzer-service)
- **Порт:** 8083
- **Назначение:** Анализ писем с помощью LLM
- **Технологии:** RabbitMQ, Mistral API
- **Функции:**
  - Анализ текста писем
  - Извлечение встреч и дедлайнов
  - Создание задач в Core Service
  - Обработка естественного языка

## Инфраструктура

### База данных: PostgreSQL
- **Версия:** 15
- **Назначение:** Хранение всех данных приложения
- **Схемы:**
  - Пользователи
  - Задачи
  - Профили
  - Настройки почты
  - Черный список токенов

### Очередь сообщений: RabbitMQ
- **Версия:** 3-management
- **Назначение:** Асинхронная коммуникация между сервисами
- **Использование:** Передача email-сообщений для анализа, уведомлений и других асинхронных задач

### Веб-сервер: Nginx
- **Роль:** Обратный прокси-сервер
- **Функции:** Балансировка нагрузки, SSL-терминация, статическая раздача

## Зависимости

### Основные зависимости Go:
- `go.uber.org/fx` - фреймворк для внедрения зависимостей
- `go.uber.org/zap` - высокопроизводительный логгер
- `github.com/labstack/echo/v4` - веб-фреймворк
- `github.com/streadway/amqp` - клиент RabbitMQ
- `github.com/google/uuid` - генерация UUID

## Конфигурация

### Переменные окружения (.env)

#### PostgreSQL:
- `POSTGRES_USER` - имя пользователя БД (по умолчанию: reminder)
- `POSTGRES_PASSWORD` - пароль пользователя БД (по умолчанию: reminder)
- `POSTGRES_DB` - имя базы данных (по умолчанию: reminderhub)

#### RabbitMQ:
- `RABBITMQ_DEFAULT_USER` - имя пользователя RabbitMQ (по умолчанию: guest)
- `RABBITMQ_DEFAULT_PASS` - пароль пользователя RabbitMQ (по умолчанию: guest)

#### API Gateway:
- `AUTH_SERVICE_URL` - URL сервиса аутентификации
- `CORE_SERVICE_URL` - URL ядревого сервиса
- `COLLECTOR_SERVICE_URL` - URL сервиса сбора
- `INTERNAL_API_TOKEN` - токен для внутренней аутентификации
- `JWT_SECRET` - секретный ключ для JWT
- `SERVER_PORT` - порт запуска (по умолчанию: 8080)

#### Auth Service:
- `DB_URL` - строка подключения к БД
- `JWT_SECRET` - секретный ключ для JWT
- `SERVER_PORT` - порт запуска (по умолчанию: 8081)

#### Core Service:
- `CORE_DB_URL` - строка подключения к БД
- `RABBIT_URL` - строка подключения к RabbitMQ
- `SERVER_PORT` - порт запуска (по умолчанию: 8082)

#### Analyzer Service:
- `RABBIT_URL` - строка подключения к RabbitMQ
- `OPENAI_API_KEY` - ключ API OpenAI
- `SERVER_PORT` - порт запуска (по умолчанию: 8083)

#### Collector Service:
- `DB_URL` - строка подключения к БД
- `RABBIT_URL` - строка подключения к RabbitMQ
- `SERVER_PORT` - порт запуска (по умолчанию: 8084)

## Запуск приложения

### Предварительные требования:
1. Docker и Docker Compose
2. Mistral API Key

### Установка:
1. Клонируйте репозиторий
2. Настройте переменные окружения:

### Запуск:
```bash
docker-compose up --build
```

Приложение будет доступно по адресу `http://localhost`

## Разработка

### Добавление нового сервиса:
1. Создайте директорию в `/services/{service-name}`
2. Добавьте Dockerfile
3. Обновите `docker-compose.yml`
4. Добавьте зависимости в `go.mod`
5. Обновите документацию

### Структура сервиса:
Каждый сервис должен содержать:
- `cmd/main.go` - точка входа
- `internal/` - внутренние пакеты сервиса
- `go.mod` - модуль Go для сервиса
- `Dockerfile` - инструкции для сборки контейнера

## Безопасность

### Аутентификация:
- JWT токены с подписью HMAC
- Refresh токены с возможностью отзыва
- Валидация токенов через Auth Service

### Защита:
- Валидация входных данных
- Защита от SQL-инъекций через подготовленные выражения
- CORS политики в API Gateway

## Мониторинг и логирование

### Логирование:
- Централизованное логирование с использованием Zap
- Формат JSON для удобства парсинга
- Уровни логирования: debug, info, warn, error

### Health checks:
- Все сервисы имеют endpoint `/health`
- Docker Compose проверяет статус сервисов

## Масштабирование

### Горизонтальное масштабирование:
- Микросервисная архитектура позволяет масштабировать каждый сервис независимо
- RabbitMQ обеспечивает балансировку нагрузки на обработчики сообщений
- PostgreSQL может быть заменен на кластер для масштабируемости

## Тестирование

### Типы тестов:
- Unit тесты для бизнес-логики
- Integration тесты для взаимодействия с БД и очередями
- E2E тесты для проверки сквозных сценариев

## Деплоймент

### CI/CD:
- Контейнеризация с Docker
- Зависимости сервисов в docker-compose.yml




## Технические детали

### Используемые технологии:
- **Язык программирования:** Go (Golang)
- **Веб-фреймворк:** Echo
- **DI Framework:** Uber FX
- **База данных:** PostgreSQL
- **Очереди сообщений:** RabbitMQ
- **Контейнеризация:** Docker
- **Оркестрация:** Docker Compose
- **Reverse Proxy:** Nginx
- **LLM:** OpenAI API
- **Аутентификация:** JWT

### Паттерны проектирования:
- **Dependency Injection** через Uber FX
- **Clean Architecture** в сервисах
- **Event-driven** архитектура через RabbitMQ
- **API Gateway** паттерн для маршрутизации

## Документация полного API маршрута

Эта секция покрывает полный пользовательский путь. Начиная от регистрации в сервисе `auth` и заканчивая получением данных из сервиса `collector`

### 1. Регистрация пользователя

`POST /auth/register`

**Пример запроса:**
```json
{
    "password":  "password123",
    "email":  "testuser_1@example.com"
}
```

**Пример ответа:**
```json
{
    "message":"User created successfully",
    "user_id":"<USER_ID>"
}
```



### 2. Логин Пользователя

`POST /auth/login`

**Пример запроса:**
```json
{
    "password":  "password123",
    "email":  "testuser_1@example.com"
}
```

**Пример ответа:**
```json
{
    "access_token":"<ACCESS_TOKEN>",
    "expires_in":900,
    "refresh_token":"<REFRESH_TOKEN>",
    "token_type":"Bearer"
}
```

### 3. Создание интеграции с Email

`POST /api/v1/integrations/email`

**Пример запроса:**
```json
{
    "imap_host":  "imap.example.com",
    "email_address":  "testuser_1@example.com",
    "imap_port":  993,
    "use_ssl":  true,
    "password":  "password123"
}

```

**Пример ответа:**
```json
{   
    "id":"<ID>",
    "status":"created"
}
```



### 4. Симуляция отправки Email в RabbitMQ:

**Exchange:** 
`Raw_emails`

```json
{"emails":
[
    {
        "email_id":"<EMAIL_ID>",    
        "subject":"Action Required: Finalize Q4 Report Slides","body_text":"Hi team, Just a reminder that the presentation for the Q4 financial review is coming up. I\u0027ve put together the initial draft, but I need someone to take ownership of the final slides. Specifically, can you please create a new summary slide that visualizes the YoY growth for our main product lines? Also, please double-check all the figures against the master spreadsheet in the shared drive. This needs to be completed by Friday, December 26th, EOD. Let me know if you have any questions. Thanks!","from_address":"manager@example.com","sync_timestamp":"2025-12-24T19:33:52.4931351Z","message_id":"test-1806753482",
        "date_received":"2025-12-24T19:33:52.4850641Z","user_id":"7fbdb3d5-eced-4736-a6e5-eddc9a8dbf79"}
]
}
```



### 5. Get Processed Tasks

`GET /api/v1/reminders/`

**Example Response:**

```json
[
    {  
        "id":"<ID>",
        "user_id":"<USER_ID>",
        "email_id":"<EMAIL_ID>",
        "title":"<TITLE>", 
        "Description":"<DESCRIPTION>",
        "deadline":"2025-12-26T23:00:00Z",
        "status":"pending",
        "priority":"urgent",
        "created_at":"2025-12-24T19:33:53.810241Z","updated_at":"2025-12-24T19:33:53.810241Z"
    }
]

```

