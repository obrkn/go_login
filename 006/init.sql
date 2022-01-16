CREATE DATABASE IF NOT EXISTS dbname character SET utf8mb4 collate utf8mb4_bin;

USE dbname;

DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
  id INT UNSIGNED NOT NULL PRIMARY KEY auto_increment,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  last_login_at DATETIME,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) character SET utf8mb4 collate utf8mb4_bin;