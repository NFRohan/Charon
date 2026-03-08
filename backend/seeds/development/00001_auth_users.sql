DELETE FROM auth_sessions;

INSERT INTO users (
	role,
	name,
	institutional_id,
	status,
	fare_exempt,
	password_hash
)
VALUES
	('student', 'Student Demo', '220041234', 'ACTIVE', FALSE, '$argon2id$v=19$m=19456,t=2,p=1$t8bKp8mhgNX4oA5lbb3ebA$lQGYtT6jDWBVlGNfd/ZQ4oS6OnI7eVNs2B0yzLCf8Qc'),
	('driver', 'Driver Demo', 'DRV1001', 'ACTIVE', FALSE, '$argon2id$v=19$m=19456,t=2,p=1$t8bKp8mhgNX4oA5lbb3ebA$lQGYtT6jDWBVlGNfd/ZQ4oS6OnI7eVNs2B0yzLCf8Qc'),
	('cashier', 'Cashier Demo', 'CASH1001', 'ACTIVE', FALSE, '$argon2id$v=19$m=19456,t=2,p=1$t8bKp8mhgNX4oA5lbb3ebA$lQGYtT6jDWBVlGNfd/ZQ4oS6OnI7eVNs2B0yzLCf8Qc'),
	('admin', 'Admin Demo', 'ADM1001', 'ACTIVE', FALSE, '$argon2id$v=19$m=19456,t=2,p=1$t8bKp8mhgNX4oA5lbb3ebA$lQGYtT6jDWBVlGNfd/ZQ4oS6OnI7eVNs2B0yzLCf8Qc'),
	('student', 'Suspended Demo', '220049999', 'SUSPENDED', FALSE, '$argon2id$v=19$m=19456,t=2,p=1$t8bKp8mhgNX4oA5lbb3ebA$lQGYtT6jDWBVlGNfd/ZQ4oS6OnI7eVNs2B0yzLCf8Qc')
ON CONFLICT (institutional_id) DO UPDATE
SET
	role = EXCLUDED.role,
	name = EXCLUDED.name,
	status = EXCLUDED.status,
	fare_exempt = EXCLUDED.fare_exempt,
	password_hash = EXCLUDED.password_hash,
	password_changed_at = NOW(),
	updated_at = NOW();
