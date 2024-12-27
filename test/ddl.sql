create database IF NOT EXISTS test;

USE test;

CREATE TABLE IF NOT EXISTS `Source` (
  `id` bigint unsigned NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT '',
  `type` tinyint NOT NULL DEFAULT 0,
  `description` text NOT NULL,
  `is_deleted` boolean NOT NULL DEFAULT false,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
