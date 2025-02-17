--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `partnered_banks` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `bank_name` VARCHAR(128) NOT NULL,
    `default_picture_path` LONGTEXT NOT NULL DEFAULT '',
    `partnership_status` BOOLEAN NOT NULL DEFAULT FALSE,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    UNIQUE (`bank_name`)
) ENGINE = InnoDB;
--rollback DROP TABLE `partnered_banks`;