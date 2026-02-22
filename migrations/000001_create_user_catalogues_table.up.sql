-- use migrate with cli: 
-- chạy cli migrate gần nhất: migrate up 1
-- example migrate up 1: migrate -path migrations -database "mysql://root:@tcp(127.0.0.1:3306)/db_golang?multiStatements=true" up 1
-- chạy cli migrate tất cả: migrate up
-- example migrate up: migrate -path migrations -database "mysql://root:@tcp(127.0.0.1:3306)/db_golang?multiStatements=true" up

-- tạo bảng user_catalogues
CREATE TABLE IF NOT EXISTS user_catalogues (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NULL,
    role VARCHAR(255) NOT NULL,
    publish TINYINT(1) DEFAULT 2 COMMENT '0: is private, 1: is draft, 2: is publish',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);