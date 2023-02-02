# Links
Links page inspired by MySpace

## Design
- A monolith providing Linktree-like features:
    - Register an account and log in/out with it
    - Edit a list of links associated with your account
    - Other people can see this list of links by going your account's page
- All forms are protected via a CSRF token which is a 
    by-product of a CSRF session cookie. See the crypto/csrf package.
- Sessions are stored in Redis as encoding/gob encoded byte arrays.
    See the account/session package.
- Sha256 is used to hash passwords for simplicity's sake,
    but it is not that good for hashing passwords, Argon2 would be a better choice.
- Chi is used as the router, along with various middlewares,
    like CSRF injectors/validators and session validators, and 
    one middleware that injects the timestamp of when we started
    handling the request into the context.
- Sqlx with pgx is used as the database driver. Not much to say about that,
    they do the job. Although I created a small migrator library because I wanted
    one that was minimal and accepted a DB connection directly instead of a DSN.
    I used squirrel in a couple of places to dynamically build SQL.
- For the frontend, I used the html/template package of the stdlib. This can be found
    in the views package, where I store the templates in .html files, and some static files,
    and they are all exposed via a handler.
- For development, we have a docker compose file that spins up Redis and Postgres
    instances. Then we can do a `go run .` to connect to them and we run our server
    pretty much instantly.
- For configuration, I used godotenv to load up the environment variables,
    and envconfig to parse the variables into a struct which we pass around
    in main.go to set up various handlers.
- For testing, I used the testify package alongside the standard testing package,
    mostly for the mock and require packages which provide useful helpers for testing.
