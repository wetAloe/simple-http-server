-- Create a new UTF-8 `snippetbox` database.
CREATE DATABASE snippetbox;

CREATE TABLE snippets (
  id SERIAL NOT NULL PRIMARY KEY,
  title VARCHAR(100) NOT NULL,
  content TEXT NOT NULL,
  created TIMESTAMP NOT NULL,
  expires TIMESTAMP NOT NULL
);

-- Add an index on the created column.
CREATE INDEX idx_snippets_created ON snippets(created);


-- Add some dummy records (which we'll use in the next couple of chapters).
INSERT INTO snippets (title, content, created, expires) VALUES (
   'An old silent pond',
   'An old silent pond...\nA frog jumps into the pond,\nsplash! Silence again.\n\n– Matsuo Bashō',
   NOW(),
   NOW() + INTERVAL '1 year'
);

INSERT INTO snippets (title, content, created, expires) VALUES (
   'Over the wintry forest',
   'Over the wintry\nforest, winds howl in rage\nwith no leaves to blow.\n\n– Natsume Soseki',
   NOW(),
   NOW() + INTERVAL '1 year'
);

INSERT INTO snippets (title, content, created, expires) VALUES (
   'First autumn morning',
   'First autumn morning\nthe mirror I stare into\nshows my father''s face.\n\n– Murakami Kijo',
   NOW(),
   NOW() + INTERVAL '7 day'
);

CREATE USER web PASSWORD '123';
GRANT
    SELECT, INSERT, UPDATE, DELETE
    ON ALL TABLES IN SCHEMA public
    TO web;
GRANT
    USAGE
    ON ALL SEQUENCES IN SCHEMA public
    TO web;