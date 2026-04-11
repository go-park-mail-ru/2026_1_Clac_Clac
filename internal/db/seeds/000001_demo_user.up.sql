-- Создания демо-пользователя
-- Запускать только после основных миграций
-- Использовать как скрипт, а не полноценную миграцию
-- Так как могут быть конфилкты на стороне БД
-- psql [DSN] -f ./internal/db/migrations/seeds/000001_demo_user.up.sql
--
-- Почта: demo@demo.ru
-- Пароль: 12345678

DO $$
DECLARE
    d_user_id INT;
    d_user_link UUID;

    b_id INT; b_link UUID;
    s_id INT; s_link UUID;
    t_id INT; t_link UUID;
BEGIN
    INSERT INTO "user" (display_name, password_hash, email, avatar_key)
    VALUES (
        'Демо',
        '$2a$10$JSP2k5H45X9.iDiPYtdRl.vQ23OpGyKm.lULMxP961tCvOpcnBF.C',
        'demo@demo.ru',
        'https://s3.nexus.com/avatars/demo.png'
    )
    RETURNING user_id, link INTO d_user_id, d_user_link;

    -- ========================================================================
    -- ДОСКА 1: 🚀 Запуск своего стартапа (Pet-project)
    -- ========================================================================
    INSERT INTO board DEFAULT VALUES RETURNING board_id, link INTO b_id, b_link;
    INSERT INTO board_version (board_id, board_name, description_board, url_path_background, valid_from)
    VALUES (b_id, '🚀 Запуск своего стартапа', 'Делаем следующий Unicorn', '#2c3e50', now());

    INSERT INTO member_board (board_link, user_link, level_member, is_like)
    VALUES (b_link, d_user_link, 'creator', true);

    -- Секция: Идеи
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '💡 Идеи', 1, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Прикрутить AI', 'Все любят ИИ, надо добавить куда-нибудь', 1, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Монетизация', 'Freemium или подписка?', 2, now());

    -- Секция: В разработке
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, is_mandatory, valid_from) VALUES (s_link, '⚙️ В разработке', 2, true, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Спроектировать API', 'RESTful или GraphQL? Решить до пятницы', 1, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);
    INSERT INTO subtask (task_id, description, is_done, position) VALUES
        (t_id, 'Написать Swagger', false, 1),
        (t_id, 'Описать модели', true, 2);
    INSERT INTO comment_task (task_id, text) VALUES (t_id, 'Решил делать REST, так надежнее.');

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Настроить CI/CD пайплайн', 'GitLab Actions + Docker', 2, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);

    -- Секция: Готово
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '✅ Готово', 3, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Купить домен', 'nexus-startup.io', 1, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);


    -- ========================================================================
    -- ДОСКА 2: 🏠 Ремонт квартиры
    -- ========================================================================
    INSERT INTO board DEFAULT VALUES RETURNING board_id, link INTO b_id, b_link;
    INSERT INTO board_version (board_id, board_name, description_board, url_path_background, valid_from)
    VALUES (b_id, '🏠 Ремонт', 'Смета, закупки, контроль', '#d35400', now());

    INSERT INTO member_board (board_link, user_link, level_member, is_like)
    VALUES (b_link, d_user_link, 'creator', false);

    -- Секция: Покупки
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '🛒 Покупки', 1, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Материалы для ванной', 'Плитка, затирка, клей', 1, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);
    INSERT INTO subtask (task_id, description, is_done, position) VALUES
        (t_id, 'Керамогранит 60x60 (10 пачек)', false, 1),
        (t_id, 'Влагостойкий гипсокартон', true, 2);

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Выбрать диван', 'Угловой, серый, раскладной', 2, now());

    -- Секция: В процессе
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '👷 В процессе', 2, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, due_date, valid_from)
    VALUES (t_link, s_link, 'Электрика', 'Проложить провода по потолку', 1, now() + interval '2 days', now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);

    -- Секция: Завершено
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '🏁 Завершено', 3, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Демонтаж старых обоев', '', 1, now());


    -- ========================================================================
    -- ДОСКА 3: 📚 Изучение Go & System Design
    -- ========================================================================
    INSERT INTO board DEFAULT VALUES RETURNING board_id, link INTO b_id, b_link;
    INSERT INTO board_version (board_id, board_name, description_board, url_path_background, valid_from)
    VALUES (b_id, '📚 Развитие: Go & System Design', 'План обучения на год', '#27ae60', now());

    INSERT INTO member_board (board_link, user_link, level_member, is_like)
    VALUES (b_link, d_user_link, 'creator', true);

    -- Секция: Бэклог материалов
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, 'Бэклог материалов', 1, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Книга "High Performance Go"', 'Почитать на выходных', 1, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Курс по микросервисам', 'Найти хороший на Udemy', 2, now());

    -- Секция: Изучаю сейчас
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '🔥 Изучаю сейчас', 2, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Разобраться с Goroutines', 'Синхронизация, каналы, WaitGroup', 1, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);
    INSERT INTO subtask (task_id, description, is_done, position) VALUES
        (t_id, 'Прочитать официальную доку', true, 1),
        (t_id, 'Написать парсер-воркер', false, 2),
        (t_id, 'Понять как работает Mutex', false, 3);

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Изучить PostgreSQL индексы', 'B-Tree, Hash, GiST', 2, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);

    -- Секция: Применено на практике
    INSERT INTO section (board_link) VALUES (b_link) RETURNING section_id, section_link INTO s_id, s_link;
    INSERT INTO section_version (section_link, section_name, position, valid_from) VALUES (s_link, '🏆 Применено на практике', 3, now());

    INSERT INTO task (author_link) VALUES (d_user_link) RETURNING task_id, task_link INTO t_id, t_link;
    INSERT INTO task_version (task_link, section_link, title, description, position, valid_from)
    VALUES (t_link, s_link, 'Настроить чистую архитектуру', 'Уже внедрено в проекте NeXus', 1, now());
    INSERT INTO worker_task (assignee_id, task_id) VALUES (d_user_id, t_id);
    INSERT INTO comment_task (task_id, text) VALUES (t_id, 'Получилось отлично, тесты пишутся легко.');

END $$;
