CREATE TABLE Countries (
    country_id INT PRIMARY KEY,
    country_name VARCHAR(100),
    default_language VARCHAR(50),
    currency VARCHAR(20),
    timezone VARCHAR(50),
);

CREATE TABLE Users (
    user_id INT PRIMARY KEY,
    username VARCHAR(50),
    email VARCHAR(100),
    country_id INT,
    preferred_language VARCHAR(50),
    FOREIGN KEY (country_id) REFERENCES Countries(country_id)
);

CREATE TABLE Features (
    feature_id INT PRIMARY KEY,
    feature_name VARCHAR(100),
    description TEXT,
    is_active BOOLEAN
);

CREATE TABLE FeatureCountries (
    feature_id INT,
    country_id INT,
    PRIMARY KEY (feature_id, country_id),
    FOREIGN KEY (feature_id) REFERENCES Features(feature_id),
    FOREIGN KEY (country_id) REFERENCES Countries(country_id)
);
