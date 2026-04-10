-- 创建用户表
CREATE TABLE "users" (
    "id" SERIAL PRIMARY KEY,
    "open_id" VARCHAR(128) NOT NULL UNIQUE, -- 小程序唯一标识
    "union_id" VARCHAR(128),                -- 微信全平台唯一标识
    "nickname" VARCHAR(100),                -- 昵称
    "avatar_url" TEXT,                      -- 头像
    "gender" SMALLINT DEFAULT 0,            -- 性别 0:未知 1:男 2:女
    "last_login_time" TIMESTAMPTZ,            -- 最后登录时间
    "created_time" TIMESTAMPTZ NOT NULL,
    "updated_time" TIMESTAMPTZ,
    "deleted_time" TIMESTAMPTZ                -- 软删除支持
);
-- 用户表补充字段
ALTER TABLE users 
ADD COLUMN bio VARCHAR(255) DEFAULT '',
ADD COLUMN background VARCHAR(255) DEFAULT '';
ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;


-- 为常用查询字段建立索引
CREATE INDEX "idx_users_open_id" ON "users" ("open_id");
CREATE INDEX "idx_users_union_id" ON "users" ("union_id");

-- user_post帖子表
CREATE TABLE user_post (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    content TEXT,
    title VARCHAR(255) NOT NULL DEFAULT '',
    is_featured BOOLEAN DEFAULT FALSE,
    favorite_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    image_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_post_created ON user_post(created_at DESC);
CREATE INDEX idx_post_featured ON user_post(is_featured);

-- 帖子图片
CREATE TABLE post_images (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    image_url TEXT,
    sort INT DEFAULT 0
);

--收藏
CREATE TABLE post_favorite (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    post_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, post_id)
);

--评论
CREATE TABLE post_comment (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT,
    user_id BIGINT,
    parent_id BIGINT DEFAULT 0,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

--评论图片
CREATE TABLE comment_images (
    id BIGSERIAL PRIMARY KEY,
    comment_id BIGINT,
    image_url TEXT
);

-- user_follow（关注表）
CREATE TABLE user_follow (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,        -- 谁发起关注
    follow_user_id BIGINT NOT NULL, -- 被关注的人
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, follow_user_id)
);
