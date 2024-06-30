CREATE TABLE IF NOT EXISTS public.user
(
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name   TEXT NOT NULL,
    hash_password TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email ON public.user(email);

CREATE TABLE IF NOT EXISTS app
(
    id     INTEGER PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE,
    refresh_secret TEXT NOT NULL UNIQUE 
);
CREATE INDEX IF NOT EXISTS idx_app_id ON app(id);

CREATE TABLE IF NOT EXISTS admin
(
    id       INTEGER PRIMARY KEY,
    user_id  INTEGER NOT NULL REFERENCES public.user(id),
    app_id   INTEGER NOT NULL REFERENCES app(id)
);
CREATE INDEX IF NOT EXISTS idx_admin_user_id ON admin(user_id);