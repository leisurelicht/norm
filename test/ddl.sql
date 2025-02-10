create database IF NOT EXISTS test;

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
