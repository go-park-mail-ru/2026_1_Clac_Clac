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
    description_user TEXT DEFAULT '' NOT NULL,
    email TEXT NOT NULL UNIQUE,
    avatar_key TEXT DEFAULT '' NOT NULL,

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    CONSTRAINT check_display_name CHECK (char_length(display_name) <= 128),
    CONSTRAINT check_length_description CHECK (char_length(description_user) <= 1000)
);

CREATE TRIGGER set_user_updated_at
BEFORE UPDATE ON "user"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
