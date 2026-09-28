CREATE TABLE IF NOT EXISTS medicines (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    ingredients TEXT NOT NULL,
    allergen_groups TEXT,
    dosage_form VARCHAR(64) NOT NULL DEFAULT '',
    strength VARCHAR(64) NOT NULL DEFAULT '',
    unit VARCHAR(32) NOT NULL,
    requires_prescription BOOLEAN NOT NULL DEFAULT FALSE,
    reorder_level INT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_medicines_code (code),
    KEY idx_medicines_name (name)
);

CREATE TABLE IF NOT EXISTS batches (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    medicine_id BIGINT NOT NULL,
    batch_number VARCHAR(64) NOT NULL,
    expiry_date DATE NOT NULL,
    quantity INT NOT NULL,
    received_quantity INT NOT NULL,
    unit_cost BIGINT NOT NULL DEFAULT 0,
    supplier VARCHAR(255) NOT NULL DEFAULT '',
    received_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_batches_medicine_number (medicine_id, batch_number),
    KEY idx_batches_medicine_expiry (medicine_id, expiry_date),
    KEY idx_batches_expiry (expiry_date),
    CONSTRAINT chk_batches_quantity CHECK (quantity >= 0)
);

CREATE TABLE IF NOT EXISTS stock_movements (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    medicine_id BIGINT NOT NULL,
    batch_id BIGINT NOT NULL,
    type VARCHAR(16) NOT NULL,
    quantity INT NOT NULL,
    reason VARCHAR(255) NOT NULL DEFAULT '',
    reference VARCHAR(255) NOT NULL DEFAULT '',
    created_by VARCHAR(128) NOT NULL DEFAULT '',
    created_at DATETIME(3) NULL,
    KEY idx_stock_movements_medicine (medicine_id, id)
);

CREATE TABLE IF NOT EXISTS interaction_rules (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ingredient_a VARCHAR(128) NOT NULL,
    ingredient_b VARCHAR(128) NOT NULL,
    severity VARCHAR(32) NOT NULL,
    description TEXT,
    source VARCHAR(16) NOT NULL DEFAULT 'local',
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_interaction_rules_pair (ingredient_a, ingredient_b)
);

CREATE TABLE IF NOT EXISTS ingredient_aliases (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    alias VARCHAR(128) NOT NULL,
    canonical VARCHAR(128) NOT NULL,
    source VARCHAR(16) NOT NULL DEFAULT 'local',
    created_at DATETIME(3) NULL,
    UNIQUE KEY uk_ingredient_aliases_alias (alias)
);

CREATE TABLE IF NOT EXISTS ingredient_allergen_groups (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ingredient VARCHAR(128) NOT NULL,
    allergen_group VARCHAR(128) NOT NULL,
    source VARCHAR(16) NOT NULL DEFAULT 'local',
    created_at DATETIME(3) NULL,
    UNIQUE KEY uk_ingredient_allergen_groups (ingredient, allergen_group)
);

CREATE TABLE IF NOT EXISTS dispenses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    patient_ref VARCHAR(128) NOT NULL,
    prescription_ref VARCHAR(128) NOT NULL DEFAULT '',
    dispensed_by VARCHAR(128) NOT NULL,
    override_reason TEXT,
    warnings TEXT,
    created_at DATETIME(3) NULL,
    KEY idx_dispenses_patient (patient_ref, id)
);

CREATE TABLE IF NOT EXISTS dispense_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    dispense_id BIGINT NOT NULL,
    medicine_id BIGINT NOT NULL,
    batch_id BIGINT NOT NULL,
    batch_number VARCHAR(64) NOT NULL,
    expiry_date DATE NOT NULL,
    quantity INT NOT NULL,
    KEY idx_dispense_items_dispense (dispense_id)
);

-- Reference data: seeded with INSERT IGNORE so rows a pharmacist has added or
-- changed through the API are never overwritten. Review before clinical use.

INSERT IGNORE INTO ingredient_aliases (alias, canonical, source, created_at) VALUES
    ('aspirin', 'acetylsalicylic acid', 'reference', NOW(3)),
    ('acid acetylsalicylic', 'acetylsalicylic acid', 'reference', NOW(3)),
    ('acetaminophen', 'paracetamol', 'reference', NOW(3)),
    ('nitroglycerin', 'glyceryl trinitrate', 'reference', NOW(3)),
    ('albuterol', 'salbutamol', 'reference', NOW(3)),
    ('adrenaline', 'epinephrine', 'reference', NOW(3)),
    ('adrenalin', 'epinephrine', 'reference', NOW(3)),
    ('frusemide', 'furosemide', 'reference', NOW(3)),
    ('lignocaine', 'lidocaine', 'reference', NOW(3)),
    ('amoxycillin', 'amoxicillin', 'reference', NOW(3)),
    ('cephalexin', 'cefalexin', 'reference', NOW(3)),
    ('glyburide', 'glibenclamide', 'reference', NOW(3)),
    ('dipyrone', 'metamizole', 'reference', NOW(3));

