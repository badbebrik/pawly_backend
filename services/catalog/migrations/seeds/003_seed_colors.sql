BEGIN;

WITH src(name_ru, name_en, hex) AS (
    VALUES
        ('Белый', 'white', '#FFFFFF'),
        ('Черный', 'black', '#000000'),
        ('Серый', 'gray', '#808080'),
        ('Коричневый', 'brown', '#8B4513'),
        ('Рыжий', 'ginger', '#D2691E'),
        ('Кремовый', 'cream', '#FFFDD0'),
        ('Бежевый', 'beige', '#F5F5DC'),
        ('Золотой', 'golden', '#D4AF37'),
        ('Шоколадный', 'chocolate', '#7B3F00'),
        ('Голубой', 'blue', '#6FA8DC'),
        ('Серебристый', 'silver', '#C0C0C0'),
        ('Красный', 'red', '#C0392B')
),
to_insert AS (
    SELECT s.name_ru, s.name_en, s.hex
    FROM src s
    WHERE NOT EXISTS (
        SELECT 1
        FROM colors c
        WHERE lower(c.name_en) = lower(s.name_en)
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
INSERT INTO colors (name_ru, name_en, hex, is_active, version, created_at, updated_at)
SELECT
    t.name_ru,
    t.name_en,
    t.hex,
    TRUE,
    v.version,
    NOW(),
    NOW()
FROM to_insert t
CROSS JOIN v;

COMMIT;
