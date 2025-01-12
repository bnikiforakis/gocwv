CREATE TABLE crux_metrics (
                              id SERIAL PRIMARY KEY,
                              created_at timestamp DEFAULT (now()),
                              url TEXT NOT NULL,
                              metric TEXT NOT NULL,
                              score DOUBLE PRECISION,
                              good DOUBLE PRECISION,
                              needs_improvement DOUBLE PRECISION,
                              poor DOUBLE PRECISION
);

CREATE INDEX ON "crux_metrics" ("url", "metric");