# Download Service API
Сервис предоставляет REST API для управления задачами загрузки файлов по URL. Основные возможности:

## Реализованный функционал
### Создание задач:

- Автоматическое создание нового пользователя при первом запросе
- Создание задач для существующих пользователей
- Ограничение на максимальное количество задач для одного пользователя

### Управление задачами:

- Добавление URL для загрузки в существующую задачу
- Проверка статуса выполнения задачи
- Автоматическая упаковка файлов в ZIP-архив при завершении
- Ограничение по максимальному количеству ссылок в одной задачи

### Загрузка файлов:

- Валидация URL перед загрузкой
- Проверка типа и размера файлов
-  Ограничения на размер и тип файлов
- Скачивание готовых архивов

### Технические особенности

- Задачи и пользователи хранятся в памяти (in-memory)
- Загруженные файлы сохраняются на диск в указанную директорию
- Валидация входных данных (UUID, URL)
- Реализована Swagger документация

## Ограничения и нерешенные проблемы

- ❌ Не реализована автоматическая очистка старых файлов
- ❌ Нет ограничения на общий размер хранилища
- ❌ Файлы хранятся бессрочно, что может привести к переполнению диска
- ❌ Данные хранятся в памяти и теряются при перезапуске


## Как использовать
Запуск сервиса:

```
go run cmd/main.go
```
Доступные эндпоинты:

- POST /createTask - создать новую задачу <br />
Тело запроса (опционально):
```
{
  "user_id": "uuid-строка" // если не указать - создаст нового пользователя
}
```
Пример ответа:
```
{
  "user_id": "97557152-c723-44c8-bb92-241d37a81344",
  "task_id": "bcdfcfc5-f1ab-4118-a5a6-681f69df1698"
}
```
- POST /addTaskItems - добавить URL в задачу <br />
Тело запроса (обязательно):
```
{
  "user_id": "uuid-строка",
  "task_id": "uuid-строка",
  "links": [
    "https://example.com/file1.pdf",
    "https://example.com/image.jpg"
  ]
}
```
Пример ответа:
```
{
  "id": "bcdfcfc5-f1ab-4118-a5a6-681f69df1698",
  "status": "working",
  "links": [...],
  "errors": []
}
```

- GET /taskStatus - проверить статус задачи <br />
Тело запроса (обязательно):
```
{
  "user_id": "97557152-c723-44c8-bb92-241d37a81344",
  "task_id": "bcdfcfc5-f1ab-4118-a5a6-681f69df1698"
}
```
Возможные ответы: <br />
1. Задача в процессе:
```
{"status": "working"}
```
2. Задача в процессе:
```
{
  "status": "completed",
  "download_url": "http://your-server/download/archive.zip"
}
```
- GET /download/{file_id} - скачать архив <br />
Параметры: <br />
 file_id - имя архива (формат: uuid.zip) <br />
Ответ: <br />
Content-Type: application/zip
Content-Disposition: attachment; filename=archive.zip
### Документация API
Доступна по адресу:
```
http://localhost:8080/swagger
```
