BEGIN;

-- Очищаем все затронутые таблицы
TRUNCATE TABLE "user", board, board_version, member_board, section, section_version, task, task_version, subtask, task_dependency, worker_task, listener_task, comment_task, board_template, section_template RESTART IDENTITY CASCADE;

-- 1. Пользователи (заменено avatar на avatar_key)
INSERT INTO "user" (display_name, password_hash, email, avatar_key) VALUES
('Артем Бусыгин', 'hash_1', 'artem@nexus.com', 'https://s3.nexus.com/avatars/artem.png'),
('Иван Иванов', 'hash_2', 'ivan@nexus.com', ''),
('Елена Смирнова', 'hash_3', 'elena@nexus.com', 'https://s3.nexus.com/avatars/elena.png'),
('Дмитрий Тестиров', 'hash_4', 'qa@nexus.com', ''),
('Анна Дизайнер', 'hash_5', 'anna@nexus.com', 'https://s3.nexus.com/avatars/anna.png'),
('Сергей Девопс', 'hash_6', 'devops@nexus.com', '');

-- 2. Доски
INSERT INTO board DEFAULT VALUES; -- 1
INSERT INTO board DEFAULT VALUES; -- 2
INSERT INTO board DEFAULT VALUES; -- 3

INSERT INTO board_version (board_id, board_name, description_board, url_path_background, valid_from, valid_to) VALUES
(1, 'BuisnesClac', 'Старое название', '#cccccc', now() - interval '30 days', now()),
(1, 'NeXus Core', 'Главная доска разработки проекта', '#1e1e2e', now(), NULL),
(2, 'Маркетинг 2026', 'Продвижение и реклама', 'https://s3.nexus.com/bg/market.png', now(), NULL),
(3, 'UI/UX Design', 'Кнопки, цвета, шрифты', '#ffcc00', now(), NULL);

-- 3. Участники досок (Используем view board_actual для удобства)
INSERT INTO member_board (board_link, user_link, level_member, is_like) VALUES
((SELECT link FROM board_actual WHERE name = 'NeXus Core' LIMIT 1), (SELECT link FROM "user" WHERE email = 'artem@nexus.com'), 'creator', true),
((SELECT link FROM board_actual WHERE name = 'NeXus Core' LIMIT 1), (SELECT link FROM "user" WHERE email = 'ivan@nexus.com'), 'admin', false),
((SELECT link FROM board_actual WHERE name = 'NeXus Core' LIMIT 1), (SELECT link FROM "user" WHERE email = 'qa@nexus.com'), 'viewer', false),
((SELECT link FROM board_actual WHERE name = 'Маркетинг 2026' LIMIT 1), (SELECT link FROM "user" WHERE email = 'elena@nexus.com'), 'creator', true),
((SELECT link FROM board_actual WHERE name = 'UI/UX Design' LIMIT 1), (SELECT link FROM "user" WHERE email = 'anna@nexus.com'), 'creator', true),
((SELECT link FROM board_actual WHERE name = 'UI/UX Design' LIMIT 1), (SELECT link FROM "user" WHERE email = 'artem@nexus.com'), 'admin', false);

-- 4. Секции (в DDL board_link - это UUID, достаем его через SELECT)
INSERT INTO section (board_link) VALUES
((SELECT link FROM board WHERE board_id = 1)),
((SELECT link FROM board WHERE board_id = 1)),
((SELECT link FROM board WHERE board_id = 1)),
((SELECT link FROM board WHERE board_id = 1)),
((SELECT link FROM board WHERE board_id = 2)),
((SELECT link FROM board WHERE board_id = 2)),
((SELECT link FROM board WHERE board_id = 3)),
((SELECT link FROM board WHERE board_id = 3));

-- Версии секций (section_link - это UUID)
INSERT INTO section_version (section_link, section_name, position, is_mandatory, max_tasks, valid_from, valid_to) VALUES
((SELECT section_link FROM section WHERE section_id = 1), 'Куча идей', 1, false, NULL, now() - interval '10 days', now()),
((SELECT section_link FROM section WHERE section_id = 1), 'Бэклог', 1, true, NULL, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 2), 'В разработке', 2, true, 5, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 3), 'Ревью кода', 3, true, 3, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 4), 'Готово', 4, false, NULL, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 5), 'Контент-план', 1, false, NULL, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 6), 'Опубликовано', 2, false, NULL, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 7), 'Черновики', 1, false, NULL, now(), NULL),
((SELECT section_link FROM section WHERE section_id = 8), 'Утверждено', 2, false, NULL, now(), NULL);

-- 5. Задачи (author_link - это UUID)
INSERT INTO task (author_link) VALUES
((SELECT link FROM "user" WHERE user_id = 1)),
((SELECT link FROM "user" WHERE user_id = 1)),
((SELECT link FROM "user" WHERE user_id = 2)),
((SELECT link FROM "user" WHERE user_id = 3)),
((SELECT link FROM "user" WHERE user_id = 4)),
((SELECT link FROM "user" WHERE user_id = 6)),
((SELECT link FROM "user" WHERE user_id = 5)),
((SELECT link FROM "user" WHERE user_id = 5)),
((SELECT link FROM "user" WHERE user_id = 1)),
((SELECT link FROM "user" WHERE user_id = 3));

