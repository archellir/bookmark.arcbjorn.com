
CREATE TABLE "bookmarks" (
  "id" int generated always as identity PRIMARY KEY,
  "name" varchar NOT NULL,
  "url" varchar NOT NULL,
  "group_id" int DEFAULT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

COMMENT ON COLUMN "bookmarks"."name" IS 'Title of the web page document';
ALTER TABLE "bookmarks" ADD CONSTRAINT "unique_name_url" UNIQUE ("name", "url");

CREATE TABLE "tags" (
  "id" int generated always as identity PRIMARY KEY,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "bookmarks_tags" (
  "bookmark_id" int,
  "tag_id" int,
  PRIMARY KEY ("bookmark_id", "tag_id")
);

ALTER TABLE "bookmarks_tags" ADD FOREIGN KEY ("bookmark_id") REFERENCES "bookmarks" ("id") ON DELETE CASCADE;
ALTER TABLE "bookmarks_tags" ADD FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON DELETE CASCADE;

CREATE TABLE "groups" (
  "id" int generated always as identity PRIMARY KEY,
  "name" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "bookmarks" ADD FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE SET NULL;

CREATE TABLE "users" (
  "id" int generated always as identity PRIMARY KEY,
  "username" varchar,
  "hashed_password" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);
