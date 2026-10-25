-- card_number_hash was an unkeyed SHA-256 of the full PAN, which together with the stored last four
-- digits can be brute-forced (PCI DSS 3.5.1.1). New cards use HMAC-SHA256 with security.card_hash_key;
-- the PANs behind existing hashes were never kept, so those hashes are replaced with an opaque value.
UPDATE cards SET card_number_hash = 'legacy:' || id::text WHERE card_number_hash NOT LIKE 'legacy:%';