-- Версии задач (task_link и section_link - это UUID)
INSERT INTO task_version (task_link, section_link, title, description, position, due_date, valid_from, valid_to) VALUES
((SELECT task_link FROM task WHERE task_id = 1), (SELECT section_link FROM section WHERE section_id = 1), 'Спроектировать БД', 'Сделать BCNF', 1, now() + interval '1 day', now() - interval '5 days', now() - interval '2 days'),
((SELECT task_link FROM task WHERE task_id = 1), (SELECT section_link FROM section WHERE section_id = 2), 'Спроектировать БД', 'Сделать BCNF', 1, now() + interval '1 day', now() - interval '2 days', now() - interval '1 day'),
((SELECT task_link FROM task WHERE task_id = 1), (SELECT section_link FROM section WHERE section_id = 4), 'Спроектировать БД', 'Сделать BCNF. Все отлично!', 1, now() + interval '1 day', now() - interval '1 day', NULL),
((SELECT task_link FROM task WHERE task_id = 2), (SELECT section_link FROM section WHERE section_id = 2), 'Написать модели на Go', 'Структуры для работы с БД NeXus', 2, now() + interval '5 days', now(), NULL),
((SELECT task_link FROM task WHERE task_id = 3), (SELECT section_link FROM section WHERE section_id = 1), 'Настроить Docker', 'Написать Dockerfile и docker-compose', 2, NULL, now(), NULL),
((SELECT task_link FROM task WHERE task_id = 4), (SELECT section_link FROM section WHERE section_id = 1), 'Написать Unit-тесты', 'Покрытие 80%', 3, now() + interval '1 day', now(), NULL),
((SELECT task_link FROM task WHERE task_id = 5), (SELECT section_link FROM section WHERE section_id = 5), 'Пост про запуск NeXus', 'Текст для Telegram', 1, now() + interval '3 days', now(), NULL),
((SELECT task_link FROM task WHERE task_id = 6), (SELECT section_link FROM section WHERE section_id = 6), 'Анонс фичи', 'Уже выложили', 1, NULL, now(), NULL),
((SELECT task_link FROM task WHERE task_id = 7), (SELECT section_link FROM section WHERE section_id = 7), 'Иконки для меню', 'Сделать SVG', 1, NULL, now(), NULL),
((SELECT task_link FROM task WHERE task_id = 8), (SELECT section_link FROM section WHERE section_id = 8), 'Логотип NeXus', 'Финальный вектор', 1, NULL, now(), NULL),
((SELECT task_link FROM task WHERE task_id = 9), (SELECT section_link FROM section WHERE section_id = 3), 'Проверить PR #42', 'Смотрит Иван', 1, now() + interval '1 day', now(), NULL),
((SELECT task_link FROM task WHERE task_id = 10), (SELECT section_link FROM section WHERE section_id = 2), 'Интеграция с S3', 'Сохранение аватарок', 3, now() + interval '10 days', now(), NULL);

-- 6. Подзадачи (Тут в DDL используется INT task_id, поэтому ваш код корректен)
INSERT INTO subtask (task_id, description, is_done, position) VALUES
(2, 'Структура User', true, 1),
(2, 'Структура Board', true, 2),
(2, 'Теги JSON', false, 3),
(3, 'Установить Docker', true, 1),
(3, 'Написать Compose', false, 2);

-- Зависимости
INSERT INTO task_dependency (blocking_task_id, blocked_task_id) VALUES
(1, 2),
(3, 4);

-- Исполнители (тут DDL тоже использует INT)
INSERT INTO worker_task (assignee_id, task_id) VALUES
(1, 1), (1, 2), (6, 3), (4, 4), (5, 7), (5, 8), (1, 10);

-- Наблюдатели (тут DDL тоже использует INT)
INSERT INTO listener_task (listener_id, task_id) VALUES
(3, 1), (3, 2), (2, 2), (1, 5);

-- Комментарии
INSERT INTO comment_task (task_id, parent_id, text) VALUES
(2, NULL, 'Артем, как успехи с моделями?'),
(2, 1, 'Почти готово, добавляю JSON теги.'),
(4, NULL, 'Сроки горят!'),
(10, NULL, 'Ключи от S3 лежат в .env файле');

-- Шаблоны
INSERT INTO board_template (author_id, template_name) VALUES
(1, 'Agile Sprint'),
(3, 'Roadmap Проекта');

INSERT INTO section_template (btemplate_id, position, is_mandatory, max_tasks, section_name) VALUES
(1, 1, true, NULL, 'To Do'),
(1, 2, true, 5, 'In Progress'),
(1, 3, true, NULL, 'Done'),
(2, 1, false, NULL, 'Q1 2026'),
(2, 2, false, NULL, 'Q2 2026');

COMMIT;
