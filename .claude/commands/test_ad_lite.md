Run the full AdLite test suite (Go backend integration tests + Vue frontend unit tests).

## Steps

1. **Ensure infrastructure is running**
   ```
   cd /Users/apollo/Projects/ad_lite/server && docker compose up -d
   ```
   Wait up to 10 seconds for Postgres to be ready:
   ```
   until docker compose exec -T postgres pg_isready -U postgres -q; do sleep 1; done
   ```

2. **Create the test database if it doesn't exist**
   ```
   docker compose exec -T postgres psql -U postgres -c "CREATE DATABASE ad_lite_test;" 2>/dev/null || true
   ```

3. **Run Go integration tests**
   ```
   cd /Users/apollo/Projects/ad_lite/server && \
   TEST_DATABASE_URL="postgres://adlite:adlite@localhost:5432/ad_lite_test?sslmode=disable" \
   TEST_REDIS_URL="redis://localhost:6379/1" \
   go test -p 1 ./... -v -count=1 -timeout 120s
   ```

4. **Run Vue/Nuxt frontend unit tests** (if Vitest is configured)
   ```
   cd /Users/apollo/Projects/ad_lite/portal && npm run test --if-present
   ```

5. **Report results** — summarise pass/fail counts for each package. Call out any failures with the full error message so they can be fixed immediately.

## Notes
- The Go tests require docker compose to be running (`make up` in the server directory).
- Tests use a separate `ad_lite_test` database and Redis DB index 1 — they never touch the dev database.
- If a test fails due to a missing `ad_lite_test` database, step 2 above creates it automatically.
- Use `-run TestFoo` to run a single test: `go test ./internal/bookings/... -run TestCreateBooking`.
