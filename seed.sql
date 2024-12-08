--note: you need to register as a new user

--users_db:
create database users_db;
use users_db;

create TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    membership ENUM('Basic', 'Premium') NOT NULL
);


--booking_db
create DATABASE booking_db;
USE booking_db;

CREATE TABLE vehicles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_available BOOLEAN DEFAULT TRUE
);


INSERT INTO vehicles (name, is_available) VALUES
('Tesla Model 3', TRUE),
('Nissan Leaf', TRUE),
('Chevy Bolt', TRUE);


CREATE TABLE bookings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    vehicle_id INT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    price DECIMAL(10, 2),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);
