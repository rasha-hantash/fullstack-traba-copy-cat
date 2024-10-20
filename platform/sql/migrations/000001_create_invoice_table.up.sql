create EXTENSION if not exists "uuid-ossp";

create table shifts (
    id serial primary key,
    start_date varchar(255) not null,
    end_date date not null,
    shifts_filled date not null,
    invoice_id uuid not null,
    invoice_amount numeric(10, 2) not null,
    status varchar(255) not null, -- pending, approved, rejected
    user_id uuid not null, -- user who created the shift
    created_at timestamp not null default now(),
    created_by uuid not null,
    updated_at timestamp not null default now(),
    updated_by uuid not null
);