CREATE TYPE appeal_status AS ENUM ('new', 'in_progress', 'closed');

CREATE TYPE appeal_category AS ENUM ('bug', 'proposal', 'complaint');

CREATE TABLE appeal (
    appeal_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    appeal_link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    user_link UUID,
    supporter_link UUID,

    mail TEXT,
    display_name TEXT NOT NULL,

    "status" appeal_status DEFAULT 'new' NOT NULL,
    category appeal_category NOT NULL,

    "description" TEXT DEFAULT '' NOT NULL,
    attachment_key TEXT DEFAULT '',

    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT check_length_display_name CHECK (char_length(display_name) <= 128),
    CONSTRAINT check_length_appeal_desc CHECK (char_length(description) <= 2000)
);

CREATE TRIGGER set_appeal_updated_at
BEFORE UPDATE ON "appeal"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


CREATE TYPE support_role AS ENUM ('admin', 'support');

CREATE TABLE support (
    support_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    support_link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,

    user_link UUID NOT NULL REFERENCES "user"(link) ON DELETE CASCADE,

    "role" support_role DEFAULT 'support' NOT NULL,

    CONSTRAINT unique_support_user UNIQUE (user_link)
);
