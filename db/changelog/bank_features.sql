--liquibase formatted sql

--changeset Voxtmault:1
CREATE TABLE IF NOT EXISTS `bank_features` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(64) NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL DEFAULT NULL,
    UNIQUE(`name`)
) ENGINE = InnoDB;
--rollback DROP TABLE `bank_features`;

--changeset Voxtmault:2
INSERT INTO `bank_features` (`name`) VALUES ('Open Auth');
INSERT INTO `bank_features` (`name`) VALUES ('Bill Presentmnt');
INSERT INTO `bank_features` (`name`) VALUES ('Payment Flag');
INSERT INTO `bank_features` (`name`) VALUES ('Payment Status');
INSERT INTO `bank_features` (`name`) VALUES ('Account Balance');
INSERT INTO `bank_features` (`name`) VALUES ('External Account Inquiry');
INSERT INTO `bank_features` (`name`) VALUES ('Internal Account Inquiry');
INSERT INTO `bank_features` (`name`) VALUES ('Intrabank Transfer');
INSERT INTO `bank_features` (`name`) VALUES ('Interbank Transfer');
INSERT INTO `bank_features` (`name`) VALUES ('Bank Statement');
INSERT INTO `bank_features` (`name`) VALUES ('Transaction Status Inquiry');