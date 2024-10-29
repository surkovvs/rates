# RATES

## Запуск
```
git clone https://github.com/surkovvs/rates && \
make docker-build
```

## Для проверки работоспособности сервиса
```
make run_bombardier
```

## Конфигурация
Конфигурирование выполнено в следующем приоритете:
- значения по умолчанию;
- .env файл (путь может быть задан флагом --dotenvpath='path to file' или -c='path to file');
- переменные окружения;
- флаги.

Применяются следующие конфигурационные параметры:
- DB_NAME - наименование БД
- DB_USER - имя пользователя БД
- DB_PASSWORD - пароль БД
- DB_HOST - хост БД
- DB_PORT - порт доступа к БД
- DB_MIGR_PATH - путь к файлам миграции БД (используется https://github.com/golang-migrate)
- GRPC_HOST - хост gRPC сервера
- GRPC_PORT - порт gRPC сервера
- HTTP_HOST - хост HTTP сервера
- HTTP_PORT - порт HTTP сервера
- LOG_LVL - уровень логирования (https://github.com/uber-go/zap/blob/0ba452dbe15478739ad4bab1067706018a3062c6/level.go#L30-L49)
- MARKET - id рынка для https://garantexio.github.io/#market-data
- METRICS - включение/выключение метрик ('true'/'false')