# Geocore - Ядро геоповещения на Go (Нерв)

Основной бэкенд-сервис для системы гео-уведомлений «Нерв».
Написан на Go, PostgreSQL и Redis.

## Возможности
- CRUD Инцидентов (Опасных зон)
- Проверка местоположения (координаты против опасных зон)
- Асинхронные Webhook-уведомления через очередь Redis
- Мониторинг здоровья системы

## Требования
- Docker & Docker Compose
- Go 1.25+ 

## Как запустить (Docker Compose)
Самый простой способ запустить всю систему (Сервис + БД + Redis + Мок-сервер) — использовать Docker Compose.

### 1. Запуск всего
```bash
docker-compose up --build
```
Эта команда запускает все 4 сервиса:
- **Geocore Service**: `http://localhost:8080`
- **Mock Webhook Server**: `http://localhost:9090`
- **PostgreSQL**: `localhost:5433` (внутренний порт 5432)
- **Redis**: `localhost:6380` (внутренний порт 6379)

### 2. Применение миграций
Сервис не накатывает миграции автоматически при старте (выбор Clean Architecture). Вам нужно применить схему вручную:
```bash
# Найти имя контейнера postgres
docker ps
# Запустить миграцию (поправьте имя контейнера если нужно)
cat migrations/000001_initial_schema.up.sql | docker exec -i geocore-postgres-1 psql -U user -d geocore
```

Еще для локального запуска миграций можно использовать утилиту `golang-migrate` через Makefile:
```bash
cd ./scripts/migrations && make migrate-up
```

### 3. Проверка
Проверить здоровье: 
```
curl http://localhost:8080/api/v1/system/health
```


## API Эндпоинты

### Incidents (Инциденты) - Требуется API Key
Для доступа к методам управления инцидентами необходимо передавать заголовок `X-API-Key`.
API Key (для теста): `secret-key-123`

- `GET /api/v1/incidents` - Список инцидентов (params: limit, offset)
  ```bash
  curl "http://localhost:8080/api/v1/incidents?limit=10&offset=0" \
  -H "X-API-Key: secret-key-123"
  ```
- `POST /api/v1/incidents` - Создать инцидент
  ```bash
  curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-key-123" \
  -d '{
    "title": "Gas Leak",
    "description": "Beware of smell",
    "latitude": 55.7558,
    "longitude": 37.6173,
    "radius_meters": 500
  }'
  ```
- `GET /api/v1/incidents/:id` - Получить инцидент
  ```bash
  # Замените 1 на реальный ID инцидента
  curl http://localhost:8080/api/v1/incidents/1 \
  -H "X-API-Key: secret-key-123"
  ```
- `PUT /api/v1/incidents/:id` - Обновить инцидент
  ```bash
  # Замените 1 на реальный ID инцидента
  curl -X PUT http://localhost:8080/api/v1/incidents/1 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-key-123" \
  -d '{
    "title": "Gas Leak Updated",
    "description": "Smell gone",
    "latitude": 55.7558,
    "longitude": 37.6173,
    "radius_meters": 0
  }'
  ```
- `DELETE /api/v1/incidents/:id` - Удалить инцидент
  ```bash
  # Замените 1 на реальный ID инцидента
  curl -X DELETE http://localhost:8080/api/v1/incidents/1 \
  -H "X-API-Key: secret-key-123"
  ```
- `GET /api/v1/incidents/stats` - Получить статистику пользователей по зоне
  ```bash
  curl http://localhost:8080/api/v1/incidents/stats \
  -H "X-API-Key: secret-key-123"
  ```

### Location Check (Проверка местоположения)
- `POST /api/v1/location/check`
  ```bash
  curl -X POST http://localhost:8080/api/v1/location/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-001",
    "latitude": 55.7559,
    "longitude": 37.6174
  }'
  ```
  Возвращает совпадающие зоны. Если найдено совпадение, асинхронно отправляет вебхук.

### Просмотр Webhook-уведомлений (Mock Server)
Когда пользователь попадает в опасную зону, Geocore отправляет webhook на Mock Server.
Вы можете увидеть полученные уведомления двумя способами:

1. **В браузере или через curl**:
   Зайдите на [http://localhost:9090](http://localhost:9090) или выполните:
   ```bash
   curl http://localhost:9090/
   ```
   Это вернет JSON-список всех полученных событий (хранится в памяти, сбрасывается при перезапуске).

2. **В логах Docker**:
   ```bash
   docker logs -f geocore-mock-1
   ```

   ```bash
   docker logs -f geocore-mock-1
   ```

### Тестирование через Ngrok (публичный URL)
Для проверки доставки уведомлений через реальный интернет:

1. **Добавьте токен Ngrok**:
   Зарегистрируйтесь на [ngrok.com](https://ngrok.com), получите автотокен и добавьте его в `.env`:
   ```env
   NGROK_AUTHTOKEN="ваш-токен"
   ```

2. **Запустите проект**:
   ```bash
   docker-compose up -d
   ```

3. **Получите публичный URL**:
   Откройте веб-интерфейс Ngrok: [http://localhost:4040](http://localhost:4040). Скопируйте HTTPS URL (вида `https://xyz.ngrok-free.app`).

4. **Обновите конфигурацию**:
   Вставьте полученный URL в `.env`:
   ```env
   WEBHOOK_URL="https://xyz.ngrok-free.app"
   ```

5. **Примените изменения**:
   Перезапустите приложение, чтобы оно подхватило новый адрес:
   ```bash
   docker-compose restart app worker
   ```

Теперь уведомления будут уходить на публичный адрес ngrok и перенаправляться на ваш локальный мок-сервер.

## Конфигурация
Настройки приложения загружаются из переменных окружения.

1. **Основные переменные** берутся из файла `.env`.
   Пример конфигурации находится в файле `.env.example`.
   Ключевые переменные:
   - `HTTP_PORT`
   - `DATABASE_URL` (или компоненты подключения `POSTGRES_*`)
   - `REDIS_ADDR` (или компоненты `REDIS_*`)
   - `MOCK_SERVER_URL`
   - `API_KEY`
   - `STATS_TIME_WINDOW_MINUTES`

2. **Docker Compose**:
   При запуске через `docker-compose.yml`, переменные из `.env` передаются в контейнеры.
   Дополнительно, `docker-compose.yml` может переопределять некоторые переменные (например, хосты сервисов `postgres`, `redis`) для корректной работы внутри сети Docker.

## Архитектура
- **Handler**: HTTP Transport (Gin)
- **Service**: Бизнес-логика (Incident, Geo)
- **Repository**: Доступ к данным (Postgres, Redis)
- **Worker**: Обработчик фоновых задач (Webhooks)
