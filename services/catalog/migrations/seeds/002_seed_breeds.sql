BEGIN;

WITH src(species_en, name_ru, name_en) AS (
    VALUES
        ('cat', 'Домашняя короткошерстная', 'domestic shorthair'),
        ('cat', 'Британская короткошерстная', 'british shorthair'),
        ('cat', 'Шотландская вислоухая', 'scottish fold'),
        ('cat', 'Мейн-кун', 'maine coon'),
        ('cat', 'Сиамская', 'siamese'),
        ('cat', 'Сибирская', 'siberian'),
        ('cat', 'Сфинкс', 'sphynx'),
        ('dog', 'Метис', 'mixed breed'),
        ('dog', 'Лабрадор-ретривер', 'labrador retriever'),
        ('dog', 'Немецкая овчарка', 'german shepherd'),
        ('dog', 'Золотистый ретривер', 'golden retriever'),
        ('dog', 'Французский бульдог', 'french bulldog'),
        ('dog', 'Корги', 'corgi'),
        ('dog', 'Пудель', 'poodle'),
        ('rabbit', 'Карликовый кролик', 'dwarf rabbit'),
        ('rabbit', 'Рекс', 'rex rabbit'),
        ('bird', 'Волнистый попугай', 'budgerigar'),
        ('bird', 'Канарейка', 'canary'),
        ('rodent', 'Хомяк сирийский', 'syrian hamster'),
        ('rodent', 'Морская свинка', 'guinea pig')
),
resolved AS (
    SELECT
        sp.id AS species_id,
        s.name_ru,
        s.name_en
    FROM src s
    JOIN species sp ON lower(sp.name_en) = lower(s.species_en)
),
to_insert AS (
    SELECT r.species_id, r.name_ru, r.name_en
    FROM resolved r
    WHERE NOT EXISTS (
        SELECT 1
        FROM breeds b
        WHERE b.species_id = r.species_id
          AND lower(b.name_en) = lower(r.name_en)
    )
),
bump AS (
    UPDATE catalog_version
    SET version = version + 1
    WHERE id = 1
      AND EXISTS (SELECT 1 FROM to_insert)
    RETURNING version
),
v AS (
    SELECT COALESCE(
        (SELECT version FROM bump),
        (SELECT version FROM catalog_version WHERE id = 1)
    ) AS version
)
INSERT INTO breeds (species_id, name_ru, name_en, is_active, version, created_at, updated_at)
SELECT
    t.species_id,
    t.name_ru,
    t.name_en,
    TRUE,
    v.version,
    NOW(),
    NOW()
FROM to_insert t
CROSS JOIN v;

COMMIT;
