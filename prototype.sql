
CREATE DATABASE fastpound;

INSERT INTO system.locations VALUES ('region', 'us-east1', 33.836082, -81.163727);
INSERT INTO system.locations VALUES ('region', 'us-central1', 42.032974, -93.581543);
INSERT INTO system.locations VALUES ('region', 'us-west1', 43.804133, -120.554201);

CREATE TABLE content(
    id UUID NOT NULL,
    name STRING,
    scoring INT,
    content JSONB,
    created_at TIMESTAMP NULL,
    categories VARCHAR,
    PRIMARY KEY (id)
);

CREATE TABLE users(
    region VARCHAR,
    id UUID,
    preferences VARCHAR ARRAY, --TBD
    PRIMARY KEY (region, id)
) --PARTITION BY (region);
;

CREATE TABLE user_content_history(
    region VARCHAR,
    user_history_id UUID,
    content_id UUID references content (id),
    user_id UUID,
    ts TIMESTAMP,
    CONSTRAINT fk_users FOREIGN KEY (region, user_id) REFERENCES users (region, id),
    PRIMARY KEY (region, user_id, content_id, ts)
)

INSERT
INTO
    users (region, preferences, id)
VALUES
    ('us', ARRAY['memes', 'jokes'], gen_random_uuid());

INSERT
INTO
    content (id, name, scoring, content, created_at, categories)
VALUES
    (
        gen_random_uuid(),
        'Pooh',
        123,
        '{"url": "https://cockroachlabs.com"}'::JSONB,
        now(),
        'memes'
    );


SELECT
    c.scoring, c.content
FROM
    content AS c, users AS u AS OF SYSTEM TIME experimental_follower_read_timestamp()
WHERE
    c.types = ANY u.preferences
    AND c.id
        NOT IN (
                SELECT
                    h.content_id
                FROM
                    user_content_history AS h
                WHERE
                    h.content_id = c.id AND h.user_id = u.id
            )
ORDER BY
    scoring;