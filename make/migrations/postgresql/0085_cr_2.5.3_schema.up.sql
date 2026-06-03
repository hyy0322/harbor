CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_blob_digest_blob ON artifact_blob (digest_blob);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_digest_project_id ON artifact (digest, project_id);
