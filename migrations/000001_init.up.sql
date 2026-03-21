BEGIN;

-- Пользователи и роли START

-- Роли и права (для гибкости, если будешь расширять ACL)
CREATE TABLE permission (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    code            VARCHAR(63) NOT NULL UNIQUE,
    deleted         BOOLEAN DEFAULT false
);

COMMENT ON TABLE permission IS 'Таблица прав';
COMMENT ON COLUMN permission.name IS 'Название прав';
COMMENT ON COLUMN permission.code IS 'Код прав';
COMMENT ON COLUMN permission.deleted IS 'Пометка об удалении';


CREATE TABLE role (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    code            VARCHAR(63) NOT NULL UNIQUE,
    deleted         BOOLEAN DEFAULT false
);

COMMENT ON TABLE role IS 'Таблица ролей';
COMMENT ON COLUMN role.name IS 'Название роли';
COMMENT ON COLUMN role.code IS 'Код роли';
COMMENT ON COLUMN role.deleted IS 'Пометка об удалении';

INSERT INTO role (id, name, code) VALUES 
    (1, 'Гость', 'guest'),
    (2, 'Админ', 'admin'),
    (3, 'Наблюдатель', 'observer');

-- Роли и права (для гибкости, если будешь расширять ACL)
CREATE TABLE role_permission (
    role_id         INT REFERENCES role(id) ON DELETE CASCADE,
    permission_id   INT REFERENCES permission(id) ON DELETE CASCADE,
    
    PRIMARY KEY (role_id, permission_id)
);

COMMENT ON TABLE role_permission IS 'Связь many_to_many между таблицами ролей (role) и прав (permission)';
COMMENT ON COLUMN role_permission.role_id IS 'Привязка к роли (role)';
COMMENT ON COLUMN role_permission.permission_id IS 'Привязка к правам (permission)';

-- Пользователи системы (администраторы, разработчики)
CREATE TABLE app_user (
    id              SERIAL PRIMARY KEY,
	name            VARCHAR(255) NOT NULL,
    username        VARCHAR(100) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE,
    tg_id           BIGINT UNIQUE,
    password_hash   TEXT NOT NULL,
    role_id         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    deleted         BOOLEAN DEFAULT false,

    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE SET DEFAULT
) ;

COMMENT ON TABLE app_user IS 'Таблица ролей';
COMMENT ON COLUMN app_user.name IS 'Имя пользователя для отображения';
COMMENT ON COLUMN app_user.username IS 'Имя пользователя';
COMMENT ON COLUMN app_user.email IS 'Электронная почта пользователя';
COMMENT ON COLUMN app_user.tg_id IS 'Идентификатор в телеграмме';
COMMENT ON COLUMN app_user.password_hash IS 'Хэш пароля';
COMMENT ON COLUMN app_user.role_id IS 'Роль пользователя';
COMMENT ON COLUMN app_user.created_at IS 'Метка времени когда создан';
COMMENT ON COLUMN app_user.updated_at IS 'Метка времени когда обновлен последний раз';
COMMENT ON COLUMN app_user.deleted IS 'Пометка об удалении';

CREATE INDEX idx_user_tg_id ON app_user(tg_id);
CREATE INDEX idx_user_created_at ON app_user(created_at);
CREATE INDEX idx_user_updated_at ON app_user(updated_at);


CREATE TABLE refresh_token (
    id              SERIAL PRIMARY KEY,
	user_id         INT NOT NULL,
    jti             UUID UNIQUE NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT false,
    
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE
);

COMMENT ON TABLE refresh_token IS 'Таблица refresh токенов';
COMMENT ON COLUMN refresh_token.user_id IS 'Привязка к пользователя (app_user)';;
COMMENT ON COLUMN refresh_token.jti IS 'Идентификатор токена';
COMMENT ON COLUMN refresh_token.expires_at IS 'Срок действия';
COMMENT ON COLUMN refresh_token.revoked IS 'Пометка об аннулировании';

CREATE INDEX idx_refresh_token_expires_at ON refresh_token(expires_at);
CREATE INDEX idx_refresh_token_jti ON refresh_token(jti);

-- Пользователи и роли END

COMMIT;