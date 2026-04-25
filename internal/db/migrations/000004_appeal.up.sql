CREATE TABLE comment_apppeal (
    comment_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    appeal_id INT NOT NULL,
    parent_id INT,
    link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    text TEXT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_length_text CHECK (char_length(text) <= 1000),
    CONSTRAINT fk_comment_task FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    CONSTRAINT fk_comment_parent FOREIGN KEY (parent_id) REFERENCES comment_task(comment_id) ON DELETE CASCADE
);
