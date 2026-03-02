BEGIN;

WITH src(name_ru, name_en) AS (
    VALUES
        ('Кошка', 'cat'),
        ('Собака', 'dog'),
        ('Кролик', 'rabbit'),
        ('Птица', 'bird'),
        ('Грызун', 'rodent'),
        ('Рептилия', 'reptile'),
        ('Хорек', 'ferret'),
        ('Лошадь', 'horse')
),
to_insert AS (
    SELECT s.name_ru, s.name_en
    FROM src s
    WHERE NOT EXISTS (
        SELECT 1
        FROM species x
        WHERE lower(x.name_en) = lower(s.name_en)
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
INSERT INTO species (name_ru, name_en, is_active, version, created_at, updated_at)
SELECT
    t.name_ru,
    t.name_en,
    TRUE,
    v.version,
    NOW(),
    NOW()
FROM to_insert t
CROSS JOIN v;

COMMIT;
