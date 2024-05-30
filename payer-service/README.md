CREATE TABLE payers (
  payer_id INT PRIMARY KEY AUTO_INCREMENT,
  amount DECIMAL(10,2) NOT NULL,
  currency CHAR(3) NOT NULL,
  payer_method_id INT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  reference_id VARCHAR(255) DEFAULT NULL,
  status ENUM('pending', 'succeeded', 'failed') NOT NULL,
  FOREIGN KEY (payer_method_id) REFERENCES payerMethods(payer_method_id)
);

CREATE TABLE payerMethods (
  payer_method_id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(1000) DEFAULT NULL
);