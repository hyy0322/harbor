BEGIN;
    SET lock_timeout = '1s';

    CREATE TABLE IF NOT EXISTS signature
    (
        id               SERIAL PRIMARY KEY NOT NULL,
        repository_id    int NOT NULL,
        artifact_id      int NOT NULL,
        digest           text NOT NULL,
        key_id           text NOT NULL,
        key_region       text NOT NULL,
        algorithm_method text NOT NULL,
        signature        text,
        err_msg          text,
        creation_time    timestamp default CURRENT_TIMESTAMP,
        update_time      timestamp default CURRENT_TIMESTAMP,
        CONSTRAINT unique_signature_artifact_key_algorithm UNIQUE (artifact_id, key_id, algorithm_method),
        FOREIGN KEY (artifact_id) REFERENCES artifact(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_signature_repository_digest
        ON signature (repository_id, digest);

    DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_trigger
            WHERE tgname = 'signature_update_time_at_modtime'
              AND tgrelid = 'signature'::regclass
        ) THEN
            CREATE TRIGGER signature_update_time_at_modtime
                BEFORE UPDATE ON signature
                FOR EACH ROW
                EXECUTE PROCEDURE update_update_time_at_column();
        END IF;
    END $$;
END;
