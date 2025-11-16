CREATE TABLE IF NOT EXISTS users(
    user_id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    is_active BOOLEAN NOT NULL,
    team_name TEXT REFERENCES teams(team_name)
);
