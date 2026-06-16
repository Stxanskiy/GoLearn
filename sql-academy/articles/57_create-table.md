# Создание и удаление таблиц

## Создание таблицы

Перед созданием таблицы необходимо выбрать базу данных, в которую таблица будет записана. Это делается с помощью оператора USE:
```
MySQL 8.1
```

```
USE имя_базы_данных;

```

```

```

Для создания таблицы используется оператор CREATE TABLE. Его базовый синтаксис имеет следующий вид:

```
MySQL 8.1
```

```
CREATE TABLE [IF NOT EXISTS] имя_таблицы (
     столбец_1 тип_данных,
    [столбец_2 тип_данных,]
    ...
    [столбец_n тип_данных,]
);

```

```

```

Например, создадим таблицу пользователей.

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255),
    age INTEGER
);

```

```

```

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255),
    age INTEGER
);

```

```

```

INTEGER, VARCHAR(255) - типы данных: числовой и строковый соответственно. Более подробно о них можно будет узнать в следующих статьях.
INTEGER, VARCHAR(255) - типы данных: числовой и строковый соответственно. Более подробно о них можно будет узнать в следующих статьях.

## Дополнительные параметры определения столбцов

Вышеприведённое определение столбцов в таблице является упрощённым.
Помимо названия столбца и его типа в определение иногда необходимо добавлять следующие необязательные параметры:

- 
PRIMARY KEY
Указывает колонку или множество колонок как первичный ключ.

- 
AUTO_INCREMENT
Указывает, что значение данной колонки будет автоматически увеличиваться при добавлении новых записей в таблицу. Каждая таблица имеет максимум одну AUTO_INCREMENT колонку.
Стоит отметить, что данный параметр можно применять только к целочисленным типам и к типам с плавающей запятой.

- 
SERIAL или GENERATED ALWAYS AS IDENTITY
Указывает, что значение данной колонки будет автоматически увеличиваться при добавлении новых записей в таблицу. SERIAL — это сокращение для создания автоинкрементного поля.

- 
UNIQUE
Указывает, что значения в данной колонке для всех записей должны быть отличными друг от друга.

- 
NOT NULL
Указывает, что значения в данной колонке должны быть отличными от NULL.

- 
DEFAULT
Указывает значение по умолчанию.

Для нашей таблицы пользователей можно указать следующие параметры:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18
);

```

```

```

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18
);

```

```

```

Так, в данном примере:

- id - поле числового типа, являющееся первичным ключом с автоинкрементом;

- id - поле типа SERIAL (автоинкрементное целое число), являющееся первичным ключом;

- name - поле строкового типа с максимальной длиной в 255 символов, являющееся обязательным к заполнению;

- age - поле числового типа со значением по умолчанию равным 18.

## CURRENT_TIMESTAMP как значение по умолчанию

CURRENT_TIMESTAMP удобно использовать, когда нужно автоматически записывать время создания строки. Например, его можно использовать вместе с типом TIMESTAMP.

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

```

```

```

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

```

```

```

## Описание таблицы

Для того чтобы посмотреть описание созданной таблицы, можно воспользоваться оператором DESCRIBE.
```
MySQL 8.1
```

```
DESCRIBE Users;

```

```

```
FieldTypeNullKeyDefaultExtraidintNOPRINULLauto_incrementnamevarchar(255)NONULLageintNO18
Для того чтобы посмотреть описание созданной таблицы, можно воспользоваться SQL запросом к информационной схеме:
```
MySQL 8.1
```

```
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_schema = current_schema() AND table_name = 'users';

```

```

```
column_namedata_typeis_nullablecolumn_defaultidintegerNOnextval('users_id_seq'::regclass)namecharacter varyingNO<NULL>ageintegerNO18

## Дополнительные параметры определения таблицы

Помимо описания столбцов, при создании таблицы можно дополнительно указать следующие параметры:

- 
Первичный ключ.
Если вы не определили первичный ключ с помощью параметров столбца, то это можно сделать с помощью дополнительных параметров таблицы, добавив запись PRIMARY KEY (<столбец_1>, <столбец_n>) после перечисления столбцов:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    PRIMARY KEY (id)
);

```

```

```

- 
Первичный ключ.
Если вы не определили первичный ключ с помощью параметров столбца, то это можно сделать с помощью дополнительных параметров таблицы, добавив запись PRIMARY KEY (<столбец_1>, <столбец_n>) после перечисления столбцов:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    PRIMARY KEY (id)
);

```

```

```

- 
Внешние ключи.
Предположим, что мы хотим хранить данные о компании, в которой работают наши пользователи. Давайте создадим небольшую таблицу Companies, в которой мы будем хранить уникальный идентификатор и название компании:

```
MySQL 8.1
```

```
CREATE TABLE Companies (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);

```

```

```

Дальше нужно добавить в таблицу Users поле company – место работы нашего пользователя, которое будет ссылаться на запись в таблице Companies. Полный запрос для создания таблицы будет выглядеть так:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    company INTEGER,
    PRIMARY KEY (id)
);

