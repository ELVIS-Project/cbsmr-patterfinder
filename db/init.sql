CREATE TABLE IF NOT EXISTS Piece (
		pid SERIAL PRIMARY KEY,
		fmt VARCHAR(8),
		data TEXT,
		path TEXT)
;
