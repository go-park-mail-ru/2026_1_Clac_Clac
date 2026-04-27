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
