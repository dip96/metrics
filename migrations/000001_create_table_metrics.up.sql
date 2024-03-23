CREATE TABLE IF NOT EXISTS metrics (
    id smallserial PRIMARY KEY,
    name_metric CHARACTER VARYING(100) UNIQUE,
    type CHARACTER VARYING(30) NOT NULL,
    delta bigint,
    value double precision
)