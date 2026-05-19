CREATE TYPE invite_status AS ENUM ('active', 'closed');

CREATE TABLE invite (
    invite_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    invite_link UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    board_link UUID NOT NULL,
    user_link UUID,
    default_role user_level DEFAULT 'viewer' NOT NULL,
    expire_time TIMESTAMPTZ,
    status invite_status DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT fk_invite_board FOREIGN KEY (board_link) REFERENCES board(link) ON DELETE CASCADE
);