INSERT IGNORE INTO ingredient_allergen_groups (ingredient, allergen_group, source, created_at) VALUES
    ('amoxicillin', 'penicillin', 'reference', NOW(3)),
    ('amoxicillin', 'beta-lactam', 'reference', NOW(3)),
    ('ampicillin', 'penicillin', 'reference', NOW(3)),
    ('ampicillin', 'beta-lactam', 'reference', NOW(3)),
    ('benzylpenicillin', 'penicillin', 'reference', NOW(3)),
    ('benzylpenicillin', 'beta-lactam', 'reference', NOW(3)),
    ('phenoxymethylpenicillin', 'penicillin', 'reference', NOW(3)),
    ('phenoxymethylpenicillin', 'beta-lactam', 'reference', NOW(3)),
    ('cloxacillin', 'penicillin', 'reference', NOW(3)),
    ('cloxacillin', 'beta-lactam', 'reference', NOW(3)),
    ('piperacillin', 'penicillin', 'reference', NOW(3)),
    ('piperacillin', 'beta-lactam', 'reference', NOW(3)),
    ('cefalexin', 'cephalosporin', 'reference', NOW(3)),
    ('cefalexin', 'beta-lactam', 'reference', NOW(3)),
    ('cefadroxil', 'cephalosporin', 'reference', NOW(3)),
    ('cefadroxil', 'beta-lactam', 'reference', NOW(3)),
    ('cefaclor', 'cephalosporin', 'reference', NOW(3)),
    ('cefaclor', 'beta-lactam', 'reference', NOW(3)),
    ('cefuroxime', 'cephalosporin', 'reference', NOW(3)),
    ('cefuroxime', 'beta-lactam', 'reference', NOW(3)),
    ('cefixime', 'cephalosporin', 'reference', NOW(3)),
    ('cefixime', 'beta-lactam', 'reference', NOW(3)),
    ('cefpodoxime', 'cephalosporin', 'reference', NOW(3)),
    ('cefpodoxime', 'beta-lactam', 'reference', NOW(3)),
    ('ceftriaxone', 'cephalosporin', 'reference', NOW(3)),
    ('ceftriaxone', 'beta-lactam', 'reference', NOW(3)),
    ('cefotaxime', 'cephalosporin', 'reference', NOW(3)),
    ('cefotaxime', 'beta-lactam', 'reference', NOW(3)),
    ('ceftazidime', 'cephalosporin', 'reference', NOW(3)),
    ('ceftazidime', 'beta-lactam', 'reference', NOW(3)),
    ('meropenem', 'carbapenem', 'reference', NOW(3)),
    ('meropenem', 'beta-lactam', 'reference', NOW(3)),
    ('imipenem', 'carbapenem', 'reference', NOW(3)),
    ('imipenem', 'beta-lactam', 'reference', NOW(3)),
    ('sulfamethoxazole', 'sulfonamide', 'reference', NOW(3)),
    ('sulfadiazine', 'sulfonamide', 'reference', NOW(3)),
    ('ibuprofen', 'nsaid', 'reference', NOW(3)),
    ('diclofenac', 'nsaid', 'reference', NOW(3)),
    ('naproxen', 'nsaid', 'reference', NOW(3)),
    ('meloxicam', 'nsaid', 'reference', NOW(3)),
    ('piroxicam', 'nsaid', 'reference', NOW(3)),
    ('ketoprofen', 'nsaid', 'reference', NOW(3)),
    ('celecoxib', 'nsaid', 'reference', NOW(3)),
    ('acetylsalicylic acid', 'nsaid', 'reference', NOW(3)),
    ('azithromycin', 'macrolide', 'reference', NOW(3)),
    ('clarithromycin', 'macrolide', 'reference', NOW(3)),
    ('erythromycin', 'macrolide', 'reference', NOW(3)),
    ('ciprofloxacin', 'fluoroquinolone', 'reference', NOW(3)),
    ('levofloxacin', 'fluoroquinolone', 'reference', NOW(3)),
    ('ofloxacin', 'fluoroquinolone', 'reference', NOW(3)),
    ('moxifloxacin', 'fluoroquinolone', 'reference', NOW(3)),
    ('doxycycline', 'tetracycline', 'reference', NOW(3)),
    ('tetracycline', 'tetracycline', 'reference', NOW(3)),
    ('minocycline', 'tetracycline', 'reference', NOW(3)),
    ('codeine', 'opioid', 'reference', NOW(3)),
    ('morphine', 'opioid', 'reference', NOW(3)),
    ('tramadol', 'opioid', 'reference', NOW(3)),
    ('fentanyl', 'opioid', 'reference', NOW(3)),
    ('pethidine', 'opioid', 'reference', NOW(3));

