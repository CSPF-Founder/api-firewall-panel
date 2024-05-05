-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS `users` (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(64) NULL,
    email VARCHAR(255) NULL,
    `role_id` INTEGER
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `app_config` (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    keyword VARCHAR(64) NOT NULL,
    value TEXT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `endpoints` (
    `id` bigint(20) unsigned NOT NULL DEFAULT uuid_short(),
    `label` varchar(128) NOT NULL,
    `api_url` varchar(2048) NOT NULL,
    `listening_port` SMALLINT UNSIGNED NOT NULL,
    `health_port` SMALLINT UNSIGNED NOT NULL,
    `request_mode` varchar(128) NOT NULL DEFAULT 'monitor',
    `created_at` datetime DEFAULT NOW(),
    `user_id` bigint(20) unsigned NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE(`listening_port`),
    UNIQUE(`label`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `endpoint_configs` (
`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`endpoint_id` bigint(20) unsigned NOT NULL,
`config_key` varchar(128) NOT NULL,
`config_value` varchar(255) DEFAULT NULL,
PRIMARY KEY (`id`),
FOREIGN KEY (`endpoint_id`) REFERENCES `endpoints` (`id`) ON DELETE
CASCADE,
UNIQUE KEY (`endpoint_id`, `config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `allowed_ips` (
`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`endpoint_id` bigint(20) unsigned NOT NULL,
`ip_data` varchar(128) NOT NULL,
`is_range` tinyint(1) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`),
FOREIGN KEY (`endpoint_id`) REFERENCES `endpoints` (`id`) ON DELETE
CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `denied_tokens` (
`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`endpoint_id` bigint(20) unsigned NOT NULL,
`token` varchar(4096) NOT NULL,
PRIMARY KEY (`id`),
FOREIGN KEY (`endpoint_id`) REFERENCES `endpoints` (`id`) ON DELETE
CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS `roles` (
    `id`          INTEGER PRIMARY KEY AUTO_INCREMENT,
    `keyword`        VARCHAR(255) NOT NULL UNIQUE,
    `name`        VARCHAR(255) NOT NULL UNIQUE,
    `description` VARCHAR(255)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ALTER TABLE `users` ADD COLUMN `role_id` INTEGER;

CREATE TABLE IF NOT EXISTS `permissions` (
    `id`          INTEGER PRIMARY KEY AUTO_INCREMENT,
    `keyword`        VARCHAR(255) NOT NULL UNIQUE,
    `name`        VARCHAR(255) NOT NULL UNIQUE,
    `description` VARCHAR(255)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


CREATE TABLE IF NOT EXISTS `role_permissions` (
    `role_id`       INTEGER NOT NULL,
    `permission_id` INTEGER NOT NULL
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `roles` (`keyword`, `name`, `description`)
VALUES
    ("admin", "Admin", "System administrator with full permissions"),
    ("customer", "Customer", "Customer role with edit access");

INSERT INTO `permissions` (`keyword`, `name`, `description`)
VALUES
    ("view_objects", "View Objects", "View objects"),
    ("modify_objects", "Modify Objects", "Create and edit objects"),
    ("modify_system", "Modify System", "Manage system-wide configuration");

-- Our rules for generating the admin user are:
-- * The user with the name `admin`
-- * OR the first user, if no `admin` user exists
-- MySQL apparently makes these queries gross. Thanks MySQL.
UPDATE `users` SET `role_id`=(
    SELECT `id` FROM `roles` WHERE `keyword`="admin")
WHERE `id`=(
    SELECT `id` FROM (
        SELECT * FROM `users`
    ) as u WHERE `username`="admin"
    OR `id`=(
        SELECT MIN(`id`) FROM (
            SELECT * FROM `users`
        ) as u
    ) LIMIT 1);

-- Every other user will be considered a standard user account. The admin user
-- will be able to change the role of any other user at any time.
UPDATE `users` SET `role_id`=(
    SELECT `id` FROM `roles` AS role_id WHERE `keyword`="customer")
WHERE role_id IS NULL;

-- Our default permission set will:
-- * Allow admins the ability to do anything
-- * Allow users to modify objects

-- Allow any user to view objects
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id FROM roles AS r, `permissions` AS p
WHERE r.id IN (SELECT `id` FROM roles WHERE `keyword`="admin" OR `keyword`="customer")
AND p.id=(SELECT `id` FROM `permissions` WHERE `keyword`="view_objects");

-- Allow admins and users to modify objects
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id FROM roles AS r, `permissions` AS p
WHERE r.id IN (SELECT `id` FROM roles WHERE `keyword`="admin" OR `keyword`="customer")
AND p.id=(SELECT `id` FROM `permissions` WHERE `keyword`="modify_objects");

-- Allow admins to modify system level configuration
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id FROM roles AS r, `permissions` AS p
WHERE r.id IN (SELECT `id` FROM roles WHERE `keyword`="admin")
AND p.id=(SELECT `id` FROM `permissions` WHERE `keyword`="modify_system");

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE `users`;
DROP TABLE `app_config`;
DROP TABLE `endpoints`;
DROP TABLE `endpoint_configs`;
DROP TABLE `allowed_ips`;
DROP TABLE `denied_tokens`;
DROP TABLE `roles`
DROP TABLE `user_roles`
DROP TABLE `permissions`

