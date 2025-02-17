--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `payment_methods` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `id_bank` INT NOT NULL,
    `method_name` VARCHAR(64) NOT NULL,
    `method_picture_path` LONGTEXT NOT NULL DEFAULT '',
    `method_status` BOOLEAN NOT NULL DEFAULT FALSE,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    CONSTRAINT `FK1_PaymentMethod_PartneredBanks` FOREIGN KEY (`id_bank`) REFERENCES `partnered_banks`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE = InnoDB;
--rollback DROP TABLE `payment_methods`;