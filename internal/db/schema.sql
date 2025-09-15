CREATE TABLE IF NOT EXISTS characters (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT,
    species TEXT,
    type TEXT,
    gender TEXT,
    image TEXT,
    url TEXT,
    created TIMESTAMPTZ NOT NULL,
    origin_id INT,
    location_id INT
);