CREATE TABLE dedicated_servers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    hostname VARCHAR(255),
    ip VARCHAR(45),
    name VARCHAR(255),
    status ENUM('active', 'inactive', 'maintenance', 'decommissioned') DEFAULT 'active',
    country VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    provider VARCHAR(255),
    processor VARCHAR(255),
    ram VARCHAR(255),
    os VARCHAR(255),
    storage VARCHAR(255),
    used_for ENUM('game_server', 'virtualization', 'voice_server', 'all_purpose') DEFAULT 'all_purpose',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE server_stats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    server_id INT NOT NULL,
    server_running_state ENUM('low', 'medium', 'high', 'critical') DEFAULT 'low',
    cpu_usage DECIMAL(5,2),
    ram_usage DECIMAL(5,2),
    disk_usage DECIMAL(5,2),
    network_in BIGINT UNSIGNED,
    network_out BIGINT UNSIGNED,
    collected_at DATETIME NOT NULL,
    FOREIGN KEY (server_id) REFERENCES dedicated_servers(id) ON DELETE CASCADE,
    INDEX idx_server_time (server_id, collected_at)
);