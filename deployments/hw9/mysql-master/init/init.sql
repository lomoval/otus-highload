CREATE USER 'user'@'%' IDENTIFIED BY 'pass';
GRANT ALL PRIVILEGES ON *.* TO 'user'@'%' WITH GRANT OPTION;
ALTER USER 'user'@'%' IDENTIFIED WITH mysql_native_password BY 'pass';

#CREATE USER 'root'@'%' IDENTIFIED BY 'pass';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%'  WITH GRANT OPTION;
ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY 'pass';
FLUSH PRIVILEGES;

CREATE DATABASE db;