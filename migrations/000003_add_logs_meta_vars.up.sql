BEGIN;

CREATE TYPE log_meta_var_type AS ENUM ('number', 'string', 'date', 'datetime');

CREATE TABLE log_meta_var (
    id              SERIAL PRIMARY KEY,
    app_id          INT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(255) NOT NULL,
    type            log_meta_var_type NOT NULL,
    deleted         BOOLEAN DEFAULT false,

    UNIQUE (app_id, code),
    FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE
);

COMMENT ON TABLE log_meta_var IS 'Таблица мета переменных логов';
COMMENT ON COLUMN log_meta_var.app_id IS 'Связь many_to_one к сервису (app)';
COMMENT ON COLUMN log_meta_var.name IS 'Название';
COMMENT ON COLUMN log_meta_var.code IS 'Код';
COMMENT ON COLUMN log_meta_var.type IS 'Тип переменной';
COMMENT ON COLUMN log_meta_var.deleted IS 'Пометка об удалении';

COMMIT;