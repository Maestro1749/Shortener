CREATE TABLE Links(
    id SERIAL PRIMARY KEY,
    link TEXT UNIQUE,
    link_clicks INT
);