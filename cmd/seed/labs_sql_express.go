package main

// Fixtures + auto-checks for "Экспресс-курс по SQL" (sql-express).
//
// These lessons run in golearn/sandbox-pg, which ships a real PostgreSQL
// server; `pg-start` boots the cluster and gives root a superuser role, so the
// student can type `psql` exactly as the lesson describes.

// sqlSchemaDDL is the schema the later lessons assume already exists.
const sqlSchemaDDL = `psql -qd training_shop <<'SQL'
CREATE SCHEMA IF NOT EXISTS store;
CREATE TABLE IF NOT EXISTS store.customers (
    id            SERIAL PRIMARY KEY,
    full_name     TEXT NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS store.products (
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    category TEXT NOT NULL,
    price    NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    in_stock INT NOT NULL CHECK (in_stock >= 0)
);
CREATE TABLE IF NOT EXISTS store.orders (
    id          SERIAL PRIMARY KEY,
    customer_id INT NOT NULL REFERENCES store.customers(id),
    status      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS store.order_items (
    id         SERIAL PRIMARY KEY,
    order_id   INT NOT NULL REFERENCES store.orders(id),
    product_id INT NOT NULL REFERENCES store.products(id),
    quantity   INT NOT NULL CHECK (quantity > 0),
    price      NUMERIC(10,2) NOT NULL CHECK (price >= 0)
);
SQL`

// q runs a scalar query against training_shop and compares it to want.
func q(sql, want string) string {
	return `[ "$(psql -tAd training_shop -c ` + "\"" + sql + "\"" + ` 2>/dev/null | tr -d ' ')" = ` + want + ` ]`
}

