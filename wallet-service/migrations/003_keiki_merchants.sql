-- Platform banks and merchants that keikiapp and the store services charge
-- through. Fixed ids: doordash-service (WALLET_MERCHANT_ID), basket-service
-- (wallet_service.merchant_id) and keikiapp (WALLET_ISSUER_BANK_ID) reference them.
INSERT INTO banks (id, name, code, country) VALUES
    ('7a1e0000-0000-4000-8000-000000000001', 'Keiki Issuer',   'KEIKIISS', 'VN'),
    ('7a1e0000-0000-4000-8000-000000000002', 'Keiki Acquirer', 'KEIKIACQ', 'VN')
ON CONFLICT DO NOTHING;

INSERT INTO merchants (id, name, mcc, acquirer_bank_id, country) VALUES
    ('55555555-5555-5555-5555-555555555555', 'Keiki Food', '5812', '7a1e0000-0000-4000-8000-000000000002', 'VN'),
    ('66666666-6666-6666-6666-666666666666', 'Keiki Shop', '5399', '7a1e0000-0000-4000-8000-000000000002', 'VN')
ON CONFLICT DO NOTHING;