INSERT IGNORE INTO interaction_rules (ingredient_a, ingredient_b, severity, description, source, created_at, updated_at) VALUES
    ('acetylsalicylic acid', 'warfarin', 'MAJOR', 'Tăng nguy cơ chảy máu', 'reference', NOW(3), NOW(3)),
    ('ibuprofen', 'warfarin', 'MAJOR', 'Tăng nguy cơ chảy máu', 'reference', NOW(3), NOW(3)),
    ('diclofenac', 'warfarin', 'MAJOR', 'Tăng nguy cơ chảy máu', 'reference', NOW(3), NOW(3)),
    ('naproxen', 'warfarin', 'MAJOR', 'Tăng nguy cơ chảy máu', 'reference', NOW(3), NOW(3)),
    ('fluconazole', 'warfarin', 'MAJOR', 'Ức chế chuyển hóa warfarin, tăng INR và nguy cơ chảy máu', 'reference', NOW(3), NOW(3)),
    ('metronidazole', 'warfarin', 'MAJOR', 'Ức chế chuyển hóa warfarin, tăng INR và nguy cơ chảy máu', 'reference', NOW(3), NOW(3)),
    ('clarithromycin', 'simvastatin', 'CONTRAINDICATED', 'Tăng mạnh nồng độ simvastatin, nguy cơ tiêu cơ vân', 'reference', NOW(3), NOW(3)),
    ('itraconazole', 'simvastatin', 'CONTRAINDICATED', 'Tăng mạnh nồng độ simvastatin, nguy cơ tiêu cơ vân', 'reference', NOW(3), NOW(3)),
    ('ketoconazole', 'simvastatin', 'CONTRAINDICATED', 'Tăng mạnh nồng độ simvastatin, nguy cơ tiêu cơ vân', 'reference', NOW(3), NOW(3)),
    ('glyceryl trinitrate', 'sildenafil', 'CONTRAINDICATED', 'Tụt huyết áp nặng, có thể đe dọa tính mạng', 'reference', NOW(3), NOW(3)),
    ('isosorbide mononitrate', 'sildenafil', 'CONTRAINDICATED', 'Tụt huyết áp nặng, có thể đe dọa tính mạng', 'reference', NOW(3), NOW(3)),
    ('isosorbide dinitrate', 'sildenafil', 'CONTRAINDICATED', 'Tụt huyết áp nặng, có thể đe dọa tính mạng', 'reference', NOW(3), NOW(3)),
    ('glyceryl trinitrate', 'tadalafil', 'CONTRAINDICATED', 'Tụt huyết áp nặng, có thể đe dọa tính mạng', 'reference', NOW(3), NOW(3)),
    ('isosorbide mononitrate', 'tadalafil', 'CONTRAINDICATED', 'Tụt huyết áp nặng, có thể đe dọa tính mạng', 'reference', NOW(3), NOW(3)),
    ('isosorbide dinitrate', 'tadalafil', 'CONTRAINDICATED', 'Tụt huyết áp nặng, có thể đe dọa tính mạng', 'reference', NOW(3), NOW(3)),
    ('methotrexate', 'trimethoprim', 'MAJOR', 'Tăng độc tính ức chế tủy xương của methotrexat', 'reference', NOW(3), NOW(3)),
    ('potassium chloride', 'spironolactone', 'MAJOR', 'Nguy cơ tăng kali máu', 'reference', NOW(3), NOW(3)),
    ('clopidogrel', 'omeprazole', 'MODERATE', 'Giảm hoạt hóa clopidogrel, giảm hiệu quả chống kết tập tiểu cầu', 'reference', NOW(3), NOW(3)),
    ('sertraline', 'tramadol', 'MAJOR', 'Nguy cơ hội chứng serotonin và co giật', 'reference', NOW(3), NOW(3)),
    ('fluoxetine', 'tramadol', 'MAJOR', 'Nguy cơ hội chứng serotonin và co giật', 'reference', NOW(3), NOW(3)),
    ('allopurinol', 'azathioprine', 'MAJOR', 'Tăng độc tính tủy xương của azathioprin, cần giảm liều', 'reference', NOW(3), NOW(3)),
    ('amiodarone', 'digoxin', 'MAJOR', 'Tăng nồng độ digoxin, nguy cơ ngộ độc digoxin', 'reference', NOW(3), NOW(3)),
    ('ciprofloxacin', 'tizanidine', 'CONTRAINDICATED', 'Tăng mạnh nồng độ tizanidin, tụt huyết áp và an thần quá mức', 'reference', NOW(3), NOW(3)),
    ('ciprofloxacin', 'theophylline', 'MAJOR', 'Tăng nồng độ theophylin, nguy cơ co giật và loạn nhịp', 'reference', NOW(3), NOW(3)),
    ('clarithromycin', 'colchicine', 'MAJOR', 'Tăng nồng độ colchicin, nguy cơ ngộ độc', 'reference', NOW(3), NOW(3)),
    ('acetylsalicylic acid', 'ibuprofen', 'MODERATE', 'Giảm tác dụng chống kết tập tiểu cầu của aspirin, tăng nguy cơ xuất huyết tiêu hóa', 'reference', NOW(3), NOW(3));
