CREATE DATABASE `minions` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

use minions;

#MediaLog
create table MediaLog (
Id			   BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
`Key`			VARCHAR(32) NOT NULL DEFAULT "",
MediaId           	CHAR(64) NOT NULL DEFAULT "",
Ip                     VARCHAR(24) NULL,
RefId BIGINT NOT NULL DEFAULT 0,
CreateTs 	INT NOT NULL DEFAULT 0
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 DEFAULT collate=utf8mb4_bin;

#errlog
create table ErrLog (
Id			   BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
ErrType			VARCHAR(32) NOT NULL DEFAULT "",
ErrContent           	VARCHAR(64) NOT NULL DEFAULT "",
MediaId           	CHAR(64) NOT NULL DEFAULT "",
Ip                     VARCHAR(24) NULL,
CreateTs 	INT NOT NULL DEFAULT 0
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 DEFAULT collate=utf8mb4_bin;

#acclog
create table AccLog (
Id			   BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
IP 		VARCHAR(24) NULL,
URI 		VARCHAR(128) NULL,
Method   	VARCHAR(8) NOT NULL Default "",
UA     		VARCHAR(256) NULL,
StatusCode     	INT NOT NULL DEFAULT 0,
ContentLength   INT NULL,
ResponseTs   	INT NOT NULL DEFAULT 0,
CreateTs 	INT NOT NULL DEFAULT 0
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 DEFAULT collate=utf8mb4_bin;


#minions
create table Minions (
Id			   BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
CreateTs 	INT NOT NULL DEFAULT 0
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 DEFAULT collate=utf8mb4_bin;


#manage
create table Manage(
UserId VARCHAR(32),
OpenId VARCHAR(32),
PassCode VARCHAR(32),
AccessCode VARCHAR(16),
ExpiredTs int,
Times int,
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 DEFAULT collate=utf8mb4_bin;


#config
create table Config(
`Key` varchar(12),
value int
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 DEFAULT collate=utf8mb4_bin;

