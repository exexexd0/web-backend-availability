# Репозиторий Backend по курсу "Разработка интернет-приложений"

 Оценка отказоустойчивости системы. Услуги - компоненты системы (одиночные, с репликацией, с кластеризацией), заявка - расчет доступности системы (uptime) в зависимости от конфигурации выбранных компонентов.
---
* Ссылка на репозиторий бэкенда:

* Ссылка на репозиторий фронтенда:

* Ссылка на макет в Figma: https://www.figma.com/design/oL1m9AIE9W2jIa9YIOPLic/%D0%9C%D0%B0%D0%BA%D0%B5%D1%82?node-id=0-1&t=mWSe3q1G0KAJVAWq-1

## MinIO

Приложение читает медиа по постоянным публичным URL `http://localhost:9000/media/<файл>`. `docker-compose.yml` не меняется: после первого запуска контейнера один раз откройте bucket `media` для анонимного чтения, если доступ ещё не сохранён в volume:

```bash
docker exec -it minio_storage mc alias set myminio http://localhost:9000 root rootpassword
docker exec -it minio_storage mc mb myminio/media
docker exec -it minio_storage mc anonymous set public myminio/media
```
