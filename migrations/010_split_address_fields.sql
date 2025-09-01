-- +goose Up
-- Split address field into separate address components
ALTER TABLE client ADD COLUMN address1 TEXT;
ALTER TABLE client ADD COLUMN address2 TEXT;
ALTER TABLE client ADD COLUMN address3 TEXT;
ALTER TABLE client ADD COLUMN city TEXT;
ALTER TABLE client ADD COLUMN state TEXT;
ALTER TABLE client ADD COLUMN zip_code TEXT;

-- Migrate existing address data to address1 (basic migration)
UPDATE client SET address1 = address WHERE address IS NOT NULL AND address != '';

-- Drop the old address column
ALTER TABLE client DROP COLUMN address;

-- +goose Down
-- Recreate the single address column
ALTER TABLE client ADD COLUMN address TEXT;

-- Migrate data back (combine all address fields)
UPDATE client SET address = 
    CASE 
        WHEN address1 IS NOT NULL OR address2 IS NOT NULL OR address3 IS NOT NULL OR city IS NOT NULL OR state IS NOT NULL OR zip_code IS NOT NULL
        THEN TRIM(
            COALESCE(address1, '') || 
            CASE WHEN address2 IS NOT NULL AND address2 != '' THEN ', ' || address2 ELSE '' END ||
            CASE WHEN address3 IS NOT NULL AND address3 != '' THEN ', ' || address3 ELSE '' END ||
            CASE WHEN city IS NOT NULL AND city != '' THEN ', ' || city ELSE '' END ||
            CASE WHEN state IS NOT NULL AND state != '' THEN ', ' || state ELSE '' END ||
            CASE WHEN zip_code IS NOT NULL AND zip_code != '' THEN ' ' || zip_code ELSE '' END
        )
        ELSE NULL
    END
WHERE address1 IS NOT NULL OR address2 IS NOT NULL OR address3 IS NOT NULL OR city IS NOT NULL OR state IS NOT NULL OR zip_code IS NOT NULL;

-- Remove the split address columns
ALTER TABLE client DROP COLUMN zip_code;
ALTER TABLE client DROP COLUMN state;
ALTER TABLE client DROP COLUMN city;
ALTER TABLE client DROP COLUMN address3;
ALTER TABLE client DROP COLUMN address2;
ALTER TABLE client DROP COLUMN address1;