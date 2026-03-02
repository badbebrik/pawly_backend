BEGIN;

WITH src(name_ru, name_en, icon_key) AS (
    VALUES
        ('Сплошной', 'solid', 'solid'),
        ('Тигровый', 'tabby', 'tabby'),
        ('Биколор', 'bicolor', 'bicolor'),
        ('Триколор', 'tricolor', 'tricolor'),
        ('Черепаховый', 'tortoiseshell', 'tortoiseshell'),
        ('Пятнистый', 'spotted', 'spotted'),
        ('Полосатый', 'striped', 'striped'),
        ('Дымчатый', 'smoke', 'smoke'),
        ('Мраморный', 'marble', 'marble'),
        ('Мерль', 'merle', 'merle'),
        ('Тигрово-полосатый', 'brindle', 'brindle')
),
to_insert AS (
    SELECT s.name_ru, s.name_en, s.icon_key
    FROM src s
    WHERE NOT EXISTS (
        SELECT 1
        FROM patterns p
        WHERE lower(p.name_en) = lower(s.name_en)
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
INSERT INTO patterns (name_ru, name_en, icon_key, is_active, version, created_at, updated_at)
SELECT
    t.name_ru,
    t.name_en,
    t.icon_key,
    TRUE,
    v.version,
    NOW(),
    NOW()
FROM to_insert t
CROSS JOIN v;

COMMIT;
