BEGIN;

TRUNCATE TABLE "user", board, section, task, board_template, section_template RESTART IDENTITY CASCADE;

INSERT INTO "user" (display_name, password_hash, email, avatar) VALUES
('Артем Бусыгин', 'hash_1', 'artem@nexus.com', 'https://s3.nexus.com/avatars/artem.png'),
('Иван Иванов', 'hash_2', 'ivan@nexus.com', ''),
('Елена Смирнова', 'hash_3', 'elena@nexus.com', 'https://s3.nexus.com/avatars/elena.png'),
('Дмитрий Тестиров', 'hash_4', 'qa@nexus.com', ''),
('Анна Дизайнер', 'hash_5', 'anna@nexus.com', 'https://s3.nexus.com/avatars/anna.png'),
('Сергей Девопс', 'hash_6', 'devops@nexus.com', '');

INSERT INTO board DEFAULT VALUES;
INSERT INTO board DEFAULT VALUES;
INSERT INTO board DEFAULT VALUES;

INSERT INTO board_version (board_id, board_name, description_board, url_path_background, valid_from, valid_to) VALUES
(1, 'BuisnesClac', 'Старое название', '#cccccc', now() - interval '30 days', now()),
(1, 'NeXus Core', 'Главная доска разработки проекта', '#1e1e2e', now(), NULL),
(2, 'Маркетинг 2026', 'Продвижение и реклама', 'https://s3.nexus.com/bg/market.png', now(), NULL),
(3, 'UI/UX Design', 'Кнопки, цвета, шрифты', '#ffcc00', now(), NULL);

INSERT INTO member_board (board_id, user_id, is_like, is_archive, level_member) VALUES
(1, 1, true, false, 'creater'),
(1, 2, false, false, 'editor'),
(1, 3, true, false, 'admin'),
(1, 4, false, false, 'viewer'),
(1, 6, false, false, 'editor'),
(2, 3, true, false, 'creater'),
(2, 5, false, false, 'editor'),
(3, 5, true, false, 'creater'),
(3, 1, false, false, 'viewer');

INSERT INTO section (board_id) VALUES (1), (1), (1), (1);
INSERT INTO section (board_id) VALUES (2), (2);
INSERT INTO section (board_id) VALUES (3), (3);

INSERT INTO section_version (section_id, section_name, position, is_mandatory, max_tasks, valid_from, valid_to) VALUES
(1, 'Куча идей', 1, false, NULL, now() - interval '10 days', now()),
(1, 'Бэклог', 1, true, NULL, now(), NULL),
(2, 'В разработке', 2, true, 5, now(), NULL),
(3, 'Ревью кода', 3, true, 3, now(), NULL),
(4, 'Готово', 4, false, NULL, now(), NULL),
(5, 'Контент-план', 1, false, NULL, now(), NULL),
(6, 'Опубликовано', 2, false, NULL, now(), NULL),
(7, 'Черновики', 1, false, NULL, now(), NULL),
(8, 'Утверждено', 2, false, NULL, now(), NULL);

INSERT INTO task (author_id) VALUES
(1), (1), (2), (3), (4), (6), (5), (5), (1), (3);

INSERT INTO task_version (task_id, section_id, title, description, position, due_date, valid_from, valid_to) VALUES
(1, 1, 'Спроектировать БД', 'Сделать BCNF', 1, now() + interval '1 day', now() - interval '5 days', now() - interval '2 days'),
(1, 2, 'Спроектировать БД', 'Сделать BCNF', 1, now() + interval '1 day', now() - interval '2 days', now() - interval '1 day'),
(1, 4, 'Спроектировать БД', 'Сделать BCNF. Все отлично!', 1, now() + interval '1 day', now() - interval '1 day', NULL),
(2, 2, 'Написать модели на Go', 'Структуры для работы с БД NeXus', 2, now() + interval '5 days', now(), NULL),
(3, 1, 'Настроить Docker', 'Написать Dockerfile и docker-compose', 2, NULL, now(), NULL),
(4, 1, 'Написать Unit-тесты', 'Покрытие 80%', 3, now() - interval '1 day', now(), NULL),
(5, 5, 'Пост про запуск NeXus', 'Текст для Telegram', 1, now() + interval '3 days', now(), NULL),
(6, 6, 'Анонс фичи', 'Уже выложили', 1, NULL, now(), NULL),
(7, 7, 'Иконки для меню', 'Сделать SVG', 1, NULL, now(), NULL),
(8, 8, 'Логотип NeXus', 'Финальный вектор', 1, NULL, now(), NULL),
(9, 3, 'Проверить PR #42', 'Смотрит Иван', 1, now() + interval '1 day', now(), NULL),
(10, 2, 'Интеграция с S3', 'Сохранение аватарок', 3, now() + interval '10 days', now(), NULL);

INSERT INTO subtask (task_id, description, is_done, position) VALUES
(2, 'Структура User', true, 1),
(2, 'Структура Board', true, 2),
(2, 'Теги JSON', false, 3),
(3, 'Установить Docker', true, 1),
(3, 'Написать Compose', false, 2);

INSERT INTO task_dependency (blocking_task_id, blocked_task_id) VALUES
(1, 2),
(3, 4);

INSERT INTO worker_task (assignee_id, task_id) VALUES
(1, 1), (1, 2), (6, 3), (4, 4), (5, 7), (5, 8), (1, 10);

INSERT INTO listener_task (listener_id, task_id) VALUES
(3, 1), (3, 2), (2, 2), (1, 5);

INSERT INTO comment_task (task_id, parent_id, text) VALUES
(2, NULL, 'Артем, как успехи с моделями?'),
(2, 1, 'Почти готово, добавляю JSON теги.'),
(4, NULL, 'Сроки горят!'),
(10, NULL, 'Ключи от S3 лежат в .env файле');

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
