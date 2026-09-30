# Репозиторий Backend по курсу "Разработка интернет-приложений"

 Оценка отказоустойчивости системы. Услуги - компоненты системы (одиночные, с репликацией, с кластеризацией), заявка - расчет доступности системы (uptime) в зависимости от конфигурации выбранных компонентов.
---
* Ссылка на репозиторий бэкенда:

* Ссылка на репозиторий фронтенда:

* Ссылка на макет в Figma: https://www.figma.com/design/oL1m9AIE9W2jIa9YIOPLic/%D0%9C%D0%B0%D0%BA%D0%B5%D1%82?node-id=0-1&t=mWSe3q1G0KAJVAWq-1

## Запуск

```bash
docker compose up -d
go run ./cmd/migrate
# в Adminer http://localhost:8083 заполнить таблицы SQL из scripts/seed.sql
# система: PostgreSQL, сервер: postgres, БД: mydb, пользователь/пароль из .env
go run ./cmd/app
```

Приложение: http://localhost:8081/components  
Adminer: http://localhost:8083

## MinIO

Приложение читает медиа по постоянным публичным URL `http://localhost:9000/media/<файл>`. `docker-compose.yml` не меняется для MinIO: после первого запуска контейнера один раз откройте bucket `media` для анонимного чтения, если доступ ещё не сохранён в volume:

```bash
docker exec -it minio_storage mc alias set myminio http://localhost:9000 root rootpassword
docker exec -it minio_storage mc mb myminio/media
docker exec -it minio_storage mc anonymous set public myminio/media
```

Фото и видео, выбранные на шаге «Далее» страницы добавления, загружаются в bucket `media` (папки `images/` и `videos/`), а в `components.image_url` / `components.video_url` сохраняется их публичный URL. Параметры подключения берутся из окружения (значения по умолчанию совпадают с `docker-compose.yml`):

| Переменная | По умолчанию |
|---|---|
| `MINIO_ENDPOINT` | `localhost:9000` |
| `MINIO_ACCESS_KEY` | `root` |
| `MINIO_SECRET_KEY` | `rootpassword` |
| `MINIO_BUCKET` | `media` |
| `MINIO_USE_SSL` | `false` |
| `MINIO_PUBLIC_URL` | `http://<MINIO_ENDPOINT>` |

Фото и видео необязательны. Если файл не выбран (или URL недоступен), на шаге подтверждения, в ленте и на странице компонентов показывается статика, а не MinIO: `/static/img/default.jpg` и `/static/img/default.mp4`.

## Добавление компонента

1. Шаг А: изображение, видео, название, описание и тип конфигурации → «Далее» создаёт черновик.
2. Шаг Б: те же данные только для просмотра (вместо полей загрузки — превью фото и видео), редактируются только «Расчётная доступность» и «Влияние на систему» → «Опубликовать».

Краткая информация в ленте не хранится в БД: до нажатия «ЕЩЁ» показываются первые 30 символов описания.

## Сиды и порядок показа лабораторной 2

В `scripts/seed.sql` уже есть: 3 пользователя; услуги в статусах published/draft/deleted; строки M:N в `component_likes`. Черновик принадлежит user2 (`creator_id = 3`), чтобы у пользователя 1 можно было нажать «Далее».

1. Скриншоты 1–2: Adminer — логически удалить услугу через `status`, затем `SELECT`.
2. Скриншоты 3–10: три страницы приложения — поиск, удалить услугу, открыть URL удалённой, добавить услугу, `SELECT`, опубликовать, снова `SELECT`.
3. Скриншоты 11–13: в БД изменить `config_type` / `uptime_percent` и число строк в `component_likes`, показать в приложении.
4. Скриншоты 14–21: модели, 5 ORM-контроллеров, удаление через SQL `UPDATE`.
5. Скриншот 22: фото и видео по умолчанию в HTML (`/static/img/default.*`).
