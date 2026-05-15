# log-parser-service

Микросервис на Go для приема архива с логами, парсинга топологии, сохранения результата в PostgreSQL и доступа к данным через REST API.

## Стек

- Go 1.22+
- `net/http`
- PostgreSQL
- Docker Compose

## Что Парсится

Сервис ожидает zip-архив внутри папки `data/`. В архиве должны быть файлы:

- `*.db_csv` - основной файл с секциями `START_NODES`, `START_PORTS`, `START_SWITCHES`, `START_SYSTEM_GENERAL_INFORMATION`.
- `*.sharp_an_info` - дополнительная информация по switch-узлам. Файл опциональный.

Пример структуры:

```text
data/
  log.zip
```

Внутри архива:

```text
ibdiagnet2.db_csv
ibdiagnet2.sharp_an_info
```

При ошибке парсинга файл отклоняется, а запись в `logs` получает статус `failed`.

## Запуск

```bash
docker compose up -d --build
```

Приложение будет доступно на порту `8080`.

Проверка:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```text
OK
```

Остановка:

```bash
docker compose down
```

Остановка с удалением данных PostgreSQL:

```bash
docker compose down -v
```

## Переменные Окружения

| Переменная | Описание | Значение в Docker Compose |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://log_parser:log_parser@db:5432/log_parser?sslmode=disable` |
| `PORT` | Порт HTTP-сервера | `8080` |
| `LOG_LEVEL` | Уровень логирования | `info` |

## Подготовка Архива

Если файлы лежат в `data/`, архив можно создать так:

```bash
cd data
zip log.zip ibdiagnet2.db_csv ibdiagnet2.sharp_an_info
cd ..
```

Файлы в `data/` не коммитятся в git, кроме `data/.gitkeep`.

## API

### POST `/api/v1/parse/`

Парсит архив с логом и возвращает `log_id`.

Запрос:

```bash
curl -X POST http://localhost:8080/api/v1/parse/ \
  -H "Content-Type: application/json" \
  -d '{"path":"data/log.zip"}'
```

Ответ:

```json
{
  "log_id": 1,
  "msg": "log parsed"
}
```

Для обратной совместимости также поддерживается прямой путь к двум файлам:

```json
{
  "csv_path": "data/ibdiagnet2.db_csv",
  "sharp_path": "data/ibdiagnet2.sharp_an_info"
}
```

Основной формат по заданию - `path` до zip-архива.

### GET `/api/v1/log/{log_id}`

Возвращает мета-информацию о загруженном логе.

```bash
curl http://localhost:8080/api/v1/log/1
```

Пример ответа:

```json
{
  "id": 1,
  "filename": "data/log.zip",
  "status": "success",
  "nodes_count": 5,
  "ports_count": 151,
  "created_at": "2026-05-15T04:00:12.508701Z"
}
```

### GET `/api/v1/topology/{log_id}`

Возвращает узлы и группы топологии.

```bash
curl http://localhost:8080/api/v1/topology/1
```

Пример ответа:

```json
{
  "log_id": 1,
  "nodes": [
    {
      "id": 1,
      "log_id": 1,
      "source_id": "host1",
      "name": "HOST_1",
      "type": "host"
    }
  ],
  "groups": {
    "hosts": [],
    "switches": []
  },
  "edges": []
}
```

### GET `/api/v1/node/{node_id}`

Возвращает узел и дополнительную информацию по нему.

```bash
curl http://localhost:8080/api/v1/node/1
```

### GET `/api/v1/port/{node_id}`

Возвращает порты узла.

```bash
curl http://localhost:8080/api/v1/port/1
```

## База Данных

Миграции применяются автоматически при старте приложения.

Минимальные таблицы:

- `logs` - загруженные логи, статус обработки, счетчики узлов и портов.
- `nodes` - узлы топологии.
- `ports` - порты узлов.
- `nodes_info` - дополнительная информация по узлам в формате key/value.

Связи:

- `nodes.log_id -> logs.id`
- `ports.log_id -> logs.id`
- `ports.node_id -> nodes.id`
- `nodes_info.node_id -> nodes.id`

## Топология

Сейчас сервис строит базовый граф:

- узлы типа `host`;
- узлы типа `switch`;
- порты, связанные с узлами;
- группы `hosts` и `switches`.

Поле `edges` уже присутствует в ответе `/api/v1/topology/{log_id}`, но надежные связи между портами пока не вычисляются, потому что в текущем наборе входных данных нет явного списка линков вида `local_node/local_port -> remote_node/remote_port`.

Какие связи можно строить дальше:

- связь `node -> port` уже есть через `ports.node_id`;
- связь `switch -> switch` или `host -> switch` можно пытаться выводить по `LID`, `PortGUID`, `LocalPortNum`, таблицам forwarding и состояниям портов;
- для надежного построения `edges` лучше использовать секцию лога, где явно указаны оба конца линка.

## Логирование

Сервис пишет структурные JSON-логи в stdout.

Пример request-лога:

```json
{"duration_ms":30,"event":"request","level":"info","method":"POST","path":"/api/v1/parse/","status":200}
```

Пример parse-лога:

```json
{"duration_ms":4,"event":"parse","level":"info","log_id":1,"nodes_count":5,"path":"data/log.zip","ports_count":151}
```

## Код-стайл И Lint

В проекте есть `Makefile` с командами для проверки кода.

Форматирование:

```bash
make fmt
```

Проверка форматирования без изменения файлов:

```bash
make fmt-check
```

Lint-проверка через стандартный Go analyzer:

```bash
make lint
```

Полная локальная проверка:

```bash
make check
```

## Postman

В репозитории есть коллекция:

```text
postman_collection.json
```

Импортируйте ее в Postman и запустите запросы по порядку:

1. `Health`
2. `Parse archive`
3. `Get log`
4. `Get topology`
5. `Get node`
6. `Get ports`

Коллекция использует переменные:

- `base_url` - по умолчанию `http://localhost:8080`
- `log_id`
- `node_id`
