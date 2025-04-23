CREATE DATABASE IF NOT EXISTS test
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_0900_ai_ci;

USE test;

CREATE TABLE IF NOT EXISTS `source` (
  `id` bigint unsigned NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT '',
  `type` tinyint NOT NULL DEFAULT 0,
  `description` text NOT NULL,
  `is_deleted` boolean NOT NULL DEFAULT false,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (11, 'Acfun', 1, 'A 站', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (12, 'Acfun', 2, 'A 站', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (13, 'Acfun', 3, 'A 站', 0, '2024-03-19 15:16:24', '2024-03-19 15:16:24');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (21, 'Bilibili', 1, 'B 站', 0, '2024-03-19 15:16:24', '2024-03-19 15:16:24');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (22, 'Bilibili', 2, 'B 站', 0, '2024-03-19 15:16:25', '2024-03-19 15:16:25');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (23, 'Bilibili', 3, 'B 站', 0, '2024-03-19 15:16:25', '2024-03-19 15:16:25');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (31, 'Apple', 1, '苹果', 1, '2024-05-16 17:33:22', '2024-05-16 17:33:22');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (32, 'Apple', 2, '苹果', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (33, 'Apple', 3, '苹果', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (41, 'Google', 1, '谷歌', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (42, 'Google', 2, '谷歌', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (43, 'Google', 3, '谷歌', 1, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (51, 'Microsoft', 1, '微软', 0, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (52, 'Microsoft', 2, '微软', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');
INSERT INTO source (id, name, type, description, is_deleted, create_time, update_time) VALUES (53, 'Microsoft', 3, '微软', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');


CREATE TABLE IF NOT EXISTS `property` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `source_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '类型 ID',
  `column_name` varchar(255) NOT NULL DEFAULT '' COMMENT '字段名',
  `show_name` varchar(255) NOT NULL DEFAULT '' COMMENT '展示字段名',
  `description` text NOT NULL COMMENT '描述',
  `is_deleted` boolean NOT NULL DEFAULT false COMMENT '是否删除',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='字段';

INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (1, 11, 'title', '标题', '标题', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (2, 11, 'description', '描述', '描述', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (3, 11, 'cover', '封面', '封面', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (4, 12, 'title', '标题', '标题', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (5, 12, 'description', '描述', '描述', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (6, 12, 'cover', '封面', '封面', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');   
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (7, 13, 'title', '标题', '标题', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (8, 13, 'description', '描述', '描述', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (9, 13, 'cover', '封面', '封面', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (10, 21, 'title', '标题', '标题', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (11, 21, 'description', '描述', '描述', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (12, 21, 'cover', '封面', '封面', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (13, 22, 'title', '标题', '标题', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (14, 22, 'description', '描述', '描述', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (15, 22, 'cover', '封面', '封面', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (16, 23, 'title', '标题', '标题', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (17, 23, 'description', '描述', '描述', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (18, 23, 'cover', '封面', '封面', 0, '2024-03-19 15:16:23', '2024-03-19 15:16:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (19, 31, 'title', '标题', '标题', 1, '2024-05-16 17:33:22', '2024-05-16 17:33:22');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (20, 31, 'description', '描述', '描述', 1, '2024-05-16 17:33:22', '2024-05-16 17:33:22');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (21, 31, 'cover', '封面', '封面', 1, '2024-05-16 17:33:22', '2024-05-16 17:33:22');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (22, 32, 'title', '标题', '标题', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (23, 32, 'description', '描述', '描述', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (24, 32, 'cover', '封面', '封面', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (25, 33, 'title', '标题', '标题', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (26, 33, 'description', '描述', '描述', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (27, 33, 'cover', '封面', '封面', 1, '2024-05-16 17:33:23', '2024-05-16 17:33:23');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (28, 41, 'title', '标题', '标题', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (29, 41, 'description', '描述', '描述', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (30, 41, 'cover', '封面', '封面', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (31, 42, 'title', '标题', '标题', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (32, 42, 'description', '描述', '描述', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (33, 42, 'cover', '封面', '封面', 1, '2024-05-16 17:33:24', '2024-05-16 17:33:24');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (34, 43, 'title', '标题', '标题', 1, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (35, 43, 'description', '描述', '描述', 1, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (36, 43, 'cover', '封面', '封面', 1, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (37, 51, 'title', '标题', '标题', 0, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (38, 51, 'description', '描述', '描述', 0, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (39, 51, 'cover', '封面', '封面', 0, '2024-05-16 17:33:25', '2024-05-16 17:33:25');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (40, 52, 'title', '标题', '标题', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (41, 52, 'description', '描述', '描述', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (42, 52, 'cover', '封面', '封面', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (43, 53, 'title', '标题', '标题', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (44, 53, 'description', '描述', '描述', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');
INSERT INTO property (id, source_id, column_name, show_name, description, is_deleted, create_time, update_time) VALUES (45, 53, 'cover', '封面', '封面', 0, '2024-05-16 17:33:26', '2024-05-16 17:33:26');