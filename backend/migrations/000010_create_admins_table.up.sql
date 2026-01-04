-- Create admin role enum
CREATE TYPE admin_role AS ENUM ('super_admin', 'admin', 'operator');

-- Create admin status enum
CREATE TYPE admin_status AS ENUM ('active', 'inactive', 'suspended');

-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- Admin details
    full_name VARCHAR(100),
    role admin_role NOT NULL DEFAULT 'operator',
    status admin_status NOT NULL DEFAULT 'active',

    -- Permissions (can be extended)
    permissions JSONB DEFAULT '[]'::jsonb,

    -- Security
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip VARCHAR(45),
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES admins(id),
    updated_by UUID REFERENCES admins(id),

    -- Soft delete
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_admins_username ON admins(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_admins_email ON admins(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_admins_role ON admins(role) WHERE deleted_at IS NULL;
CREATE INDEX idx_admins_status ON admins(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_admins_created_at ON admins(created_at);

-- Add comments
COMMENT ON TABLE admins IS 'Administrative users with access to admin panel and APIs';
COMMENT ON COLUMN admins.role IS 'Admin role: super_admin (full access), admin (most access), operator (limited access)';
COMMENT ON COLUMN admins.status IS 'Admin account status: active, inactive, or suspended';
COMMENT ON COLUMN admins.permissions IS 'Additional granular permissions as JSON array';
COMMENT ON COLUMN admins.failed_login_attempts IS 'Counter for failed login attempts (for account lockout)';
COMMENT ON COLUMN admins.locked_until IS 'Account locked until this timestamp (after too many failed attempts)';

-- Insert default super admin (password: "admin123" - CHANGE IN PRODUCTION!)
-- Password hash generated with bcrypt cost 10
INSERT INTO admins (username, email, password_hash, full_name, role, status)
VALUES (
    'superadmin',
    'admin@slotmachine.com',
    '$2a$10$oxt0Tj37q/LTdNSHmrlXzubLZnlbfQEGKD2BJ8fsei5F6s513EKA.', -- "admin123"
    'Super Administrator',
    'super_admin',
    'active'
);
