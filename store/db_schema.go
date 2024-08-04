package store

var schemaV0 = [...]string{
	`CREATE TABLE relation(
        parent INTEGER NOT NULL,
        child  INTEGER NOT NULL,
        PRIMARY KEY (parent, child),
        FOREIGN KEY(parent) REFERENCES component(id),
        FOREIGN KEY(child)  REFERENCES component(id)
    )`,
	`CREATE TABLE component(
        id        INTEGER PRIMARY KEY,
        type      TEXT NOT NULL,
        species   TEXT NOT NULL,
        token     TEXT NOT NULL,
        createdAt TEXT NOT NULL,
        notes     TEXT DEFAULT '' NOT NULL,
        gone      INTEGER DEFAULT 0 NOT NULL,
        CHECK(type IN ('SPORES', 'MYC', 'SPAWN', 'GROW')),
        CHECK(species <> ''),
        CHECK(token <> ''),
        CHECK(createdAt <> '')
    )`,
	`CREATE INDEX type ON component(type)`,
	`CREATE INDEX species ON component(species)`,
	`CREATE INDEX createdAt ON component(createdAt)`,
	`CREATE INDEX gone ON component(gone)`,
}

var schemaV1 = [...]string{
	`CREATE TABLE grow(
		id           INTEGER PRIMARY KEY,
		yield        INTEGER, -- weight in milligrams
		yieldComment TEXT DEFAULT '' NOT NULL,
		CHECK(yield >= 0),
		FOREIGN KEY(id) REFERENCES component(id)
	)`,
}
