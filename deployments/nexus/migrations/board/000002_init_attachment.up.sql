CREATE TABLE attachment(
    attachment_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    attachment_link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    task_link UUID NOT NULL,

    attachment_name TEXT DEFAULT '' NOT NULL,
    attachment_path TEXT DEFAULT '' NOT NULL,
    position INT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_task_link FOREIGN KEY (task_link) REFERENCES task(task_link) ON DELETE CASCADE
);
