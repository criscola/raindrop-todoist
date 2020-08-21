CREATE TABLE IF NOT EXISTS bookmark_with_todo (
    id_bookmark bigint PRIMARY KEY,
    id_todo bigint UNIQUE,
    insert_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);