var sqlExpressLabs = map[string]labSpec{
	// ── Lab 1: база, схема, таблицы ──
	"ch-pgsql-lab-schema": {
		Image: sandboxImagePG,
		Setup: `set -e
pg-start
psql -qc "DROP DATABASE IF EXISTS training_shop" >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: check(`psql -tAc "SELECT 1 FROM pg_database WHERE datname='training_shop'" 2>/dev/null | grep -q 1`,
				"база training_shop создана",
				"CREATE DATABASE training_shop; — список баз смотри через \\l, подключение: \\c training_shop"),
			2: check(q(`SELECT count(*) FROM information_schema.schemata WHERE schema_name='store'`, "1"),
				"схема store создана",
				"Подключись к базе (psql -d training_shop) и выполни: CREATE SCHEMA store; — список схем: \\dn"),
			3: check(`psql -tAd training_shop -c "SELECT string_agg(column_name, ',' ORDER BY column_name) FROM information_schema.columns WHERE table_schema='store' AND table_name='customers'" 2>/dev/null | grep -q 'email,full_name,id,registered_at' && `+
				q(`SELECT count(*) FROM information_schema.table_constraints WHERE table_schema='store' AND table_name='customers' AND constraint_type='PRIMARY KEY'`, "1")+` && `+
				q(`SELECT count(*) FROM information_schema.table_constraints WHERE table_schema='store' AND table_name='customers' AND constraint_type='UNIQUE'`, "1")+` && `+
				q(`SELECT count(*) FROM information_schema.columns WHERE table_schema='store' AND table_name='customers' AND column_name='registered_at' AND column_default IS NOT NULL`, "1"),
				"таблица store.customers с PK, уникальным email и значением по умолчанию",
				"CREATE TABLE store.customers (id SERIAL PRIMARY KEY, full_name TEXT NOT NULL, email TEXT UNIQUE NOT NULL, registered_at TIMESTAMPTZ DEFAULT now()); — структура: \\d store.customers"),
			4: check(`psql -tAd training_shop -c "SELECT string_agg(column_name, ',' ORDER BY column_name) FROM information_schema.columns WHERE table_schema='store' AND table_name='products'" 2>/dev/null | grep -q 'category,id,in_stock,name,price' && `+
				`[ "$(psql -tAd training_shop -c "SELECT count(*) FROM information_schema.table_constraints WHERE table_schema='store' AND table_name='products' AND constraint_type='CHECK' AND constraint_name NOT LIKE '%not_null%'" 2>/dev/null | tr -d ' ')" -ge 2 ]`,
				"таблица store.products с ограничениями на цену и остаток",
				"CREATE TABLE store.products (id SERIAL PRIMARY KEY, name TEXT NOT NULL, category TEXT NOT NULL, price NUMERIC(10,2) CHECK (price >= 0), in_stock INT CHECK (in_stock >= 0)); — ограничения видны в \\d store.products"),
			5: check(q(`SELECT count(*) FROM information_schema.table_constraints WHERE table_schema='store' AND table_name='orders' AND constraint_type='FOREIGN KEY'`, "1")+` && `+
				`psql -tAd training_shop -c "SELECT string_agg(column_name, ',' ORDER BY column_name) FROM information_schema.columns WHERE table_schema='store' AND table_name='orders'" 2>/dev/null | grep -q 'created_at,customer_id,id,status'`,
				"таблица store.orders со ссылкой на customers",
				"CREATE TABLE store.orders (id SERIAL PRIMARY KEY, customer_id INT REFERENCES store.customers(id), status TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT now());"),
		},
	},

	// ── Lab 2: вставка данных ──
	"ch-pgsql-lab-insert": {
		Image: sandboxImagePG,
		Setup: `set -e
pg-start
psql -qc "DROP DATABASE IF EXISTS training_shop" >/dev/null 2>&1 || true
psql -qc "CREATE DATABASE training_shop" >/dev/null
` + sqlSchemaDDL,
		Checks: map[int]string{
			1: check(q(`SELECT count(*) FROM store.customers WHERE email IN ('alex@example.test','max@example.test')`, "2"),
				"оба клиента добавлены",
				"INSERT INTO store.customers (full_name, email) VALUES ('Alex','alex@example.test'), ('Max','max@example.test');"),
			2: check(q(`SELECT count(*) FROM store.products WHERE (name,category,price,in_stock) IN (('PostgreSQL Guide','book',1900.00,25),('SQL Basics Course','course',4900.00,100),('Sticker Pack','merch',500.00,200))`, "3"),
				"все три товара добавлены с нужными ценами и остатками",
				"INSERT INTO store.products (name, category, price, in_stock) VALUES ('PostgreSQL Guide','book',1900.00,25), ('SQL Basics Course','course',4900.00,100), ('Sticker Pack','merch',500.00,200);"),
			3: check(q(`SELECT count(*) FROM store.orders WHERE (customer_id,status) IN ((1,'new'),(2,'paid'))`, "2"),
				"оба заказа добавлены",
				"INSERT INTO store.orders (customer_id, status) VALUES (1,'new'), (2,'paid');"),
			4: check(q(`SELECT count(*) FROM store.order_items WHERE (order_id,product_id,quantity,price) IN ((1,2,1,4900.00),(2,1,2,1900.00),(2,3,3,500.00))`, "3"),
				"все три позиции заказов добавлены",
				"INSERT INTO store.order_items (order_id, product_id, quantity, price) VALUES (1,2,1,4900.00), (2,1,2,1900.00), (2,3,3,500.00);"),
		},
	},

	// ── Lab 3: изменение данных и индексы ──
	"ch-pgsql-lab3": {
		Image: sandboxImagePG,
		Setup: `set -e
pg-start
psql -qc "DROP DATABASE IF EXISTS training_shop" >/dev/null 2>&1 || true
psql -qc "CREATE DATABASE training_shop" >/dev/null
` + sqlSchemaDDL + `
psql -qd training_shop -c "DROP TABLE IF EXISTS query_notes" >/dev/null
psql -qd training_shop -c "CREATE TABLE IF NOT EXISTS orders AS SELECT * FROM store.orders" >/dev/null
psql -qd training_shop -c "ALTER TABLE orders ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT now()" >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: check(`psql -tAd training_shop -c "SELECT string_agg(column_name, ',' ORDER BY column_name) FROM information_schema.columns WHERE table_name='query_notes'" 2>/dev/null | grep -q 'completed,created_at,id,note,topic'`,
				"таблица query_notes создана со всеми колонками",
				"CREATE TABLE query_notes (id SERIAL PRIMARY KEY, topic TEXT NOT NULL, note TEXT, completed BOOLEAN DEFAULT false, created_at TIMESTAMPTZ DEFAULT now()); — структура: \\d query_notes"),
			2: check(q(`SELECT count(*) FROM query_notes WHERE topic IN ('select','join','transaction')`, "3"),
				"три заметки добавлены",
				"INSERT INTO query_notes (topic, note) VALUES ('select','...'), ('join','...'), ('transaction','...');"),
			3: check(q(`SELECT count(*) FROM query_notes WHERE completed AND topic IN ('select','join')`, "2")+` && `+
				q(`SELECT count(*) FROM query_notes WHERE completed AND topic='transaction'`, "0"),
				"select и join отмечены выполненными",
				"UPDATE query_notes SET completed = true WHERE topic IN ('select','join');"),
			4: check(q(`SELECT count(*) FROM query_notes WHERE completed = false`, "0")+` && `+
				`[ "$(psql -tAd training_shop -c "SELECT count(*) FROM query_notes" 2>/dev/null | tr -d ' ')" -ge 1 ]`,
				"невыполненные заметки удалены, выполненные остались",
				"Сначала посмотри: SELECT * FROM query_notes WHERE completed = false; затем DELETE FROM query_notes WHERE completed = false;"),
			5: check(q(`SELECT count(*) FROM pg_indexes WHERE indexname='idx_orders_created_at'`, "1"),
				"индекс idx_orders_created_at создан",
				"CREATE INDEX idx_orders_created_at ON orders (created_at); — проверить: \\di или SELECT * FROM pg_indexes WHERE tablename='orders';"),
		},
	},
}
