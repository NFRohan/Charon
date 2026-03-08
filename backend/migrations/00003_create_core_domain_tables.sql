-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE wallet_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id) ON DELETE RESTRICT,
    system_account_code TEXT NULL,
    account_name TEXT NOT NULL,
    owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'system')),
    available_balance_minor BIGINT NOT NULL DEFAULT 0,
    overdraft_limit_minor BIGINT NOT NULL DEFAULT 0 CHECK (overdraft_limit_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'SUSPENDED', 'CLOSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallet_accounts_owner_presence CHECK (
        (owner_type = 'user' AND user_id IS NOT NULL AND system_account_code IS NULL) OR
        (owner_type = 'system' AND user_id IS NULL AND system_account_code IS NOT NULL)
    )
);

ALTER TABLE wallet_accounts
    ADD CONSTRAINT wallet_accounts_user_id_key UNIQUE (user_id);

ALTER TABLE wallet_accounts
    ADD CONSTRAINT wallet_accounts_system_account_code_key UNIQUE (system_account_code);

CREATE INDEX idx_wallet_accounts_status ON wallet_accounts (status);

CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    fare_policy_type TEXT NOT NULL CHECK (fare_policy_type IN ('FLAT_ROUTE', 'STOP_MATRIX', 'ZERO_FARE')),
    default_fare_minor BIGINT NULL CHECK (default_fare_minor IS NULL OR default_fare_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE', 'ARCHIVED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE stops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    public_label TEXT NOT NULL,
    position GEOGRAPHY(Point, 4326) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stops_position ON stops USING GIST (position);

CREATE TABLE route_stop_sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    stop_id UUID NOT NULL REFERENCES stops(id) ON DELETE CASCADE,
    stop_order INTEGER NOT NULL CHECK (stop_order > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT route_stop_sequences_route_stop_key UNIQUE (route_id, stop_id),
    CONSTRAINT route_stop_sequences_route_order_key UNIQUE (route_id, stop_order)
);

CREATE INDEX idx_route_stop_sequences_route_id ON route_stop_sequences (route_id, stop_order);

CREATE TABLE route_fare_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    stop_id UUID NOT NULL REFERENCES stops(id) ON DELETE CASCADE,
    service_label TEXT NULL,
    fare_minor BIGINT NOT NULL CHECK (fare_minor >= 0),
    effective_from TIMESTAMPTZ NULL,
    effective_to TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT route_fare_rules_effective_window CHECK (
        effective_to IS NULL OR effective_from IS NULL OR effective_to > effective_from
    )
);

CREATE UNIQUE INDEX idx_route_fare_rules_unique
    ON route_fare_rules (route_id, stop_id, service_label) NULLS NOT DISTINCT;

CREATE INDEX idx_route_fare_rules_route_id ON route_fare_rules (route_id);

CREATE TABLE buses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bus_code TEXT NOT NULL UNIQUE,
    plate TEXT NULL UNIQUE,
    default_route_id UUID NULL REFERENCES routes(id) ON DELETE SET NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'maintenance', 'retired', 'out_of_service')),
    seat_capacity INTEGER NOT NULL CHECK (seat_capacity > 0),
    qr_version INTEGER NOT NULL DEFAULT 1 CHECK (qr_version > 0),
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE service_calendars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    weekday_mask INTEGER NOT NULL CHECK (weekday_mask >= 0 AND weekday_mask <= 127),
    effective_from DATE NULL,
    effective_to DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT service_calendars_effective_window CHECK (
        effective_to IS NULL OR effective_from IS NULL OR effective_to >= effective_from
    )
);

CREATE INDEX idx_service_calendars_route_id ON service_calendars (route_id);

