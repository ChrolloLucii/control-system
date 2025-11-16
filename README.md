# Control System - Микросервисная архитектура

> Система управления пользователями и заказами на Go с микросервисной архитектурой

## Архитектура

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│          Gateway :8080              │
│  ┌────────────────────────────┐    │
│  │ Middleware:                │    │
│  │ - JWT Auth                 │    │
│  │ - CORS                     │    │
│  │ - Rate Limiter             │    │
│  │ - Request ID               │    │
│  └────────────────────────────┘    │
└────────┬──────────────┬─────────────┘
         │              │
    ┌────▼────┐    ┌───▼─────┐
    │  User   │    │  Order  │
    │ Service │    │ Service │
    │  :3001  │    │  :3002  │
    └─────────┘    └─────────┘
```

## Компоненты

### 1. **Gateway** (порт 8080)
- Единая точка входа для всех запросов
- JWT аутентификация
- Rate limiting (100 RPS, burst 200)
- CORS
- Request ID для трассировки
- Reverse proxy к микросервисам

### 2. **User Service** (порт 3001)
- Регистрация и аутентификация пользователей
- Управление профилем
- Список пользователей (admin)
- JWT токены
- Валидация данных

### 3. **Order Service** (порт 3002)
- Создание заказов
- Получение заказов с пагинацией и сортировкой
- Обновление статуса заказа
- Отмена заказов
- Доменные события (OrderCreated, OrderStatusUpdated, OrderCancelled)
- Проверка существования пользователя

## API Документация

### Swagger UI (Локально - рекомендуется!)

```bash
cd docs
.\start-swagger.ps1
# Откройте http://localhost:3000/index.html в браузере
```

**Интерактивная документация** с возможностью:
- Просмотра всех endpoints
- Тестирования API прямо из браузера
- Авторизации через JWT
- Примеров запросов и ответов

### OpenAPI спецификация

Полная документация API: [`docs/openapi.yaml`](docs/openapi.yaml)

Онлайн просмотр: https://editor.swagger.io/ (скопируйте содержимое yaml)

### Postman Collection

Импортируйте в Postman: [`docs/postman-collection.json`](docs/postman-collection.json)

### Основные endpoints

#### Публичные (без авторизации)
```
POST /api/v1/users/register  - Регистрация
POST /api/v1/users/login     - Вход
GET  /health                 - Health check
```

#### Защищённые (требуется JWT)
```
# Users
GET  /api/v1/users/profile      - Получить профиль
PUT  /api/v1/users/profile      - Обновить профиль
GET  /api/v1/users              - Список пользователей (admin)

# Orders
POST   /api/v1/orders           - Создать заказ
GET    /api/v1/orders           - Список заказов
GET    /api/v1/orders/{id}      - Получить заказ
PUT    /api/v1/orders/{id}/status - Обновить статус
DELETE /api/v1/orders/{id}      - Отменить заказ
```