```

```

```

Для того, чтобы при добавлении новых записей в таблицу Users гарантировать, что в колонке company находится идентификатор, существующий в таблице Companies,
используется внешний ключ. Он имеет следующий синтаксис:

```
MySQL 8.1
```

```
FOREIGN KEY (<столбец_1>, <столбец_n>)
REFERENCES <внешняя_таблица> (<столбец_во_внешней_таблице_1>, <столбец_во_внешней_таблице_n>)
[ON DELETE действие]
[ON UPDATE действие]

```

```

```

Полный запрос для создания таблицы с внешним ключом будет таким:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    company INTEGER,
    PRIMARY KEY (id),
    FOREIGN KEY (company) REFERENCES Companies (id)
);

```

```

```

При наличии внешних ключей можно определить поведение текущей записи, при изменении или удалении записи, на которую она ссылается.

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    company INTEGER,
    PRIMARY KEY (id),
    FOREIGN KEY (company) REFERENCES Companies (id)
    ON DELETE RESTRICT ON UPDATE CASCADE
);

```

```

```

ON DELETE RESTRICT означает, что если попробовать удалить компанию, у которой в таблице Users есть данные, база данных не даст этого сделать:

```
MySQL 8.1
```

```
Cannot delete or update a parent row: a foreign key constraint fails

```

```

```

Если бы было указано ON DELETE CASCADE, то при удалении компании были бы удалены все пользователи, ссылающиеся на эту компанию.
Есть ещё одна опция — ON DELETE SET NULL. При её использовании база данных запишет NULL в качестве значения поля company для всех пользователей, работавших в удалённой компании.
ON UPDATE CASCADE означает, что если компания изменит свой идентификатор, то все пользователи (Users) получат новый идентификатор в поле company.

- 
Внешние ключи.
Предположим, что мы хотим хранить данные о компании, в которой работают наши пользователи. Давайте создадим небольшую таблицу Companies, в которой мы будем хранить уникальный идентификатор и название компании:

```
MySQL 8.1
```

```
CREATE TABLE Companies (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);

```

```

```

Дальше нужно добавить в таблицу Users поле company – место работы нашего пользователя, которое будет ссылаться на запись в таблице Companies. Полный запрос для создания таблицы будет выглядеть так:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    company INTEGER,
    PRIMARY KEY (id)
);

```

```

```

Для того, чтобы при добавлении новых записей в таблицу Users гарантировать, что в колонке company находится идентификатор, существующий в таблице Companies,
используется внешний ключ. Он имеет следующий синтаксис:

```
MySQL 8.1
```

```
FOREIGN KEY (<столбец_1>, <столбец_n>)
REFERENCES <внешняя_таблица> (<столбец_во_внешней_таблице_1>, <столбец_во_внешней_таблице_n>)
[ON DELETE действие]
[ON UPDATE действие]

```

```

```

Полный запрос для создания таблицы с внешним ключом будет таким:

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    company INTEGER,
    PRIMARY KEY (id),
    FOREIGN KEY (company) REFERENCES Companies (id)
);

```

```

```

При наличии внешних ключей можно определить поведение текущей записи, при изменении или удалении записи, на которую она ссылается.

```
MySQL 8.1
```

```
CREATE TABLE Users (
    id INTEGER,
    name VARCHAR(255) NOT NULL,
    age INTEGER NOT NULL DEFAULT 18,
    company INTEGER,
    PRIMARY KEY (id),
    FOREIGN KEY (company) REFERENCES Companies (id)
    ON DELETE RESTRICT ON UPDATE CASCADE
);

```

```

```

ON DELETE RESTRICT означает, что если попробовать удалить компанию, у которой в таблице Users есть данные, база данных не даст этого сделать:

```
MySQL 8.1
```

```
ERROR:  update or delete on table "companies" violates foreign key constraint "users_company_fkey" on table "users"
DETAIL:  Key (id)=(1) is still referenced from table "users".

```

```

```

Если бы было указано ON DELETE CASCADE, то при удалении компании были бы удалены все пользователи, ссылающиеся на эту компанию.
Есть ещё одна опция — ON DELETE SET NULL. При её использовании база данных запишет NULL в качестве значения поля company для всех пользователей, работавших в удалённой компании.
ON UPDATE CASCADE означает, что если компания изменит свой идентификатор, то все пользователи (Users) получат новый идентификатор в поле company.

## Удаление таблицы

Удаление таблицы производится при помощи оператора DROP TABLE.

```
MySQL 8.1
```

```
DROP TABLE [IF EXISTS] имя_таблицы;

```

```

```

## Интерактивное задание

Теперь, когда вы изучили основы создания таблиц, попробуйте закрепить материал с помощью интерактивного задания:
Интерактивное упражнение недоступноДля использования интерактивного упражнения необходим экран шириной не менее 768 пикселей. Попробуйте открыть эту страницу на компьютере или планшете.