ALTER TABLE api_keys
  ADD COLUMN IF NOT EXISTS group_ids JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE api_keys
SET group_ids = CASE
  WHEN group_id IS NULL THEN '[]'::jsonb
  ELSE jsonb_build_array(group_id)
END
WHERE group_ids IS NULL OR group_ids = '[]'::jsonb;

CREATE INDEX IF NOT EXISTS idx_api_keys_group_ids_gin
  ON api_keys USING GIN (group_ids)
  WHERE deleted_at IS NULL;
