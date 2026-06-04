CREATE TABLE public.user_address (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    province_code VARCHAR(10),
    district_code VARCHAR(10),
    ward_code VARCHAR(10),

    province_name VARCHAR(100),
    district_name VARCHAR(100),
    ward_name VARCHAR(100),

    address VARCHAR(250)
);
