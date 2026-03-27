-- Add topology parameters to the fabrics table.
-- stages: number of Clos stages (2, 3, or 5)
-- radix: switch port count
-- oversubscription: oversubscription ratio (≥ 1.0)
-- *_model_id: optional device model assignments per role

ALTER TABLE fabrics ADD COLUMN stages           INTEGER NOT NULL DEFAULT 2;
ALTER TABLE fabrics ADD COLUMN radix            INTEGER NOT NULL DEFAULT 64;
ALTER TABLE fabrics ADD COLUMN oversubscription REAL    NOT NULL DEFAULT 1.0;
ALTER TABLE fabrics ADD COLUMN leaf_model_id       INTEGER REFERENCES device_models(id);
ALTER TABLE fabrics ADD COLUMN spine_model_id      INTEGER REFERENCES device_models(id);
ALTER TABLE fabrics ADD COLUMN super_spine_model_id INTEGER REFERENCES device_models(id);
