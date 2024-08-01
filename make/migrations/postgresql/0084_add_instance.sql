/*
Add new column instance for artifact_blob table to work with multi-bucket
*/
ALTER TABLE artifact_blob ADD COLUMN IF NOT EXISTS instance varchar(255);

/*
set value for instance
then set column instance as not null
*/
UPDATE artifact_blob SET instance = '0' WHERE instance IS NULL;

ALTER TABLE artifact_blob ALTER COLUMN instance SET NOT NULL;

ALTER TABLE artifact_blob DROP CONSTRAINT unique_artifact_blob;
ALTER TABLE artifact_blob ADD CONSTRAINT unique_artifact_blob UNIQUE (digest_af, digest_blob, instance);

/*
Add new column instance for blob table to work with multi-bucket
*/
ALTER TABLE blob ADD COLUMN IF NOT EXISTS instance varchar(255);

/*
set value for instance
then set column instance as not null
*/
UPDATE blob SET instance = '0' WHERE instance IS NULL;

ALTER TABLE blob ALTER COLUMN instance SET NOT NULL;

ALTER TABLE blob DROP CONSTRAINT blob_digest_key;
ALTER TABLE blob ADD CONSTRAINT blob_digest_key UNIQUE (digest, instance);
