# coattails
golang backend 

# How to run coattails
To start the server (listening on port 8080), run `go run cmd/coattails/main.go` in the root of the repo

# Dependencies
Coattails has two dependencies that must be running to fully use the server:
- redis
- postgres

By default, coattails will connect to the postgres db instance at `localhost`, port `5432` with user `postgres`, password `bdc`, and db `wardrobe`.
If your local settings are in any way not the default, feel free to override them (run `go run cmd/coattails/testbench.go --help` to view all arguments you can override)

For redis, coattails will connect to `localhost` on port `6379` by default.

# Test

