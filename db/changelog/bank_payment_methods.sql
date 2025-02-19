--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `bank_payment_methods` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `id_bank` INT NOT NULL,
    `id_payment_method` INT NOT NULL,
    `method_picture_path` LONGTEXT NOT NULL DEFAULT '',
    `method_status` BOOLEAN NOT NULL DEFAULT FALSE,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    CONSTRAINT `FK1_BankPaymentMethod_PartneredBanks` FOREIGN KEY (`id_bank`) REFERENCES `partnered_banks`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT `FK2_BankPaymentMethod_PaymentMethod` FOREIGN KEY (`id_payment_method`) REFERENCES `payment_methods`(`id`) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE = InnoDB;
--rollback DROP TABLE `bank_payment_methods`;