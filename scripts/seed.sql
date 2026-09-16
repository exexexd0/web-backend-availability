-- Наполнение БД через Adminer (система PostgreSQL, сервер: postgres, БД: mydb).
-- Сначала запустите миграции: go run ./cmd/migrate

INSERT INTO users (id, login, password, is_moderator) VALUES
  (1, 'user', 'user', false),
  (2, 'admin', 'admin', true),
  (3, 'user2', 'user2', false);

INSERT INTO components (
  id, name, short_description, description, status, image_url, video_url,
  config_type, uptime_percent, system_impact, created_at, formed_at, creator_id
) VALUES
  (
    1,
    'Кластер базы данных высокой доступности',
    'Многоузловая актив-актив архитектура кластера для работы без простоев...',
    'Активно-активный кластер из нескольких узлов обеспечивает работу без простоев во время обслуживания и автоматически переключается при отказе узла.',
    'published',
    'http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B01.jpg',
    'http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE1.mp4',
    'Clustering',
    99.99,
    60,
    '2024-09-10 10:00:00',
    '2024-09-12 12:00:00',
    1
  ),
  (
    2,
    'Узел веб-сервера Альфа',
    'Одноузловой периферийный веб-уровень для сравнения доступности...',
    'Одноузловой периферийный веб-уровень используется как базовая точка для сравнения доступности после репликации.',
    'published',
    'http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B02.jpg',
    'http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%203.mp4',
    'Single Node',
    99.99,
    21.50,
    '2024-09-11 10:00:00',
    '2024-09-13 12:00:00',
    1
  ),
  (
    3,
    'Глобальный балансировщик нагрузки',
    'Балансировщик Anycast между региональными репликами...',
    'Балансировщик Anycast между региональными репликами сохраняет доступность трафика при частичных сбоях.',
    'published',
    'http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B03.jpg',
    'http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%203.mp4',
    'Replication',
    99.95,
    10,
    '2024-09-12 10:00:00',
    '2024-09-14 12:00:00',
    1
  ),
  (
    4,
    'Основная реплика базы данных',
    'Синхронная пара основной базы и реплики с почти нулевым RPO...',
    'Синхронная пара основной базы и реплики с почти нулевым RPO для контролируемых тренировок переключения.',
    'published',
    'http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B04.jpg',
    'http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%204.mp4',
    'Replication',
    99.91,
    71.2,
    '2024-09-13 10:00:00',
    '2024-09-15 12:00:00',
    2
  ),
  (
    5,
    'Периферийный слой кэширования',
    'Уровень кэширования CDN скрывает кратковременные сбои источника...',
    'Уровень кэширования CDN снижает нагрузку на источник и скрывает кратковременные сбои источника.',
    'published',
    '',
    '',
    'Clustering',
    99.88,
    16,
    '2024-09-14 10:00:00',
    '2024-09-16 12:00:00',
    2
  ),
  (
    6,
    'Агент мониторинга',
    'Черновой агент проверок для следующего сценария оценки...',
    'Черновой одноузловой агент проверок подготовлен для следующего сценария оценки доступности.',
    'draft',
    'http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B06.jpg',
    'http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%206.mp4',
    'Single Node',
    99.50,
    89.11,
    '2024-09-18 10:00:00',
    NULL,
    3
  ),
  (
    7,
    'Устаревший брокер сообщений',
    'Устаревший брокер скрыт из интерфейса...',
    'Устаревший брокер сохранён только для истории и не должен отображаться в интерфейсе.',
    'deleted',
    'http://localhost:9000/media/%D0%BF%D0%BB%D0%B8%D1%82%D0%BA%D0%B07.jpg',
    'http://localhost:9000/media/%D0%B2%D0%B8%D0%B4%D0%B5%D0%BE%206.mp4',
    'Single Node',
    97.10,
    23.00,
    '2024-09-01 10:00:00',
    '2024-09-02 12:00:00',
    1
  );

INSERT INTO component_likes (id, user_id, component_id) VALUES
  (1, 1, 1),
  (2, 2, 1),
  (3, 3, 1),
  (4, 1, 2),
  (5, 2, 2),
  (6, 1, 3),
  (7, 3, 3),
  (8, 2, 4),
  (9, 1, 5);

SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));
SELECT setval(pg_get_serial_sequence('components', 'id'), (SELECT MAX(id) FROM components));
SELECT setval(pg_get_serial_sequence('component_likes', 'id'), (SELECT MAX(id) FROM component_likes));
