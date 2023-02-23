CREATE TABLE IF NOT EXISTS games (
                       id INT PRIMARY KEY,
                       `name` VARCHAR,
                       description TEXT,
                       genre varchar,
                       dateOfCreate timestamp,
                       dateOfUpdate timestamp,
                       dateOfDelete timestamp,
                       `version` float,
                       episode INT,
                       `size` float,
                       price float
);