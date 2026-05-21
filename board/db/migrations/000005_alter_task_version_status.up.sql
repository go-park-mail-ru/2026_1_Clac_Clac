ALTER TABLE task_version
ADD COLUMN IF NOT EXISTS status BOOLEAN DEFAULT false NOT NULL;

CREATE OR REPLACE VIEW task_actual AS
SELECT
    t.task_id,
    t.task_link,
    t.author_link,
    t.created_at,
    v.section_link,
    v.executer_link,
    v.title,
    v.description,
    v.position,
    v.due_date,
    v.start,
    v.status
FROM task t
JOIN task_version v
  ON v.task_link = t.task_link
 AND v.valid_to IS NULL;
