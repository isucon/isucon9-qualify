DROP DATABASE IF EXISTS `isucari`;
CREATE DATABASE `isucari`;

DROP USER 'isucari'@'localhost';
CREATE USER 'isucari'@'localhost' IDENTIFIED BY 'isucari';
GRANT ALL PRIVILEGES ON `isucari`.* TO 'isucari'@'localhost';

DROP USER 'isucari'@'%';
CREATE USER 'isucari'@'%' IDENTIFIED BY 'isucari';
GRANT ALL PRIVILEGES ON `isucari`.* TO 'isucari'@'%';
