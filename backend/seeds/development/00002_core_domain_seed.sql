INSERT INTO wallet_accounts (
    user_id,
    account_name,
    owner_type,
    available_balance_minor,
    overdraft_limit_minor,
    status
)
SELECT
    users.id,
    users.name || ' Wallet',
    'user',
    CASE users.institutional_id
        WHEN '220041234' THEN 10000
        WHEN '220049999' THEN 0
        ELSE 0
    END,
    CASE users.institutional_id
        WHEN '220041234' THEN 200
        ELSE 0
    END,
    CASE users.status
        WHEN 'SUSPENDED' THEN 'SUSPENDED'
        ELSE 'ACTIVE'
    END
FROM users
WHERE users.role = 'student'
ON CONFLICT (user_id) DO UPDATE
SET
    account_name = EXCLUDED.account_name,
    available_balance_minor = EXCLUDED.available_balance_minor,
    overdraft_limit_minor = EXCLUDED.overdraft_limit_minor,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO wallet_accounts (
    system_account_code,
    account_name,
    owner_type,
    available_balance_minor,
    overdraft_limit_minor,
    status
)
VALUES
    ('bus_pool', 'Bus Fare Pool', 'system', 0, 0, 'ACTIVE'),
    ('cashier_pool', 'Cashier Settlement Pool', 'system', 0, 0, 'ACTIVE')
ON CONFLICT (system_account_code) DO UPDATE
SET
    account_name = EXCLUDED.account_name,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO routes (
    id,
    code,
    name,
    fare_policy_type,
    default_fare_minor,
    status
)
VALUES
    ('00000000-0000-0000-0000-0000000000a1', 'A', 'Route A', 'FLAT_ROUTE', 2000, 'ACTIVE'),
    ('00000000-0000-0000-0000-0000000000b2', 'B', 'Route B', 'STOP_MATRIX', NULL, 'ACTIVE')
ON CONFLICT (code) DO UPDATE
SET
    name = EXCLUDED.name,
    fare_policy_type = EXCLUDED.fare_policy_type,
    default_fare_minor = EXCLUDED.default_fare_minor,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO stops (
    id,
    name,
    public_label,
    position
)
VALUES
    ('10000000-0000-0000-0000-000000000001', 'Campus', 'Campus', ST_SetSRID(ST_MakePoint(90.3804, 23.8252), 4326)::geography),
    ('10000000-0000-0000-0000-000000000002', 'Airport', 'Airport', ST_SetSRID(ST_MakePoint(90.4030, 23.8510), 4326)::geography),
    ('10000000-0000-0000-0000-000000000003', 'House Building', 'House Building', ST_SetSRID(ST_MakePoint(90.4010, 23.8750), 4326)::geography),
    ('10000000-0000-0000-0000-000000000004', 'Azampur', 'Azampur', ST_SetSRID(ST_MakePoint(90.4080, 23.8690), 4326)::geography),
    ('10000000-0000-0000-0000-000000000005', 'Abdullahpur', 'Abdullahpur', ST_SetSRID(ST_MakePoint(90.4060, 23.8860), 4326)::geography)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    public_label = EXCLUDED.public_label,
    position = EXCLUDED.position,
    updated_at = NOW();

DELETE FROM route_stop_sequences
WHERE route_id IN (
    '00000000-0000-0000-0000-0000000000a1',
    '00000000-0000-0000-0000-0000000000b2'
);

INSERT INTO route_stop_sequences (id, route_id, stop_id, stop_order)
VALUES
    ('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000a1', '10000000-0000-0000-0000-000000000001', 1),
    ('20000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-0000000000a1', '10000000-0000-0000-0000-000000000002', 2),
    ('20000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-0000000000a1', '10000000-0000-0000-0000-000000000003', 3),
    ('20000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000001', 1),
    ('20000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000004', 2),
    ('20000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000005', 3);

DELETE FROM route_fare_rules
WHERE route_id = '00000000-0000-0000-0000-0000000000b2';

INSERT INTO route_fare_rules (id, route_id, stop_id, service_label, fare_minor)
VALUES
    ('30000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000004', NULL, 1500),
    ('30000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000005', NULL, 2000),
    ('30000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000004', 'evening', 0),
    ('30000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-0000000000b2', '10000000-0000-0000-0000-000000000005', 'evening', 0);

