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

## Веб-сервис (лабораторная 3)

Базовый URL: `http://localhost:8081`. Все методы ниже начинаются с `/api` и отвечают JSON.

Успех: `{"status":"success","data":...}`. Ошибка: `{"status":"fail","message":"..."}`. Коды: 200, 201, 400, 404, 500.

Текущий пользователь во всех методах домена услуги берётся из singleton `CurrentUser()` в `internal/app/repository/repository.go`. Идентификатор зафиксирован константой `currentUserID = DefaultCreatorID` (`1`, пользователь `user` из сида). Пока нет авторизации, клиент его не передаёт.

Системные поля `id`, `status`, `creator_id`, `moderator_id`, `created_at`, `formed_at`, `completed_at`, а также `image_url` и `video_url` из запроса не принимаются. Их выставляет сервер. Лишнее поле в JSON или форме даёт 400.

Статусы меняются только так: `POST` создаёт `draft`, `PUT` переводит свой черновик в `published`. Вернуть в черновик нельзя. `DELETE` — мягкое удаление своей услуги (`deleted`). Записи `deleted` методы GET не отдают. У пользователя не больше одного черновика.

Файлы `image` и `video` передаются как multipart, не как URL. В MinIO (bucket `media`) сохраняется объект с латинским именем `images/<hex>.jpg` или `videos/<hex>.mp4`, это имя пишется в `image_url` / `video_url`. Страницы HTML по-прежнему сохраняют полный публичный URL.

### Домен услуги

`GET /api/components?uptime_percent=`

Список только опубликованных услуг. `uptime_percent` необязателен: минимальная расчётная доступность от 0 до 100. В каждой записи сериализатор списка и признак `is_owner`: 1, если создатель совпадает с текущим пользователем, иначе 0. Пустой список — `data: []`.

`GET /api/components/feed` и `GET /api/components/feed/:id`

Опубликованная услуга: без id — первая, с id — эта запись. С `?next=true` — следующая опубликованная, после последней снова первая. В ответе `is_liked`: 1, если текущий пользователь лайкнул услугу, иначе 0. Черновик и удалённая услуга без `next` дают 404.

`GET /api/components/draft`

Черновик текущего пользователя, не больше одной записи, id в пути нет. Если черновика нет — 404.

`POST /api/components`

Создание черновика, ответ 201. Тело `multipart/form-data`: `name`, `description`, `config_type` (`Single Node`, `Clustering`, `Replication`), файлы `image` и `video` необязательны. Если черновик уже есть — 400.

`PUT /api/components/:id`

Публикация своего черновика. JSON: `uptime_percent`, `system_impact` (числа от 0 до 100). Сервер ставит `status=published` и `formed_at`. Чужая или уже не черновик: 404, если записи нет среди доступных, и 400, если своя услуга не в статусе черновика.

`DELETE /api/components/:id`

Мягкое удаление только своей услуги. Ответ: успех и id. Чужая, удалённая или несуществующая — 404.

`POST /api/components/:id/like`

Лайк текущего пользователя на опубликованную услугу. JSON: `{"value":1}` ставит лайк, `{"value":0}` снимает. Повтор идемпотентен. Черновик и удалённая услуга — 404.

### Домен пользователя

`POST /api/users`

Регистрация, ответ 201. JSON: `login`, `password`. `is_moderator` с клиента не принимается и всегда `false`. Пароль в ответ не входит. Занятый логин — 400.

`POST /api/users/login`

Заглушка аутентификации для лабораторной 4. Всегда 200, сессия не создаётся.

`POST /api/users/logout`

Заглушка деавторизации для лабораторной 4. Всегда 200, сессия не сбрасывается.

### Сериализаторы

Модели `User` и `Component` в `internal/app/ds` несут теги `json`. Пароль и статус помечены `json:"-"` и в ответы не попадают.

`ComponentListSerializer` — список: поля услуги без статуса, дат и создателя, плюс `likes_count` и `is_owner`.

`ComponentDetailSerializer` — черновик, создание и публикация: поля услуги без статуса и без вложенного пользователя, плюс `likes_count`.

`ComponentFeedSerializer` — лента: те же поля, что у подробной услуги, плюс `is_liked` (0 или 1).

### Таблицы

`users`

| Поле | Тип | Ограничения |
|---|---|---|
| id | BIGINT | PK |
| login | VARCHAR(25) | UNIQUE, NOT NULL |
| password | VARCHAR(100) | NOT NULL |
| is_moderator | BOOLEAN | NOT NULL, по умолчанию false |

`components`

| Поле | Тип | Ограничения |
|---|---|---|
| id | BIGINT | PK |
| name | VARCHAR(100) | NOT NULL |
| description | VARCHAR(1000) | NOT NULL |
| status | VARCHAR(20) | NOT NULL: draft, published, deleted |
| image_url | VARCHAR(500) | NOT NULL, для API — латинское имя объекта MinIO |
| video_url | VARCHAR(500) | NOT NULL, для API — латинское имя объекта MinIO |
| config_type | VARCHAR(30) | NOT NULL |
| uptime_percent | REAL | NULL до публикации |
| system_impact | REAL | NULL до публикации |
| created_at | TIMESTAMPTZ | NOT NULL, ставит сервер при создании |
| formed_at | TIMESTAMPTZ | NULL до публикации |
| creator_id | BIGINT | NOT NULL, FK на users.id |

У одного пользователя не больше одной строки со `status = draft` (индекс `idx_components_one_draft_per_user`).

`component_likes`

| Поле | Тип | Ограничения |
|---|---|---|
| id | BIGINT | PK |
| user_id | BIGINT | NOT NULL, FK на users.id, уникальная пара с component_id |
| component_id | BIGINT | NOT NULL, FK на components.id, уникальная пара с user_id |

Внешние ключи: ON UPDATE RESTRICT, ON DELETE RESTRICT.

### Порядок десяти запросов в Insomnia или Postman

Коллекция на `http://localhost:8081`. Текущий пользователь — id 1, своего черновика в сиде у него нет.

1. `GET /api/components?uptime_percent=99.9`
2. `POST /api/components` — форма: name, description, config_type, файлы image и video
3. `GET /api/components/draft` — поля созданного черновика
4. `PUT /api/components/<id черновика>` — JSON `{"uptime_percent":99.5,"system_impact":12}`
5. `GET /api/components/feed`
6. `GET /api/components/feed/<id>?next=true`
7. `POST /api/components/<id опубликованной>/like` — JSON `{"value":1}`
8. `DELETE /api/components/<id своей услуги>`
9. `POST /api/users` — JSON `{"login":"newuser","password":"newuser"}`
10. `POST /api/users/login`

Скриншоты 11–13 — три `SELECT` в Adminer после этих запросов:

```sql
SELECT id, name, status, image_url, video_url, uptime_percent, system_impact, created_at, formed_at, creator_id FROM components ORDER BY id;
SELECT id, user_id, component_id FROM component_likes ORDER BY id;
SELECT id, login, is_moderator FROM users ORDER BY id;
```

Скриншоты 14–15 — модели и сериализаторы в `internal/app/ds/component.go` и `internal/app/ds/user.go`. Скриншоты 16–17 — `CurrentUser()` и её вызовы в методах домена услуги, затем этот README.
