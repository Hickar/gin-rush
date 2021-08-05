CREATE DATABASE IF NOT EXISTS `gin-rush`;

USE `gin-rush`;

CREATE TABLE IF NOT EXISTS `user` (
    `id` int AUTO_INCREMENT NOT NULL,
    `name` varchar(128) NOT NULL,
    `email` varchar(128) NOT NULL,
    `password` varchar(64) NOT NULL,
    `bio` varchar(512),
    `avatar` varchar(128),
    `birth_date` date,
    PRIMARY KEY (`id`)
);