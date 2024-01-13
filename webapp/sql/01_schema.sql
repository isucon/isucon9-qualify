use `isucari`;

DROP TABLE IF EXISTS `configs`;

CREATE TABLE configs (
  `name` VARCHAR(191) NOT NULL PRIMARY KEY,
  `val` VARCHAR(255) NOT NULL
) ENGINE = InnoDB DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `account_name` varchar(128) NOT NULL UNIQUE,
  `hashed_password` varbinary(191) NOT NULL,
  `address` varchar(191) NOT NULL,
  `num_sell_items` int unsigned NOT NULL DEFAULT 0,
  `last_bump` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `items`;

CREATE TABLE `items` (
  `id` bigint NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `seller_id` bigint NOT NULL,
  `buyer_id` bigint NOT NULL DEFAULT 0,
  `status` enum(
    'on_sale',
    'trading',
    'sold_out',
    'stop',
    'cancel'
  ) NOT NULL,
  `name` varchar(191) NOT NULL,
  `price` int unsigned NOT NULL,
  `description` text NOT NULL,
  `image_name` varchar(191) NOT NULL,
  `category_id` int unsigned NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_category_id (`category_id`)
) ENGINE = InnoDB DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `transaction_evidences`;

CREATE TABLE `transaction_evidences` (
  `id` bigint NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `seller_id` bigint NOT NULL,
  `buyer_id` bigint NOT NULL,
  `status` enum('wait_shipping', 'wait_done', 'done') NOT NULL,
  `item_id` bigint NOT NULL UNIQUE,
  `item_name` varchar(191) NOT NULL,
  `item_price` int unsigned NOT NULL,
  `item_description` text NOT NULL,
  `item_category_id` int unsigned NOT NULL,
  `item_root_category_id` int unsigned NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `shippings`;

CREATE TABLE `shippings` (
  `transaction_evidence_id` bigint NOT NULL PRIMARY KEY,
  `status` enum('initial', 'wait_pickup', 'shipping', 'done') NOT NULL,
  `item_name` varchar(191) NOT NULL,
  `item_id` bigint NOT NULL,
  `reserve_id` varchar(191) NOT NULL,
  `reserve_time` bigint NOT NULL,
  `to_address` varchar(191) NOT NULL,
  `to_name` varchar(191) NOT NULL,
  `from_address` varchar(191) NOT NULL,
  `from_name` varchar(191) NOT NULL,
  `img_binary` mediumblob NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `categories`;

CREATE TABLE `categories` (
  `id` int unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `parent_id` int unsigned NOT NULL,
  `category_name` varchar(191) NOT NULL
) ENGINE = InnoDB DEFAULT CHARACTER SET utf8mb4;
