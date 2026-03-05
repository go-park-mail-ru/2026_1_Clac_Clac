CREATE TABLE "user" (
    user_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    avatar TEXT DEFAULT '',

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE board (
    board_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE board_version (
    version_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    board_id INT NOT NULL,

    board_name TEXT DEFAULT '' NOT NULL,
    description_board TEXT DEFAULT '',
    background TEXT DEFAULT '',

    valid_from TIMESTAMPTZ DEFAULT now() NOT NULL,
    valid_to TIMESTAMPTZ,

    CONSTRAINT check_board_dates CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT fk_version_board FOREIGN KEY (board_id) REFERENCES board(board_id) ON DELETE CASCADE
);

CREATE TABLE member_board (
    board_id INT NOT NULL,
    user_id INT NOT NULL,

    is_like BOOLEAN DEFAULT false NOT NULL,
    is_archive BOOLEAN DEFAULT false NOT NULL,
    level_member INT DEFAULT 1 NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_member_board PRIMARY KEY (board_id, user_id),
    CONSTRAINT fk_member_user FOREIGN KEY (user_id) REFERENCES "user"(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_member_board FOREIGN KEY (board_id) REFERENCES board(board_id) ON DELETE CASCADE
);

CREATE TABLE section (
    section_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    board_id INT NOT NULL,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    CONSTRAINT fk_section_board FOREIGN KEY (board_id) REFERENCES board(board_id) ON DELETE CASCADE
);

CREATE TABLE section_version (
    version_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id INT NOT NULL,

    section_name TEXT NOT NULL,
    position INT NOT NULL,
    is_mandatory BOOLEAN DEFAULT false NOT NULL,
    max_tasks INT,

    valid_from TIMESTAMPTZ DEFAULT now() NOT NULL,
    valid_to TIMESTAMPTZ,

    CONSTRAINT check_section_dates CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT fk_version_section FOREIGN KEY (section_id) REFERENCES section(section_id) ON DELETE CASCADE
);

CREATE TABLE task (
    task_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id INT,
    section_id INT NOT NULL,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_task_author FOREIGN KEY (author_id) REFERENCES "user"(user_id) ON DELETE SET NULL,
    CONSTRAINT fk_task_section FOREIGN KEY (section_id) REFERENCES section(section_id) ON DELETE CASCADE
);


CREATE TABLE task_version (
    version_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id INT NOT NULL,
    section_id INT NOT NULL,

    title TEXT NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    position INT NOT NULL,
    task_start_at TIMESTAMPTZ,
    due_date TIMESTAMPTZ,

    valid_from TIMESTAMPTZ DEFAULT now() NOT NULL,
    valid_to TIMESTAMPTZ,

    CONSTRAINT check_task_dates CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT fk_tversion_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    CONSTRAINT fk_tversion_section FOREIGN KEY (section_id) REFERENCES section(section_id) ON DELETE CASCADE
);

CREATE TABLE subtask (
    subtask_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id INT NOT NULL,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    description TEXT NOT NULL,
    is_done BOOLEAN DEFAULT false NOT NULL,
    position INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_subtask_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE
);

CREATE TABLE task_dependency (
    blocking_task_id INT NOT NULL,
    blocked_task_id INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_task_dependency PRIMARY KEY (blocking_task_id, blocked_task_id),
    CONSTRAINT check_no_self_block CHECK (blocking_task_id != blocked_task_id),

    CONSTRAINT fk_td_blocking FOREIGN KEY (blocking_task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    CONSTRAINT fk_td_blocked FOREIGN KEY (blocked_task_id) REFERENCES task(task_id) ON DELETE CASCADE
);

CREATE TABLE worker_task (
    assignee_id INT NOT NULL,
    task_id INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_worker_task PRIMARY KEY (assignee_id, task_id),
    CONSTRAINT fk_wt_user FOREIGN KEY (assignee_id) REFERENCES "user"(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_wt_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE
);

CREATE TABLE comment_task (
    comment_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id INT NOT NULL,
    parent_id INT,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    text TEXT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_comment_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_parent FOREIGN KEY (parent_id) REFERENCES comment_task(comment_id) ON DELETE CASCADE
);
