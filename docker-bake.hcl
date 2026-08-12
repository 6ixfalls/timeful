group "default" {
  targets = ["image"]
}

# Populated by docker/metadata-action in CI.
target "docker-metadata-action" {}

target "image" {
  inherits   = ["docker-metadata-action"]
  context    = "."
  dockerfile = "Dockerfile"
}
