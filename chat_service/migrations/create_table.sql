CREATE TABLE IF NOT EXISTS `messages` (
  `id` INT NOT NULL,
  `guid` VARCHAR(100) NOT NULL,
  `sender_id` INT NOT NULL,
  `message_type` ENUM('text', 'image', 'vedio', 'audio') NOT NULL,
  `message` VARCHAR(255) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL,
  `deleted_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`));

CREATE TABLE IF NOT EXISTS `participants` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `conversation_id` INT NOT NULL,
  `users_id` INT NOT NULL,
  `type` ENUM('single', 'group') NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`));

CREATE TABLE IF NOT EXISTS `reports` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `users_id` INT NOT NULL,
  `participants_id` INT NOT NULL,
  `report_type` VARCHAR(45) NOT NULL,
  `notes` TEXT NOT NULL,
  `status` ENUM('pending', 'resolved') NOT NULL DEFAULT 'pending',
  `created_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`));

CREATE TABLE IF NOT EXISTS `block_list` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `users_id` INT NOT NULL,
  `participants_id` INT NOT NULL,
  `created_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`));

CREATE TABLE IF NOT EXISTS `deleted_messages` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `messages_id` INT NOT NULL,
  `users_id` INT NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`))

CREATE TABLE IF NOT EXISTS `conversation` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `title` VARCHAR(40) NOT NULL,
  `creator_id` INT NOT NULL,
  `channel_id` VARCHAR(45) NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`))