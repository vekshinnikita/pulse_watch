BEGIN;

CREATE TYPE metric_period_type AS ENUM ('minute', 'ten_minutes', 'hour', 'day');
CREATE TYPE metric_type AS ENUM ('total_requests', 'critical', 'errors', 'warnings', 'info', 'status', 'unique_users');

CREATE TABLE app_metric (
    id              SERIAL PRIMARY KEY,
    app_id          INT NOT NULL,
    period_start    TIMESTAMPTZ NOT NULL,
    period_type     metric_period_type NOT NULL,
    type            metric_type NOT NULL,
    is_unique       BOOL DEFAULT FALSE,
    params          jsonb,
    value           INT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),

    FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE
);

COMMENT ON TABLE app_metric IS 'Таблица метриков сервиса';
COMMENT ON COLUMN app_metric.app_id IS 'Связь many_to_one к сервису (app)';
COMMENT ON COLUMN app_metric.period_start IS 'Метка времени начала периода';
COMMENT ON COLUMN app_metric.period_type IS 'Тип периода';
COMMENT ON COLUMN app_metric.type IS 'Тип метрики';
COMMENT ON COLUMN app_metric.is_unique IS 'Метрика является подсчетом уникальных значений';
COMMENT ON COLUMN app_metric.params IS 'Параметры метрики';
COMMENT ON COLUMN app_metric.value IS 'Значение метрики';
COMMENT ON COLUMN app_metric.created_at IS 'Метка временя когда создан';

CREATE INDEX idx_app_metric_created_at ON app_metric(created_at);
CREATE INDEX idx_app_metric_period_start ON app_metric(period_start);
CREATE INDEX idx_app_metric_period_type ON app_metric(period_type);
CREATE INDEX idx_app_metric_type ON app_metric(type);
CREATE INDEX idx_app_metric_value ON app_metric(value);
CREATE INDEX idx_app_metric_is_unique ON app_metric(is_unique);

COMMIT;