CREATE TABLE trip_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    service_calendar_id UUID NOT NULL REFERENCES service_calendars(id) ON DELETE CASCADE,
    service_label TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive', 'archived')),
    early_boarding_window_minutes INTEGER NOT NULL DEFAULT 30 CHECK (early_boarding_window_minutes >= 0),
    late_grace_window_minutes INTEGER NOT NULL DEFAULT 15 CHECK (late_grace_window_minutes >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trip_templates_route_id ON trip_templates (route_id);
CREATE INDEX idx_trip_templates_service_calendar_id ON trip_templates (service_calendar_id);

CREATE TABLE trip_stop_times (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_template_id UUID NOT NULL REFERENCES trip_templates(id) ON DELETE CASCADE,
    stop_id UUID NOT NULL REFERENCES stops(id) ON DELETE CASCADE,
    offset_minutes INTEGER NOT NULL CHECK (offset_minutes >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT trip_stop_times_trip_stop_key UNIQUE (trip_template_id, stop_id)
);

CREATE INDEX idx_trip_stop_times_trip_template_id ON trip_stop_times (trip_template_id, offset_minutes);

CREATE TABLE service_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_calendar_id UUID NOT NULL REFERENCES service_calendars(id) ON DELETE CASCADE,
    service_date DATE NOT NULL,
    exception_type TEXT NOT NULL CHECK (exception_type IN ('CANCELLATION', 'TIME_OVERRIDE')),
    reason_code TEXT NOT NULL,
    override_start_time TIME NULL,
    override_end_time TIME NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT service_exceptions_calendar_date_key UNIQUE (service_calendar_id, service_date),
    CONSTRAINT service_exceptions_override_rules CHECK (
        (exception_type = 'CANCELLATION' AND override_start_time IS NULL AND override_end_time IS NULL) OR
        (exception_type = 'TIME_OVERRIDE' AND override_start_time IS NOT NULL AND override_end_time IS NOT NULL)
    )
);

CREATE INDEX idx_service_exceptions_service_date ON service_exceptions (service_date);

CREATE TABLE service_advisories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NULL REFERENCES routes(id) ON DELETE SET NULL,
    advisory_type TEXT NOT NULL,
    message TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT service_advisories_window CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX idx_service_advisories_route_id ON service_advisories (route_id);
CREATE INDEX idx_service_advisories_starts_at ON service_advisories (starts_at);

CREATE TABLE route_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_template_id UUID NULL REFERENCES trip_templates(id) ON DELETE SET NULL,
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE RESTRICT,
    bus_id UUID NOT NULL REFERENCES buses(id) ON DELETE RESTRICT,
    session_source TEXT NOT NULL CHECK (session_source IN ('scheduled', 'admin_ad_hoc')),
    service_label TEXT NOT NULL,
    scheduled_start TIMESTAMPTZ NOT NULL,
    scheduled_end TIMESTAMPTZ NULL,
    driver_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    started_at TIMESTAMPTZ NULL,
    ended_at TIMESTAMPTZ NULL,
    status TEXT NOT NULL CHECK (status IN ('scheduled', 'boarding_open', 'running', 'completed', 'expired', 'cancelled', 'conflicted', 'force_closed')),
    notes TEXT NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT route_sessions_scheduled_window CHECK (scheduled_end IS NULL OR scheduled_end > scheduled_start),
    CONSTRAINT route_sessions_runtime_window CHECK (ended_at IS NULL OR started_at IS NULL OR ended_at >= started_at)
);

