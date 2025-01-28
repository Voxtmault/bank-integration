--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `bank_egress` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `id_bank` INT NOT NULL,
    `host_ip` VARCHAR(64) NOT NULL,
    `latency` DOUBLE UNSIGNED NOT NULL DEFAULT 0,
    `http_method` VARCHAR(16) NOT NULL DEFAULT 'GET',
    `protocol` VARCHAR(16) NOT NULL DEFAULT 'HTTP/1.1',
    `uri` LONGTEXT NOT NULL DEFAULT '',
    `response_code` INT NOT NULL DEFAULT 200,
    `response_message` LONGTEXT NOT NULL DEFAULT 'Success',
    `response_content` LONGTEXT NOT NULL DEFAULT '{}',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `FK1_BankEgress_AuthenticatedBank` FOREIGN KEY (`id_bank`) REFERENCES authenticated_banks(`id`)
)ENGINE = InnoDB;
--rollback DROP TABLE `bank_egress`;

--changeset Voxtmault:2
ALTER TABLE `bank_egress` DROP FOREIGN KEY `FK1_BankEgress_AuthenticatedBank`;
ALTER TABLE `bank_egress` DROP COLUMN `id_bank`;

--changeset Voxtmault:3
ALTER TABLE `bank_egress`
ADD COLUMN `request_parameter` JSON NOT NULL DEFAULT '{}',
ADD COLUMN `request_body` JSON NOT NULL DEFAULT '{}';
--rollback ALTER TABLE `bank_egress` DROP COLUMN `request_parameter`, DROP COLUMN `body`;

--changeset Voxtmault:4
ALTER TABLE `bank_egress`
MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP AFTER `request_body`,
ADD COLUMN `response_header` JSON NOT NULL DEFAULT '{}' AFTER `uri`,
ADD COLUMN `request_header` JSON NOT NULL DEFAULT '{}' AFTER `response_content`;