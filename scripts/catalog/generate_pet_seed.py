#!/usr/bin/env python3
import argparse
import hashlib
import json
import sys
import uuid
from pathlib import Path


PATTERNS = [
    ("f1111111-1111-1111-1111-111111111111", "Сплошной", "solid", 10),
    ("f1111111-1111-1111-1111-111111111112", "Тигровый", "tabby", 20),
    ("f1111111-1111-1111-1111-111111111113", "Биколор", "bicolor", 30),
    ("f1111111-1111-1111-1111-111111111114", "Триколор", "tricolor", 40),
    ("f1111111-1111-1111-1111-111111111115", "Черепаховый", "tortoiseshell", 50),
    ("f1111111-1111-1111-1111-111111111116", "Пятнистый", "spotted", 50),
    ("f1111111-1111-1111-1111-111111111117", "Полосатый", "striped", 60),
    ("f1111111-1111-1111-1111-111111111118", "Дымчатый", "smoke", 70),
    ("f1111111-1111-1111-1111-111111111120", "Мерль", "merle", 80),
]


COLOR_PRESETS = [
    ("c1111111-1111-1111-1111-111111111111", "Белый", "#FFFFFF", 10),
    ("c1111111-1111-1111-1111-111111111112", "Черный", "#000000", 20),
    ("c1111111-1111-1111-1111-111111111113", "Серый", "#808080", 30),
    ("c1111111-1111-1111-1111-111111111114", "Коричневый", "#8B4513", 40),
    ("c1111111-1111-1111-1111-111111111115", "Рыжий", "#D2691E", 50),
    ("c1111111-1111-1111-1111-111111111116", "Бежевый", "#F3D686", 60),
    ("c1111111-1111-1111-1111-111111111127", "Голубой", "#6FA8DC", 70),
    ("c1111111-1111-1111-1111-111111111128", "Красный", "#C0392B", 80),
    ("c1111111-1111-1111-1111-111111111129", "Зеленый", "#6AD247", 90),
]


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--catalog", default="services/pet/catalog/pet_catalog.json")
    parser.add_argument("--out", default="services/pet/migrations/0002_seed_pet_reference_data.sql")
    args = parser.parse_args()

    catalog_path = Path(args.catalog)
    out_path = Path(args.out)
    catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
    validate(catalog)
    out_path.write_text(render_migration(catalog), encoding="utf-8")
    return 0


def validate(catalog: dict) -> None:
    validate_static_reference_data()

    species_codes: set[str] = set()
    for item in catalog.get("species", []):
        code = clean_code(item.get("code", ""))
        if not code or not item.get("name") or not item.get("icon_key"):
            raise ValueError("species must have code, name and icon_key")
        if code in species_codes:
            raise ValueError(f"duplicate species code: {code}")
        species_codes.add(code)

    if not species_codes:
        raise ValueError("species list is empty")

    breed_keys: set[tuple[str, str]] = set()
    for item in catalog.get("breeds", []):
        code = clean_code(item.get("species_code", ""))
        name = clean_text(item.get("name", ""))
        if code not in species_codes:
            raise ValueError(f"breed {name!r} references unknown species_code {code!r}")
        if not name:
            raise ValueError(f"breed for {code!r} has empty name")
        key = (code, name.casefold())
        if key in breed_keys:
            raise ValueError(f"duplicate breed {name!r} for species {code!r}")
        breed_keys.add(key)


def validate_static_reference_data() -> None:
    validate_unique_static_items("PATTERNS", PATTERNS)
    validate_unique_static_items("COLOR_PRESETS", COLOR_PRESETS)


def validate_unique_static_items(name: str, items: list[tuple[str, str, str, int]]) -> None:
    ids: set[str] = set()
    names: set[str] = set()
    for item_id, item_name, *_ in items:
        if item_id in ids:
            raise ValueError(f"duplicate id in {name}: {item_id}")
        ids.add(item_id)
        normalized_name = clean_text(item_name).casefold()
        if normalized_name in names:
            raise ValueError(f"duplicate name in {name}: {item_name}")
        names.add(normalized_name)


def render_migration(catalog: dict) -> str:
    parts = [
        "-- +goose Up",
        "BEGIN;",
        "",
        render_species(catalog["species"]),
        "",
        render_breeds(catalog.get("breeds", [])),
        "",
        render_patterns(),
        "",
        render_color_presets(),
        "",
        "COMMIT;",
        "",
        "-- +goose Down",
        render_down(catalog),
    ]
    return "\n".join(parts).rstrip() + "\n"


def render_species(items: list[dict]) -> str:
    rows = []
    for item in items:
        code = clean_code(item["code"])
        rows.append(
            "    "
            + sql_tuple(
                stable_uuid(f"species:{code}"),
                code,
                clean_text(item["name"]),
                clean_text(item["icon_key"]),
                int(item["sort_order"]),
                bool(item["active"]),
                raw("NOW()"),
                raw("NOW()"),
            )
        )
    return "\n".join(
        [
            "INSERT INTO species (id, code, name, icon_key, sort_order, is_active, created_at, updated_at)",
            "VALUES",
            ",\n".join(rows),
            "ON CONFLICT (code) DO UPDATE",
            "SET name = EXCLUDED.name,",
            "    icon_key = EXCLUDED.icon_key,",
            "    sort_order = EXCLUDED.sort_order,",
            "    is_active = EXCLUDED.is_active,",
            "    updated_at = NOW();",
        ]
    )


