BEGIN;

CREATE TYPE alert_rule_log_level AS ENUM ('CRITICAL', 'ERROR', 'WARNING');
CREATE TABLE alert_rule (
    id              SERIAL PRIMARY KEY,
    app_id          INT,
    name            VARCHAR(255),
    level           alert_rule_log_level NOT NULL,
	threshold       INT NOT NULL,
    message         TEXT,
	interval        INT NOT NULL,
    user_id         INT NOT NULL,
    deleted         BOOLEAN DEFAULT false,
    
    UNIQUE (app_id, level, threshold, interval, user_id),
    CONSTRAINT alert_rule_check_threshold_positive CHECK (threshold > 0),
    CONSTRAINT alert_rule_check_interval_positive CHECK (interval > 0),
    
    FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES app_user(id) ON DELETE CASCADE
);
COMMENT ON TABLE alert_rule IS 'Таблица настройки правил алертов';
COMMENT ON COLUMN alert_rule.app_id IS 'Связь many_to_one к сервису, если NULL то будет работать для всех сервисов (app)';
COMMENT ON COLUMN alert_rule.name IS 'Название';
COMMENT ON COLUMN alert_rule.level IS 'Уровень лога, который будет учитываться';
COMMENT ON COLUMN alert_rule.threshold IS 'Порог, при котором будет отправлен алерт';
COMMENT ON COLUMN alert_rule.message IS 'Текст сообщения, который будет отправлен в случае активации правила';
COMMENT ON COLUMN alert_rule.interval IS 'Период подсчета логов (минуты)';
COMMENT ON COLUMN alert_rule.user_id IS 'Пользователь, которому будет отправляться алерт (app_user)';
COMMENT ON COLUMN alert_rule.deleted IS 'Пометка об удалении';


CREATE TYPE alert_status AS ENUM ('new', 'sent');
CREATE TABLE alert (
    id              SERIAL PRIMARY KEY,
    app_id         INT NOT NULL,
    rule_id         INT NOT NULL,
    message         TEXT,
    status          alert_status DEFAULT 'new',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,

    FOREIGN KEY (rule_id) REFERENCES alert_rule(id) ON DELETE CASCADE,
    FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE
);
CREATE INDEX idx_alert_status ON alert(status);
CREATE INDEX idx_alert_created_at ON alert(created_at);
CREATE INDEX idx_alert_resolved_at ON alert(resolved_at);

COMMENT ON TABLE alert IS 'Таблица алертов';
COMMENT ON COLUMN alert.app_id IS 'Связка с приложением (app)';
COMMENT ON COLUMN alert.rule_id IS 'Выполненное правило, по которому создан алерт';
COMMENT ON COLUMN alert.message IS 'Текст сообщения';
COMMENT ON COLUMN alert.status IS 'Статус';
COMMENT ON COLUMN alert.created_at IS 'Метка времени когда создан';
COMMENT ON COLUMN alert.resolved_at IS 'Метка времени когда отправлен';


COMMIT;