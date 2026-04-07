ALTER TABLE groups
  ADD COLUMN IF NOT EXISTS grok_input_price_per_mtok DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_output_price_per_mtok DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_image_price_1k DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_image_price_2k DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_video_price_5s DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_video_price_10s DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_video_price_15s DECIMAL(20,8),
  ADD COLUMN IF NOT EXISTS grok_video_high_quality_multiplier DECIMAL(10,4);
