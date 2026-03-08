CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated_at = now();
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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

CREATE TRIGGER set_user_updated_at
BEFORE UPDATE ON "user"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

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
    url_path_background TEXT DEFAULT '',

    valid_from TIMESTAMPTZ DEFAULT now() NOT NULL,
    valid_to TIMESTAMPTZ,

    CONSTRAINT check_length_board_name CHECK (char_length(board_name) <= 255),
    CONSTRAINT check_length_description_board CHECK (char_length("description_board") <= 1000),
    CONSTRAINT check_board_dates CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT fk_version_board FOREIGN KEY (board_id) REFERENCES board(board_id) ON DELETE CASCADE
);

CREATE TYPE user_level AS ENUM ('viewer', 'editor', 'admin', 'creater')

CREATE TABLE member_board (
    board_id INT NOT NULL,
    user_id INT NOT NULL,

    is_like BOOLEAN DEFAULT false NOT NULL,
    is_archive BOOLEAN DEFAULT false NOT NULL,
    level_member user_level DEFAULT 'viewer' NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_member_board PRIMARY KEY (board_id, user_id),
    CONSTRAINT fk_member_user FOREIGN KEY (user_id) REFERENCES "user"(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_member_board FOREIGN KEY (board_id) REFERENCES board(board_id) ON DELETE CASCADE
);

CREATE TRIGGER set_member_board_updated_at
BEFORE UPDATE ON member_board
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE section (
    section_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    board_id INT NOT NULL,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_section_board FOREIGN KEY (board_id) REFERENCES board(board_id) ON DELETE CASCADE
);

CREATE TABLE section_version (
    version_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id INT NOT NULL,

    section_name TEXT DEFAULT '' NOT NULL,
    position INT NOT NULL,
    is_mandatory BOOLEAN DEFAULT false NOT NULL,
    max_tasks INT,

    valid_from TIMESTAMPTZ DEFAULT now() NOT NULL,
    valid_to TIMESTAMPTZ,

    CONSTRAINT check_length_section_name CHECK (char_length(section_name) <= 255),
    CONSTRAINT check_min_tasks CHECK (max_tasks IS NULL or max_tasks > 0),
    CONSTRAINT check_section_dates CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT fk_version_section FOREIGN KEY (section_id) REFERENCES section(section_id) ON DELETE CASCADE
);

CREATE TABLE task (
    task_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id INT,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_task_author FOREIGN KEY (author_id) REFERENCES "user"(user_id) ON DELETE SET NULL
);


CREATE TABLE task_version (
    version_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id INT NOT NULL,
    section_id INT NOT NULL,

    title TEXT DEFAULT '' NOT NULL,
    "description" TEXT DEFAULT '' NOT NULL,
    position INT NOT NULL,
    due_date TIMESTAMPTZ,

    valid_from TIMESTAMPTZ DEFAULT now() NOT NULL,
    valid_to TIMESTAMPTZ,

    CONSTRAINT check_length_title CHECK (char_length(title) <= 255),
    CONSTRAINT check_length_description CHECK (char_length("description") <= 1000),
    CONSTRAINT check_due_date CHECK (due_date IS NULL or due_date >= valid_from),
    CONSTRAINT check_task_dates CHECK (valid_to IS NULL OR valid_to > valid_from),
    CONSTRAINT fk_version_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    CONSTRAINT fk_version_section FOREIGN KEY (section_id) REFERENCES section(section_id) ON DELETE CASCADE
);

CREATE TABLE subtask (
    subtask_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id INT NOT NULL,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    "description" TEXT NOT NULL,
    is_done BOOLEAN DEFAULT false NOT NULL,
    position INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_description_length CHECK (char_length("description") <= 500),
    CONSTRAINT fk_subtask_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE
);

CREATE TRIGGER set_subtask_updated_at
BEFORE UPDATE ON subtask
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE task_dependency (
    blocking_task_id INT NOT NULL,
    blocked_task_id INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_no_self_block CHECK (blocking_task_id != blocked_task_id),
    CONSTRAINT pk_task_dependency PRIMARY KEY (blocking_task_id, blocked_task_id),
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

CREATE TABLE listener_task (
    listener_id INT NOT NULL,
    task_id INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_listener_task PRIMARY KEY (listener_id, task_id),
    CONSTRAINT fk_lt_listener FOREIGN KEY (listener_id) REFERENCES "user"(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_lt_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE
);

CREATE TABLE comment_task (
    comment_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id INT NOT NULL,
    parent_id INT,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    "text" TEXT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_length_text CHECK (char_length("text") <= 1000),
    CONSTRAINT fk_comment_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_parent FOREIGN KEY (parent_id) REFERENCES comment_task(comment_id) ON DELETE CASCADE
);


CREATE TRIGGER set_comment_task_updated_at
BEFORE UPDATE ON comment_task
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE board_template(
    btemplate_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id INT,

    template_name TEXT DEFAULT '' NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_length_btemplate_name CHECK (char_length("template_name") <= 255)
);

CREATE TRIGGER set_board_template_updated_at
BEFORE UPDATE ON board_template
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE section_template(
    stemplate_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    btemplate_id INT NOT NULL,

    position INT NOT NULL,
    is_mandatory BOOLEAN DEFAULT false NOT NULL,
    max_tasks INT,
    section_name TEXT DEFAULT '' NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_stemplate_name_length CHECK (char_length("section_name") <= 255),
    CONSTRAINT check_min_tasks CHECK (max_tasks IS NULL or max_tasks > 0),
    CONSTRAINT fk_board_template FOREIGN KEY (btemplate_id) REFERENCES board_template(btemplate_id) ON DELETE CASCADE
);

CREATE TRIGGER set_section_template_updated_at
BEFORE UPDATE ON section_template
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
