/* registry support proxy hosts and vpc_id */
BEGIN;
   SET lock_timeout = '1s';
   ALTER TABLE registry ADD column if not exists http_proxy text;
   ALTER TABLE registry ADD column if not exists https_proxy text;
   ALTER TABLE registry ADD column if not exists no_proxy text;
   ALTER TABLE registry ADD column if not exists vpc_id text;
   ALTER TABLE registry ALTER column access_secret TYPE text;
END;
