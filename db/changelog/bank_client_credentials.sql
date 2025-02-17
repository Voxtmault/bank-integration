--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `bank_client_credentials` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `id_bank` INT NOT NULL,
    `client_id` LONGTEXT NOT NULL,
    `client_secret` LONGTEXT NOT NULL,
    `credential_status` BOOLEAN NOT NULL DEFAULT FALSE,
    `credential_note` LONGTEXT NOT NULL DEFAULT '',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    CONSTRAINT `FK1_BankClientCredentials_PartneredBanks` FOREIGN KEY (`id_bank`) REFERENCES `partnered_banks`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE = InnoDB;
--rollback DROP TABLE `bank_client_credentials`;