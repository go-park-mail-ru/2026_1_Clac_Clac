TRUNCATE TABLE listener_task;

ALTER TABLE listener_task
    DROP CONSTRAINT fk_lt_task,
    DROP CONSTRAINT pk_listener_task;

ALTER TABLE listener_task
    DROP COLUMN listener_id,
    DROP COLUMN task_id,
    ADD COLUMN user_link UUID NOT NULL,
    ADD COLUMN task_link UUID NOT NULL;

ALTER TABLE listener_task
    ADD CONSTRAINT pk_listener_task PRIMARY KEY (listener_link, task_link),
    ADD CONSTRAINT fk_lt_task FOREIGN KEY (task_link) REFERENCES task(task_link) ON DELETE CASCADE;
ALTER TABLE listener_task
    ADD CONSTRAINT uq_listener_task_listener UNIQUE (user_link),
    ADD CONSTRAINT uq_listener_task_task UNIQUE (task_link);
