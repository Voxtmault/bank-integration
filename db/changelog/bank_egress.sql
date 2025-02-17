--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `bank_egress_logs` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `id_bank` INT NOT NULL,
    `id_feature` INT NOT NULL,
    `latency` VARCHAR(16) NOT NULL DEFAULT '',
    `response_code` INT NOT NULL DEFAULT 200,
    `host_ip` VARCHAR(64) NOT NULL,
    `http_method` VARCHAR(16) NOT NULL DEFAULT 'GET',
    `protocol` VARCHAR(16) NOT NULL DEFAULT 'HTTP/1.1',
    `uri` LONGTEXT NOT NULL DEFAULT '',
    `request_header` JSON NOT NULL DEFAULT '{}',
    `request_payload` JSON NOT NULL DEFAULT '{}',
    `response_header` JSON NOT NULL DEFAULT '{}',
    `response_payload` JSON NOT NULL DEFAULT '{}',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `FK1_BankEgressLog_AuthenticatedBank` FOREIGN KEY (`id_bank`) REFERENCES `partnered_banks`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT `FK2_BankEgressLog_BankFeature` FOREIGN KEY (`id_feature`) REFERENCES `bank_features`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE = InnoDB;
--rollback DROP TABLE `bank_egress_logs`;