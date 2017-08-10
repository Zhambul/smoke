CREATE TABLE account (
  id         SERIAL PRIMARY KEY,
  first_name VARCHAR NOT NULL,
  last_name  VARCHAR NOT NULL,
  chat_id    INT     NOT NULL UNIQUE
);

CREATE TABLE "group" (
  id                 SERIAL PRIMARY KEY,
  name               VARCHAR NOT NULL UNIQUE,
  creator_account_id INT     NOT NULL,
  uuid               UUID    NOT NULL UNIQUE
);

CREATE TABLE group_account (
  account_id INT NOT NULL,
  group_id   INT NOT NULL
);

CREATE TABLE message (
  id              SERIAL PRIMARY KEY,
  text            VARCHAR NOT NULL,
  from_account_id INT     NOT NULL
);

CREATE TABLE message_recipients (
  message_id    INT NOT NULL,
  to_account_id INT NOT NULL
);

ALTER TABLE group_account
  ADD FOREIGN KEY (group_id) REFERENCES "group" (id) ON DELETE CASCADE;
ALTER TABLE group_account
  ADD FOREIGN KEY (account_id) REFERENCES account (id);
ALTER TABLE "group"
  ADD FOREIGN KEY (creator_account_id) REFERENCES account (id);
ALTER TABLE group_account
  ADD CONSTRAINT uq_group_account UNIQUE (account_id, group_id);
ALTER TABLE message
  ADD FOREIGN KEY (from_account_id) REFERENCES account (id);
ALTER TABLE message_recipients
  ADD FOREIGN KEY (message_id) REFERENCES message (id);
ALTER TABLE message_recipients
  ADD FOREIGN KEY (to_account_id) REFERENCES account (id);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";