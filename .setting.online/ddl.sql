# 2026-02-12 03:56:25
ALTER TABLE models ADD COLUMN context_length INT DEFAULT 128 COMMENT '上下文长度，单位k' AFTER prompt;
# 2026-02-12 16:03:01
ALTER TABLE models
ADD COLUMN compress_enabled TINYINT(1) DEFAULT 0 COMMENT '是否启用token压缩' AFTER context_length,
ADD COLUMN compress_truncate_len INT DEFAULT 500 COMMENT '截断过长消息的长度阈值' AFTER compress_enabled,
ADD COLUMN compress_user_count INT DEFAULT 3 COMMENT '压缩的user消息倒数数量' AFTER compress_truncate_len,
ADD COLUMN compress_role_types VARCHAR(128) DEFAULT '' COMMENT '角色类型，多个用逗号分开' AFTER compress_user_count;
# 2026-02-12 16:34:48
ALTER TABLE models DROP COLUMN prompt;
# 2026-02-12 16:34:48
ALTER TABLE providers DROP COLUMN prompt;
