-- Create "todo" table
CREATE TABLE "todo" (
  "id" serial NOT NULL,
  "title" text NULL,
  "content" text NULL,
  "status" boolean NULL,
  "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
