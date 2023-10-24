USE orders;
CREATE TABLE orders (id varchar(255) NOT NULL, price float NOT NULL, tax float NOT NULL, final_price float NOT NULL, PRIMARY KEY (id));

INSERT INTO orders (id, price, tax, final_price) VALUES ("aaa", 10.9, 0.1, 11);