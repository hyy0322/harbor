CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_blob_digest_blob ON artifact_blob (digest_blob);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_digest_project_id ON artifact (digest, project_id);

BEGIN;
    SET lock_timeout = '1s';
    SET statement_timeout = '5s';
ALTER TABLE tag ADD COLUMN IF NOT EXISTS description TEXT;
END;
