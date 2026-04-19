CREATE TABLE IF NOT EXISTS games (
    id           INTEGER PRIMARY KEY,
    event        TEXT,
    site         TEXT,
    date         TEXT,
    round        TEXT,
    white        TEXT,
    black        TEXT,
    result       TEXT,
    white_elo    INTEGER,
    black_elo    INTEGER,
    eco          TEXT,
    opening      TEXT,
    pgn_raw      TEXT NOT NULL,
    ply_count    INTEGER NOT NULL,
    imported_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS moves (
    game_id    INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    ply        INTEGER NOT NULL,
    san        TEXT,
    uci        TEXT,
    fen_after  TEXT NOT NULL,
    PRIMARY KEY (game_id, ply)
);

CREATE INDEX IF NOT EXISTS idx_moves_game_ply ON moves(game_id, ply);

CREATE TABLE IF NOT EXISTS bookmarks (
    id         INTEGER PRIMARY KEY,
    game_id    INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    ply        INTEGER NOT NULL,
    note       TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bookmarks_game ON bookmarks(game_id);
