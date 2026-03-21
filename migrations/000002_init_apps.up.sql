BEGIN;

CREATE TABLE app (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    code            VARCHAR(100) UNIQUE NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted         BOOLEAN DEFAULT false 
) ;

COMMENT ON TABLE app IS 'Таблица сервисов';
COMMENT ON COLUMN app.name IS 'Название сервиса';
COMMENT ON COLUMN app.code IS 'Код сервиса';
COMMENT ON COLUMN app.description IS 'Описание сервиса';
COMMENT ON COLUMN app.created_at IS 'Метка времени когда создан';
COMMENT ON COLUMN app.deleted IS 'Пометка об удалении';


CREATE TABLE api_key (
    id              SERIAL PRIMARY KEY,
    app_id          INT NOT NULL REFERENCES app(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    key_hash        TEXT,
    expires_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ,
    revoked         BOOLEAN DEFAULT false
);

COMMENT ON TABLE api_key IS 'Таблица API ключей';
COMMENT ON COLUMN api_key.app_id IS 'Связь many_to_one к сервису (app)';
COMMENT ON COLUMN api_key.app_id IS 'Название API ключа';
COMMENT ON COLUMN api_key.key_hash IS 'Хэш ключа';
COMMENT ON COLUMN api_key.expires_at IS 'Метка времени когда истекает ключ';
COMMENT ON COLUMN api_key.created_at IS 'Метка времени когда создан';
COMMENT ON COLUMN api_key.revoked IS 'Пометка об аннулировании';

CREATE INDEX idx_api_key_key ON api_key(key_hash);
CREATE INDEX idx_api_key_created_at ON api_key(created_at);
CREATE INDEX idx_api_key_expire_at ON api_key(expires_at);

COMMIT;