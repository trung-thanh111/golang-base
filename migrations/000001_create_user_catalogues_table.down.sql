-- use rollback table with cli: 
-- chạy cli roolback migrate gần nhất: migrate down 1
-- example migrate down 1: 
-- migrate -path migrations -database "mysql://root:@tcp(127.0.0.1:3306)/db_golang?multiStatements=true" down 1
-- chạy cli rollback tất cả: migrate down
-- example migrate down: 
-- migrate -path migrations -database "mysql://root:@tcp(127.0.0.1:3306)/db_golang?multiStatements=true" down

-- xóa bảng user_catalogues
DROP TABLE IF EXISTS user_catalogues;