def render_breeds(items: list[dict]) -> str:
    if not items:
        return "-- No breeds in catalog."

    rows = []
    for item in items:
        code = clean_code(item["species_code"])
        name = clean_text(item["name"])
        rows.append(
            "        "
            + sql_tuple(
                cast(stable_uuid(f"breed:{code}:{name}"), "uuid"),
                code,
                name,
                int(item["sort_order"]),
                bool(item["active"]),
            )
        )
    return "\n".join(
        [
            "WITH src(id, species_code, name, sort_order, is_active) AS (",
            "    VALUES",
            ",\n".join(rows),
            "),",
            "resolved AS (",
            "    SELECT src.id, sp.id AS species_id, src.name, src.sort_order, src.is_active",
            "    FROM src",
            "    JOIN species sp ON sp.code = src.species_code",
            ")",
            "INSERT INTO breeds (id, species_id, name, sort_order, is_active, created_at, updated_at)",
            "SELECT id, species_id, name, sort_order, is_active, NOW(), NOW()",
            "FROM resolved",
            "ON CONFLICT (species_id, lower(name)) DO UPDATE",
            "SET sort_order = EXCLUDED.sort_order,",
            "    is_active = EXCLUDED.is_active,",
            "    updated_at = NOW();",
        ]
    )


def render_patterns() -> str:
    rows = [
        "        " + sql_tuple(cast(item_id, "uuid"), name, icon_key, sort_order)
        for item_id, name, icon_key, sort_order in PATTERNS
    ]
    return "\n".join(
        [
            "WITH src(id, name, icon_key, sort_order) AS (",
            "    VALUES",
            ",\n".join(rows),
            ")",
            "INSERT INTO patterns (id, species_id, name, icon_key, sort_order, is_active, created_at, updated_at)",
            "SELECT id, NULL, name, icon_key, sort_order, TRUE, NOW(), NOW()",
            "FROM src",
            "ON CONFLICT (COALESCE(species_id, '00000000-0000-0000-0000-000000000000'::uuid), lower(name)) DO UPDATE",
            "SET icon_key = EXCLUDED.icon_key,",
            "    sort_order = EXCLUDED.sort_order,",
            "    is_active = EXCLUDED.is_active,",
            "    updated_at = NOW();",
        ]
    )


def render_color_presets() -> str:
    rows = [
        "    " + sql_tuple(cast(item_id, "uuid"), name, hex_value, sort_order, True, raw("NOW()"), raw("NOW()"))
        for item_id, name, hex_value, sort_order in COLOR_PRESETS
    ]
    return "\n".join(
        [
            "INSERT INTO color_presets (id, name, hex, sort_order, is_active, created_at, updated_at)",
            "VALUES",
            ",\n".join(rows),
            "ON CONFLICT (lower(name)) DO UPDATE",
            "SET hex = EXCLUDED.hex,",
            "    sort_order = EXCLUDED.sort_order,",
            "    is_active = EXCLUDED.is_active,",
            "    updated_at = NOW();",
        ]
    )


def render_down(catalog: dict) -> str:
    codes = sorted(clean_code(item["code"]) for item in catalog["species"])
    return "\n".join(
        [
            "DELETE FROM color_presets",
            "WHERE id IN (",
            ",\n".join(f"    {sql_value(item_id)}" for item_id, *_ in COLOR_PRESETS),
            ");",
            "",
            "DELETE FROM patterns",
            "WHERE id IN (",
            ",\n".join(f"    {sql_value(item_id)}" for item_id, *_ in PATTERNS),
            ");",
            "",
            "DELETE FROM breeds",
            "WHERE species_id IN (SELECT id FROM species WHERE code IN (",
            ",\n".join(f"    {sql_value(code)}" for code in codes),
            "));",
            "",
            "DELETE FROM species",
            "WHERE code IN (",
            ",\n".join(f"    {sql_value(code)}" for code in codes),
            ");",
        ]
    )


def stable_uuid(key: str) -> str:
    digest = bytearray(hashlib.sha1(f"pawly.pet.catalog:{key}".encode("utf-8")).digest()[:16])
    digest[6] = (digest[6] & 0x0F) | 0x50
    digest[8] = (digest[8] & 0x3F) | 0x80
    return str(uuid.UUID(bytes=bytes(digest)))


def clean_code(value: str) -> str:
    return clean_text(value).upper()


def clean_text(value: str) -> str:
    return str(value).strip()


class raw:
    def __init__(self, value: str) -> None:
        self.value = value


class cast:
    def __init__(self, value: str, target_type: str) -> None:
        self.value = value
        self.target_type = target_type


def sql_tuple(*values: object) -> str:
    return "(" + ", ".join(sql_value(value) for value in values) + ")"


def sql_value(value: object) -> str:
    if isinstance(value, raw):
        return value.value
    if isinstance(value, cast):
        return f"{sql_value(value.value)}::{value.target_type}"
    if isinstance(value, bool):
        return "TRUE" if value else "FALSE"
    if isinstance(value, int):
        return str(value)
    escaped = str(value).replace("'", "''")
    return f"'{escaped}'"


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as err:
        print(err, file=sys.stderr)
        raise SystemExit(1)
