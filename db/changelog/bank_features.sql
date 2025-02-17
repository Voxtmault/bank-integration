--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `bank_features` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `id_bank` INT NOT NULL,
    `feature_name` VARCHAR(64) NOT NULL,
    `feature_note` LONGTEXT NOT NULL DEFAULT '',
    `feature_status` BOOLEAN NOT NULL DEFAULT FALSE,
    `feature_type` INT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    CONSTRAINT `FK1_BankFeature_PartneredBanks` FOREIGN KEY (`id_bank`) REFERENCES `partnered_banks`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT `FK2_BankFeature_FeatureType` FOREIGN KEY (`feature_type`) REFERENCES `bank_feature_types`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE = InnoDB;
--rollback DROP TABLE `bank_features`;