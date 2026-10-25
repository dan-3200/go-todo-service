env "local" {
  src = "file://database/schema.sql"
  url = getenv("ATLAS_DATABASE_URL")
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://database/migrations"
  }
}

env "ci" {
  src = "file://database/schema.sql"
  dev = "docker://postgres/16/dev?search_path=public"

  migration {
    dir = "file://database/migrations"
  }
}