CREATE INDEX idx_route_sessions_route_id ON route_sessions (route_id);
CREATE INDEX idx_route_sessions_bus_id ON route_sessions (bus_id);
CREATE INDEX idx_route_sessions_status ON route_sessions (status);
CREATE INDEX idx_route_sessions_scheduled_start ON route_sessions (scheduled_start);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL CHECK (type IN ('BOARDING_FARE', 'ADMIN_CREDIT', 'REFUND', 'SPONSORED_BOARDING', 'EMERGENCY_REDEMPTION', 'ADJUSTMENT')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED', 'CANCELLED')),
    actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    route_session_id UUID NULL REFERENCES route_sessions(id) ON DELETE RESTRICT,
    reason_code TEXT NULL,
    approval_chain_json JSONB NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_actor_user_id ON transactions (actor_user_id);
CREATE INDEX idx_transactions_route_session_id ON transactions (route_session_id);
CREATE INDEX idx_transactions_created_at ON transactions (created_at DESC);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES wallet_accounts(id) ON DELETE RESTRICT,
    direction TEXT NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries (transaction_id);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries (account_id, created_at DESC);

CREATE TABLE finance_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_account_id UUID NOT NULL REFERENCES wallet_accounts(id) ON DELETE RESTRICT,
    transaction_id UUID NULL REFERENCES transactions(id) ON DELETE SET NULL,
    adjustment_type TEXT NOT NULL CHECK (adjustment_type IN ('CREDIT', 'REFUND')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approved_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    approval_status TEXT NOT NULL CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED', 'NOT_REQUIRED')),
    reason_code TEXT NOT NULL,
    before_balance_minor BIGINT NULL,
    after_balance_minor BIGINT NULL,
    request_note TEXT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT finance_adjustments_decision_window CHECK (
        decided_at IS NULL OR decided_at >= requested_at
    )
);

CREATE INDEX idx_finance_adjustments_wallet_account_id ON finance_adjustments (wallet_account_id, requested_at DESC);
CREATE INDEX idx_finance_adjustments_approval_status ON finance_adjustments (approval_status, requested_at DESC);

CREATE TABLE emergency_ride_permits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    device_id TEXT NOT NULL,
    permit_token_hash TEXT NOT NULL UNIQUE,
    max_fare_minor BIGINT NOT NULL CHECK (max_fare_minor >= 0),
    status TEXT NOT NULL CHECK (status IN ('ISSUED', 'USED', 'REDEEMED', 'EXPIRED', 'VOID')),
    expires_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ NULL,
    redeemed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT emergency_ride_permits_usage_window CHECK (
        redeemed_at IS NULL OR redeemed_at >= issued_at
    )
);

CREATE INDEX idx_emergency_ride_permits_student_id ON emergency_ride_permits (student_id, status);

CREATE TABLE boarding_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    route_session_id UUID NOT NULL REFERENCES route_sessions(id) ON DELETE RESTRICT,
    selected_stop_id UUID NULL REFERENCES stops(id) ON DELETE SET NULL,
    paid_by_student_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    transaction_id UUID NULL REFERENCES transactions(id) ON DELETE SET NULL,
    boarding_mode TEXT NOT NULL CHECK (boarding_mode IN ('standard', 'sponsored', 'emergency_sync')),
    fare_minor BIGINT NOT NULL CHECK (fare_minor >= 0),
    fare_policy_type TEXT NOT NULL CHECK (fare_policy_type IN ('FLAT_ROUTE', 'STOP_MATRIX', 'ZERO_FARE')),
    charge_mode TEXT NOT NULL,
    exemption_reason_code TEXT NULL,
    emergency_permit_id UUID NULL REFERENCES emergency_ride_permits(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT boarding_events_student_route_session_key UNIQUE (student_id, route_session_id)
);

CREATE INDEX idx_boarding_events_route_session_id ON boarding_events (route_session_id);
CREATE INDEX idx_boarding_events_transaction_id ON boarding_events (transaction_id);

CREATE TABLE telemetry_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_session_id UUID NOT NULL REFERENCES route_sessions(id) ON DELETE CASCADE,
    bus_id UUID NOT NULL REFERENCES buses(id) ON DELETE RESTRICT,
    position GEOGRAPHY(Point, 4326) NOT NULL,
    is_replayed BOOLEAN NOT NULL DEFAULT FALSE,
    speed_kph NUMERIC(6, 2) NULL CHECK (speed_kph IS NULL OR speed_kph >= 0),
    heading NUMERIC(6, 2) NULL,
    accuracy_m NUMERIC(6, 2) NULL CHECK (accuracy_m IS NULL OR accuracy_m >= 0),
    recorded_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telemetry_points_route_session_id ON telemetry_points (route_session_id, recorded_at DESC);
CREATE INDEX idx_telemetry_points_position ON telemetry_points USING GIST (position);

