package store

// schema is the DDL for the gofind index.
//
// Two tables:
//
//   - files: one row per indexed file with metadata used for incremental
//     re-index later (mtime, size, sha256). The 'id' column is reused as
//     the FTS5 rowid below so a join is just rowid = id.
//
//   - files_fts: an FTS5 virtual table using the 'trigram' tokenizer.
//     Trigram is the right choice for code search because it matches on
//     substrings -- "etUserBy" will hit "GetUserByID". The default
//     unicode61 tokenizer would only match whole words; porter would
//     stem English words, which is wrong for source code.
//
// BM25 ranking is built into FTS5: ORDER BY bm25(files_fts) ASC ranks
// best matches first. SQLite returns BM25 scores as negative numbers
// (more negative = better match) -- a small gotcha worth remembering.
const schema = `
CREATE TABLE IF NOT EXISTS files (
    id          INTEGER PRIMARY KEY,
    path        TEXT NOT NULL UNIQUE,
    mtime_ns    INTEGER NOT NULL,
    size        INTEGER NOT NULL,
    sha256      TEXT NOT NULL,
    indexed_at  INTEGER NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS files_fts USING fts5(
    path UNINDEXED,
    content,
    tokenize = 'trigram'
);
`
