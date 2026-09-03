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

// shopSetup builds and populates the `shop` database the SELECT/report labs query
// (customers with cities incl. NULLs + VIPs, products across 4 categories with "sql"
// in some names, orders with mixed statuses/dates, order_items, support_tickets).
// pg-start boots a local PostgreSQL (postgres:15-alpine, baked offline) as superuser
// `student`, so `psql -U student -d shop` works exactly as the lesson text says.
const shopSetup = `pg-start
psql -qc "DROP DATABASE IF EXISTS shop" </dev/null 2>/dev/null || true
psql -qc "CREATE DATABASE shop" </dev/null
psql -qd shop <<'SQL'
CREATE TABLE customers (id SERIAL PRIMARY KEY, full_name TEXT NOT NULL, email TEXT UNIQUE NOT NULL, city TEXT, registered_at TIMESTAMPTZ NOT NULL DEFAULT now(), is_vip BOOLEAN NOT NULL DEFAULT false);
CREATE TABLE products (id SERIAL PRIMARY KEY, name TEXT NOT NULL, category TEXT NOT NULL, price NUMERIC(10,2) NOT NULL, in_stock INT NOT NULL);
CREATE TABLE orders (id SERIAL PRIMARY KEY, customer_id INT NOT NULL REFERENCES customers(id), status TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE order_items (id SERIAL PRIMARY KEY, order_id INT NOT NULL REFERENCES orders(id), product_id INT NOT NULL REFERENCES products(id), quantity INT NOT NULL, unit_price NUMERIC(10,2) NOT NULL);
CREATE TABLE support_tickets (id SERIAL PRIMARY KEY, customer_id INT REFERENCES customers(id), subject TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'open', created_at TIMESTAMPTZ NOT NULL DEFAULT now());
INSERT INTO customers (full_name,email,city,is_vip) VALUES
 ('Анна Иванова','anna@example.test','Москва',true),
 ('Борис Петров','boris@example.test','Санкт-Петербург',false),
 ('Вера Сидорова','vera@example.test','Москва',false),
 ('Глеб Кузнецов','gleb@example.test',NULL,false),
 ('Дарья Орлова','darya@example.test','Казань',true),
 ('Егор Смирнов','egor@example.test',NULL,false),
 ('Жанна Волкова','zhanna@example.test','Санкт-Петербург',false);
INSERT INTO products (name,category,price,in_stock) VALUES
 ('Механическая клавиатура','железо',5500,12),
 ('SSD 1TB','железо',7900,30),
 ('Книга: SQL с нуля','книги',1200,50),
 ('Книга: Чистый код','книги',1800,20),
 ('Курс: PostgreSQL для практиков','курсы',9900,999),
 ('Курс: Advanced SQL','курсы',12900,999),
 ('Футболка TOT','мерч',1500,100),
 ('Кружка TOT','мерч',700,80);
INSERT INTO orders (customer_id,status,created_at) VALUES
 (1,'paid','2026-05-12'),(2,'shipped','2026-05-15'),(3,'paid','2026-04-30'),
 (1,'pending','2026-05-20'),(5,'shipped','2026-05-11'),(2,'cancelled','2026-05-01'),(4,'paid','2026-05-18');
INSERT INTO order_items (order_id,product_id,quantity,unit_price) VALUES
 (1,1,1,5500),(1,3,2,1200),(2,5,1,9900),(3,7,3,1500),(4,2,1,7900),(5,6,1,12900),(5,4,1,1800),(7,8,4,700);
INSERT INTO support_tickets (customer_id,subject,status) VALUES
 (1,'Не пришёл заказ','open'),(2,'Вопрос по оплате','closed'),(NULL,'Общий вопрос','open');
SQL`

// joinLabSetup builds the `join_lab` database for the JOIN/UNION lab: employees
// (some without tasks), tasks (some without an assignee), 2 colours × 3 sizes for
// CROSS JOIN, and two signup lists that overlap for UNION vs UNION ALL.
const joinLabSetup = `pg-start
psql -qc "DROP DATABASE IF EXISTS join_lab" </dev/null 2>/dev/null || true
psql -qc "CREATE DATABASE join_lab" </dev/null
psql -qd join_lab <<'SQL'
CREATE TABLE employees (employee_id SERIAL PRIMARY KEY, full_name TEXT NOT NULL, email TEXT NOT NULL);
CREATE TABLE tasks (task_id SERIAL PRIMARY KEY, title TEXT NOT NULL, assignee_id INT REFERENCES employees(employee_id));
CREATE TABLE shirt_colors (color TEXT PRIMARY KEY);
CREATE TABLE shirt_sizes (size_label TEXT PRIMARY KEY);
CREATE TABLE web_signups (email TEXT);
CREATE TABLE event_signups (email TEXT);
INSERT INTO employees (full_name,email) VALUES
 ('Иван Иванов','ivan@example.test'),('Пётр Петров','petr@example.test'),
 ('Мария Кузнецова','maria@example.test'),('Ольга Новак','olga@example.test');
INSERT INTO tasks (title,assignee_id) VALUES
 ('Настроить CI',1),('Написать доку',2),('Ревью PR',1),
 ('Задача без исполнителя',NULL),('Ещё без назначения',NULL);
INSERT INTO shirt_colors (color) VALUES ('чёрный'),('белый');
INSERT INTO shirt_sizes (size_label) VALUES ('S'),('M'),('L');
INSERT INTO web_signups (email) VALUES ('anna@example.test'),('pavel@example.test'),('oleg@example.test');
INSERT INTO event_signups (email) VALUES ('pavel@example.test'),('nina@example.test'),('oleg@example.test');
SQL`

var sqlExpressLabs = map[string]labSpec{
	// Read-practice labs ("выполни запрос → посмотри → Готово"): the Setup builds the
	// database the queries need; tasks stay manual (nothing persistent to auto-check).
	"ch-pgsql-lab1":           {Image: sandboxImagePG, Setup: shopSetup},
	"ch-pgsql-lab2":           {Image: sandboxImagePG, Setup: shopSetup},
	"ch-pgsql-lab-join-types": {Image: sandboxImagePG, Setup: joinLabSetup},

	// ── Lab 1: база, схема, таблицы ──
	"ch-pgsql-lab-schema": {
		Image: sandboxImagePG,
		Setup: `set -e
pg-start
psql -qc "DROP DATABASE IF EXISTS training_shop" </dev/null >/dev/null 2>&1 || true`,
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
psql -qc "DROP DATABASE IF EXISTS training_shop" </dev/null >/dev/null 2>&1 || true
psql -qc "CREATE DATABASE training_shop" </dev/null >/dev/null
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
psql -qc "DROP DATABASE IF EXISTS training_shop" </dev/null >/dev/null 2>&1 || true
psql -qc "CREATE DATABASE training_shop" </dev/null >/dev/null
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