CREATE TABLE outbox_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload_json JSONB NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL,
    last_error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_events_pending ON outbox_events (available_at, created_at) WHERE published_at IS NULL;

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    subject_type TEXT NOT NULL,
    subject_id TEXT NULL,
    action_type TEXT NOT NULL,
    result TEXT NOT NULL,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_actor_id ON audit_logs (actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_subject ON audit_logs (subject_type, subject_id, created_at DESC);

CREATE TABLE audit_investigation_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_log_id UUID NOT NULL REFERENCES audit_logs(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    note_body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_investigation_notes_audit_log_id ON audit_investigation_notes (audit_log_id, created_at);

CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status TEXT NOT NULL CHECK (status IN ('open', 'acknowledged', 'resolved', 'muted')),
    route_id UUID NULL REFERENCES routes(id) ON DELETE SET NULL,
    bus_id UUID NULL REFERENCES buses(id) ON DELETE SET NULL,
    route_session_id UUID NULL REFERENCES route_sessions(id) ON DELETE SET NULL,
    message TEXT NOT NULL,
    investigation_notes_count INTEGER NOT NULL DEFAULT 0 CHECK (investigation_notes_count >= 0),
    last_actor_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMPTZ NULL,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_status ON alerts (status, opened_at DESC);
CREATE INDEX idx_alerts_route_id ON alerts (route_id);
CREATE INDEX idx_alerts_bus_id ON alerts (bus_id);

CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_wallet_accounts_set_updated_at
    BEFORE UPDATE ON wallet_accounts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_routes_set_updated_at
    BEFORE UPDATE ON routes
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_stops_set_updated_at
    BEFORE UPDATE ON stops
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_route_fare_rules_set_updated_at
    BEFORE UPDATE ON route_fare_rules
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_finance_adjustments_set_updated_at
    BEFORE UPDATE ON finance_adjustments
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_buses_set_updated_at
    BEFORE UPDATE ON buses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_service_calendars_set_updated_at
    BEFORE UPDATE ON service_calendars
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_trip_templates_set_updated_at
    BEFORE UPDATE ON trip_templates
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_service_exceptions_set_updated_at
    BEFORE UPDATE ON service_exceptions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_service_advisories_set_updated_at
    BEFORE UPDATE ON service_advisories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_route_sessions_set_updated_at
    BEFORE UPDATE ON route_sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_alerts_set_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_alerts_set_updated_at ON alerts;
DROP TRIGGER IF EXISTS trg_route_sessions_set_updated_at ON route_sessions;
DROP TRIGGER IF EXISTS trg_service_advisories_set_updated_at ON service_advisories;
DROP TRIGGER IF EXISTS trg_service_exceptions_set_updated_at ON service_exceptions;
DROP TRIGGER IF EXISTS trg_trip_templates_set_updated_at ON trip_templates;
DROP TRIGGER IF EXISTS trg_service_calendars_set_updated_at ON service_calendars;
DROP TRIGGER IF EXISTS trg_buses_set_updated_at ON buses;
DROP TRIGGER IF EXISTS trg_finance_adjustments_set_updated_at ON finance_adjustments;
DROP TRIGGER IF EXISTS trg_route_fare_rules_set_updated_at ON route_fare_rules;
DROP TRIGGER IF EXISTS trg_stops_set_updated_at ON stops;
DROP TRIGGER IF EXISTS trg_routes_set_updated_at ON routes;
DROP TRIGGER IF EXISTS trg_wallet_accounts_set_updated_at ON wallet_accounts;
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;

DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS audit_investigation_notes;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS telemetry_points;
DROP TABLE IF EXISTS boarding_events;
DROP TABLE IF EXISTS emergency_ride_permits;
DROP TABLE IF EXISTS finance_adjustments;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS route_sessions;
DROP TABLE IF EXISTS service_advisories;
DROP TABLE IF EXISTS service_exceptions;
DROP TABLE IF EXISTS trip_stop_times;
DROP TABLE IF EXISTS trip_templates;
DROP TABLE IF EXISTS service_calendars;
DROP TABLE IF EXISTS buses;
DROP TABLE IF EXISTS route_fare_rules;
DROP TABLE IF EXISTS route_stop_sequences;
DROP TABLE IF EXISTS stops;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS wallet_accounts;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_at();
-- +goose StatementEnd