INSERT INTO buses (
    id,
    bus_code,
    plate,
    default_route_id,
    status,
    seat_capacity,
    qr_version,
    notes
)
VALUES
    ('40000000-0000-0000-0000-000000000001', '1042', 'DHAKA-M-1042', '00000000-0000-0000-0000-0000000000a1', 'active', 50, 1, 'Primary Route A bus'),
    ('40000000-0000-0000-0000-000000000002', '1043', 'DHAKA-M-1043', '00000000-0000-0000-0000-0000000000b2', 'active', 50, 1, 'Primary Route B bus')
ON CONFLICT (bus_code) DO UPDATE
SET
    plate = EXCLUDED.plate,
    default_route_id = EXCLUDED.default_route_id,
    status = EXCLUDED.status,
    seat_capacity = EXCLUDED.seat_capacity,
    qr_version = EXCLUDED.qr_version,
    notes = EXCLUDED.notes,
    updated_at = NOW();

INSERT INTO service_calendars (
    id,
    route_id,
    weekday_mask,
    effective_from,
    effective_to
)
VALUES
    ('50000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000a1', 62, CURRENT_DATE - 30, NULL),
    ('50000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-0000000000b2', 62, CURRENT_DATE - 30, NULL)
ON CONFLICT (id) DO UPDATE
SET
    route_id = EXCLUDED.route_id,
    weekday_mask = EXCLUDED.weekday_mask,
    effective_from = EXCLUDED.effective_from,
    effective_to = EXCLUDED.effective_to,
    updated_at = NOW();

INSERT INTO trip_templates (
    id,
    route_id,
    service_calendar_id,
    service_label,
    name,
    status,
    early_boarding_window_minutes,
    late_grace_window_minutes
)
VALUES
    ('60000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-0000000000a1', '50000000-0000-0000-0000-000000000001', 'morning', 'Route A Morning Run', 'active', 30, 15),
    ('60000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-0000000000a1', '50000000-0000-0000-0000-000000000001', 'evening', 'Route A Evening Run', 'active', 30, 15),
    ('60000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-0000000000b2', '50000000-0000-0000-0000-000000000002', 'morning', 'Route B Morning Run', 'active', 30, 15),
    ('60000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-0000000000b2', '50000000-0000-0000-0000-000000000002', 'evening', 'Route B Evening Run', 'active', 30, 15)
ON CONFLICT (id) DO UPDATE
SET
    route_id = EXCLUDED.route_id,
    service_calendar_id = EXCLUDED.service_calendar_id,
    service_label = EXCLUDED.service_label,
    name = EXCLUDED.name,
    status = EXCLUDED.status,
    early_boarding_window_minutes = EXCLUDED.early_boarding_window_minutes,
    late_grace_window_minutes = EXCLUDED.late_grace_window_minutes,
    updated_at = NOW();

DELETE FROM trip_stop_times
WHERE trip_template_id IN (
    '60000000-0000-0000-0000-000000000001',
    '60000000-0000-0000-0000-000000000002',
    '60000000-0000-0000-0000-000000000003',
    '60000000-0000-0000-0000-000000000004'
);

INSERT INTO trip_stop_times (id, trip_template_id, stop_id, offset_minutes)
VALUES
    ('70000000-0000-0000-0000-000000000001', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 0),
    ('70000000-0000-0000-0000-000000000002', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002', 20),
    ('70000000-0000-0000-0000-000000000003', '60000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003', 35),
    ('70000000-0000-0000-0000-000000000004', '60000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 0),
    ('70000000-0000-0000-0000-000000000005', '60000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002', 25),
    ('70000000-0000-0000-0000-000000000006', '60000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000003', 40),
    ('70000000-0000-0000-0000-000000000007', '60000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', 0),
    ('70000000-0000-0000-0000-000000000008', '60000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000004', 25),
    ('70000000-0000-0000-0000-000000000009', '60000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000005', 40),
    ('70000000-0000-0000-0000-000000000010', '60000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', 0),
    ('70000000-0000-0000-0000-000000000011', '60000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000004', 25),
    ('70000000-0000-0000-0000-000000000012', '60000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000005', 40);
